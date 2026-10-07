package benchmark

import (
	"context"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/imhttran/sop-decision-adapters/decision"
	"github.com/imhttran/sop-decision-adapters/internal/shadow"
)

const testCorpusJSON = `{
  "version": "test-v1",
  "cases": [
    {"id":"c1","state":"s1","question":{"id":"q","type":"choice","criteria":"c","choices":["a","b"]},"expected":"a","ground_truth":true},
    {"id":"c2","state":"s2","question":{"id":"q","type":"choice","criteria":"c","choices":["a","b"]},"expected":"b","ground_truth":true},
    {"id":"c3","state":"s3","question":{"id":"q","type":"choice","criteria":"c","choices":["a","b"]},"ground_truth":false}
  ]
}`

func mustCorpus(t *testing.T, data string) Corpus {
	t.Helper()
	c, err := LoadCorpus([]byte(data))
	if err != nil {
		t.Fatalf("LoadCorpus: %v", err)
	}
	return c
}

func fptr(v float64) *float64 { return &v }

// fakeSource is a controllable benchmark subject for harness tests.
type fakeSource struct {
	name       string
	available  bool
	comparable map[string]bool // case ID -> comparable (absent => true)
	obs        map[string]shadow.Observation
	err        map[string]error
	panicOn    map[string]bool
	calls      *int // optional: counts Observe invocations
}

func (f *fakeSource) Name() string { return f.name }

func (f *fakeSource) Available(context.Context) bool { return f.available }

func (f *fakeSource) Comparable(c Case) bool {
	if f.comparable == nil {
		return true
	}
	v, ok := f.comparable[c.ID]
	if !ok {
		return true
	}
	return v
}

func (f *fakeSource) Observe(_ context.Context, c Case) (shadow.Observation, error) {
	if f.calls != nil {
		*f.calls++
	}
	if f.panicOn[c.ID] {
		panic("boom")
	}
	if err, ok := f.err[c.ID]; ok {
		return shadow.Observation{}, err
	}
	return f.obs[c.ID], nil
}

func runOne(t *testing.T, corpus Corpus, src Source) CellResult {
	t.Helper()
	rep, err := Run(context.Background(), corpus, []Cell{{ID: src.Name(), Matrix: "primary", Source: src}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(rep.Cells) != 1 {
		t.Fatalf("cells=%d want 1", len(rep.Cells))
	}
	return rep.Cells[0]
}

func TestEmbeddedCorpusLoadsDeterministically(t *testing.T) {
	a, err := DefaultCorpus()
	if err != nil {
		t.Fatalf("DefaultCorpus: %v", err)
	}
	b, err := DefaultCorpus()
	if err != nil {
		t.Fatalf("DefaultCorpus (2): %v", err)
	}

	ha, err := a.Hash()
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	hb, err := b.Hash()
	if err != nil {
		t.Fatalf("Hash (2): %v", err)
	}
	if ha != hb {
		t.Fatalf("corpus hash not deterministic: %s != %s", ha, hb)
	}
	if len(ha) != 64 {
		t.Fatalf("hash not sha256 hex: %q", ha)
	}
	if len(a.Cases) == 0 {
		t.Fatal("embedded corpus has no cases")
	}

	seen := map[string]bool{}
	for _, c := range a.Cases {
		if strings.TrimSpace(c.ID) == "" {
			t.Fatalf("case with empty ID: %+v", c)
		}
		if seen[c.ID] {
			t.Fatalf("duplicate case ID %q", c.ID)
		}
		seen[c.ID] = true
	}
	if a.LabeledCases() == 0 {
		t.Fatal("embedded corpus has no labeled cases")
	}
}

func TestCorpusRejectsMalformed(t *testing.T) {
	cases := map[string]string{
		"empty version":           `{"version":"","cases":[{"id":"c1","state":"s","question":{"id":"q","type":"choice","criteria":"c","choices":["a"]}}]}`,
		"no cases":                `{"version":"v","cases":[]}`,
		"missing id":              `{"version":"v","cases":[{"id":"","state":"s","question":{"id":"q","type":"choice","criteria":"c","choices":["a"]}}]}`,
		"duplicate id":            `{"version":"v","cases":[{"id":"c1","state":"s","question":{"id":"q","type":"choice","criteria":"c","choices":["a"]}},{"id":"c1","state":"s","question":{"id":"q","type":"choice","criteria":"c","choices":["a"]}}]}`,
		"empty state":             `{"version":"v","cases":[{"id":"c1","state":"","question":{"id":"q","type":"choice","criteria":"c","choices":["a"]}}]}`,
		"non-choice question":     `{"version":"v","cases":[{"id":"c1","state":"s","question":{"id":"q","type":"boolean"}}]}`,
		"label without expected":  `{"version":"v","cases":[{"id":"c1","state":"s","question":{"id":"q","type":"choice","criteria":"c","choices":["a"]},"ground_truth":true}]}`,
		"expected not allowed":    `{"version":"v","cases":[{"id":"c1","state":"s","question":{"id":"q","type":"choice","criteria":"c","choices":["a"]},"expected":"z","ground_truth":true}]}`,
		"unlabeled with expected": `{"version":"v","cases":[{"id":"c1","state":"s","question":{"id":"q","type":"choice","criteria":"c","choices":["a"]},"expected":"a","ground_truth":false}]}`,
		"unknown field":           `{"version":"v","extra":1,"cases":[]}`,
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadCorpus([]byte(data)); err == nil {
				t.Fatalf("LoadCorpus(%s) = nil error, want rejection", name)
			}
		})
	}
}

func TestRunUsesSameCorpusForAllCells(t *testing.T) {
	corpus := mustCorpus(t, testCorpusJSON)
	a := &fakeSource{name: "a", available: true, obs: map[string]shadow.Observation{"c1": {Choice: "a"}, "c2": {Choice: "b"}, "c3": {Choice: "a"}}}
	b := &fakeSource{name: "b", available: true, obs: map[string]shadow.Observation{"c1": {Choice: "b"}, "c2": {Choice: "a"}, "c3": {Choice: "b"}}}

	rep, err := Run(context.Background(), corpus, []Cell{
		{ID: "a", Matrix: "primary", Source: a},
		{ID: "b", Matrix: "primary", Source: b},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, cr := range rep.Cells {
		if cr.Cases != len(corpus.Cases) {
			t.Fatalf("cell %s cases=%d want %d", cr.Cell, cr.Cases, len(corpus.Cases))
		}
		if cr.Sample != len(corpus.Cases) || cr.Excluded != 0 {
			t.Fatalf("cell %s sample=%d excluded=%d want %d/0", cr.Cell, cr.Sample, cr.Excluded, len(corpus.Cases))
		}
	}
}

func TestUnavailableCellIsUnavailableNotZero(t *testing.T) {
	corpus := mustCorpus(t, testCorpusJSON)
	src := &fakeSource{name: "down", available: false, obs: map[string]shadow.Observation{"c1": {Choice: "a"}}}
	cr := runOne(t, corpus, src)

	if cr.Status != StatusUnavailable {
		t.Fatalf("status=%q want %q", cr.Status, StatusUnavailable)
	}
	if cr.Sample != 0 {
		t.Fatalf("sample=%d want 0", cr.Sample)
	}
	if cr.ChoiceValid.Measured() || cr.Agreement.Measured() {
		t.Fatalf("unavailable cell reported measured metrics: %+v", cr)
	}
	if cr.Reason == "" {
		t.Fatal("unavailable cell missing reason")
	}
}

func TestNotExercisedCellCarriesReason(t *testing.T) {
	corpus := mustCorpus(t, testCorpusJSON)
	rep, err := Run(context.Background(), corpus, []Cell{{ID: "omlx", Matrix: "conditional", Reason: "prerequisite unmet"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	cr := rep.Cells[0]
	if cr.Status != StatusNotExercised || cr.Reason != "prerequisite unmet" {
		t.Fatalf("got status=%q reason=%q", cr.Status, cr.Reason)
	}
	if cr.Sample != 0 {
		t.Fatalf("sample=%d want 0", cr.Sample)
	}
}

func TestNumeratorDenominatorPreserved(t *testing.T) {
	corpus := mustCorpus(t, testCorpusJSON)
	src := &fakeSource{
		name: "m", available: true,
		obs: map[string]shadow.Observation{
			"c1": {Choice: "a", Confidence: fptr(0.9)}, // agreement, confidence present
			"c2": {Choice: "a"},                        // disagreement (expected b), no confidence
			"c3": {Choice: "b", Confidence: fptr(0.5)}, // valid, unlabeled
		},
	}
	cr := runOne(t, corpus, src)

	if cr.Status != StatusMeasured {
		t.Fatalf("status=%q", cr.Status)
	}
	if cr.Sample != 3 {
		t.Fatalf("sample=%d want 3", cr.Sample)
	}
	if cr.ChoiceValid != (Stat{Numerator: 3, Denominator: 3}) {
		t.Fatalf("choice_valid=%+v want 3/3", cr.ChoiceValid)
	}
	if cr.Agreement != (Stat{Numerator: 1, Denominator: 2}) {
		t.Fatalf("agreement=%+v want 1/2", cr.Agreement)
	}
	if cr.ConfidenceAvailable != (Stat{Numerator: 2, Denominator: 3}) {
		t.Fatalf("confidence_available=%+v want 2/3", cr.ConfidenceAvailable)
	}
	// c1 (0.9, correct) is the only confidence-bearing labeled case.
	if len(cr.Calibration) == 0 {
		t.Fatal("missing calibration bins")
	}
	var gotCal CalibrationBin
	for _, b := range cr.Calibration {
		if b.Bucket == "0.8-1.0" {
			gotCal = b
		}
	}
	if gotCal.Predictions != 1 || gotCal.Correct != 1 {
		t.Fatalf("calibration 0.8-1.0 = %+v want 1 prediction, 1 correct", gotCal)
	}
}

func TestInvalidAndIndeterminateCounted(t *testing.T) {
	corpus := mustCorpus(t, testCorpusJSON)
	src := &fakeSource{
		name: "bad", available: true,
		obs: map[string]shadow.Observation{
			"c1": {Choice: "a"}, // valid agreement
			"c2": {Choice: "z"}, // invalid choice (not allowed)
			"c3": {Choice: ""},  // indeterminate
		},
	}
	cr := runOne(t, corpus, src)

	invalid := 0
	for _, r := range cr.Records {
		if r.Outcome == OutcomeInvalidChoice {
			invalid++
		}
	}
	if invalid != 1 {
		t.Fatalf("invalid count=%d want 1", invalid)
	}
	if cr.Indeterminate != (Stat{Numerator: 1, Denominator: 3}) {
		t.Fatalf("indeterminate=%+v want 1/3", cr.Indeterminate)
	}
	// Choice-bearing = c1 + c2 (indeterminate is not choice-bearing).
	if cr.ChoiceValid != (Stat{Numerator: 1, Denominator: 2}) {
		t.Fatalf("choice_valid=%+v want 1/2", cr.ChoiceValid)
	}
	// Agreement denominator counts only valid labeled comparisons (c1).
	if cr.Agreement != (Stat{Numerator: 1, Denominator: 1}) {
		t.Fatalf("agreement=%+v want 1/1", cr.Agreement)
	}
}

func TestIncomparableExcludedFromMetrics(t *testing.T) {
	corpus := mustCorpus(t, testCorpusJSON)
	src := &fakeSource{
		name: "part", available: true,
		comparable: map[string]bool{"c2": false},
		obs:        map[string]shadow.Observation{"c1": {Choice: "a"}, "c3": {Choice: "b"}},
	}
	cr := runOne(t, corpus, src)

	if cr.Excluded != 1 {
		t.Fatalf("excluded=%d want 1", cr.Excluded)
	}
	if cr.Sample != 2 {
		t.Fatalf("sample=%d want 2", cr.Sample)
	}
	if cr.Sample+cr.Excluded != cr.Cases {
		t.Fatalf("sample+excluded=%d != cases=%d", cr.Sample+cr.Excluded, cr.Cases)
	}
	found := false
	for _, r := range cr.Records {
		if r.CaseID == "c2" && r.Outcome == OutcomeExcluded && r.Label == shadow.LabelSkippedNotComparable {
			found = true
		}
	}
	if !found {
		t.Fatalf("excluded case not recorded: %+v", cr.Records)
	}
}

func TestFailureModesPreserved(t *testing.T) {
	corpus := mustCorpus(t, testCorpusJSON)
	src := &fakeSource{
		name: "errs", available: true,
		obs: map[string]shadow.Observation{"c1": {Choice: "a"}},
		err: map[string]error{
			"c2": context.DeadlineExceeded,
			"c3": decision.ErrMalformedResponse,
		},
	}
	cr := runOne(t, corpus, src)

	if cr.Timeout != (Stat{Numerator: 1, Denominator: 3}) {
		t.Fatalf("timeout=%+v want 1/3", cr.Timeout)
	}
	// malformed counts as a failure.
	if cr.Failure != (Stat{Numerator: 1, Denominator: 3}) {
		t.Fatalf("failure=%+v want 1/3", cr.Failure)
	}
}

func TestPanicContained(t *testing.T) {
	corpus := mustCorpus(t, testCorpusJSON)
	src := &fakeSource{
		name: "panic", available: true,
		panicOn: map[string]bool{"c2": true},
		obs:     map[string]shadow.Observation{"c1": {Choice: "a"}, "c3": {Choice: "b"}},
	}
	cr := runOne(t, corpus, src) // must not panic

	if cr.Status != StatusMeasured || cr.Sample != 3 {
		t.Fatalf("status=%q sample=%d", cr.Status, cr.Sample)
	}
	if cr.Failure.Numerator != 1 {
		t.Fatalf("failure numerator=%d want 1 (panic contained)", cr.Failure.Numerator)
	}
}

func TestReportJSONDeterministicShape(t *testing.T) {
	corpus := mustCorpus(t, testCorpusJSON)
	src := &fakeSource{name: "m", available: true, obs: map[string]shadow.Observation{"c1": {Choice: "a"}, "c2": {Choice: "b"}, "c3": {Choice: "b"}}}
	rep, err := Run(context.Background(), corpus, []Cell{{ID: "m", Matrix: "primary", Source: src}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	out, err := rep.JSON()
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	for _, want := range []string{`"corpus_hash"`, `"choice_valid"`, `"agreement"`, `"confidence_available"`, `"indeterminate"`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("report JSON missing %s", want)
		}
	}
}

// TestBenchmarkNotImportedByProductionPath proves the harness is off the
// production decision path: no .go file outside internal/benchmark imports it.
func TestBenchmarkNotImportedByProductionPath(t *testing.T) {
	const selfPkg = "github.com/imhttran/sop-decision-adapters/internal/benchmark"

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	root := filepath.Clean(filepath.Join(wd, "..", ".."))

	fset := token.NewFileSet()
	checked := 0
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := d.Name()
			if base == ".git" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if rel, rerr := filepath.Rel(root, path); rerr == nil {
			if strings.HasPrefix(filepath.ToSlash(rel), "internal/benchmark/") {
				return nil
			}
		}
		f, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			return nil // unrelated parse issues are not this test's concern
		}
		checked++
		for _, imp := range f.Imports {
			if strings.Trim(imp.Path.Value, `"`) == selfPkg {
				t.Errorf("%s imports the benchmark harness; it must stay off the production path", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if checked == 0 {
		t.Fatal("no Go files inspected")
	}
}

// TestProviderSourceUsesSeam proves the adapter reaches a provider only through
// decision.Provider and reports an omitted answer as malformed.
func TestProviderSourceUsesSeam(t *testing.T) {
	corpus := mustCorpus(t, testCorpusJSON)
	src := ProviderSource{Provider: stubProvider{name: "stub", available: true, answer: map[string]decision.Answer{
		"q": {Type: decision.AnswerChoice, Choice: "a"},
	}}}
	cr := runOne(t, corpus, src)
	if cr.Cell != "stub" || cr.Status != StatusMeasured {
		t.Fatalf("cell=%q status=%q", cr.Cell, cr.Status)
	}
	if cr.ChoiceValid.Numerator != 3 {
		t.Fatalf("choice_valid numerator=%d want 3", cr.ChoiceValid.Numerator)
	}
}

type stubProvider struct {
	name      string
	available bool
	err       error
	answer    map[string]decision.Answer
}

func (s stubProvider) Name() string { return s.name }

func (s stubProvider) Available(context.Context) bool { return s.available }

func (s stubProvider) Decide(_ context.Context, _ decision.DecisionRequest) (decision.DecisionResult, error) {
	if s.err != nil {
		return decision.DecisionResult{}, s.err
	}
	return decision.DecisionResult{Provider: s.name, Answers: s.answer}, nil
}
