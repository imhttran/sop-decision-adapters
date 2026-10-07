package benchmark

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/imhttran/sop-decision-adapters/decision"
	"github.com/imhttran/sop-decision-adapters/internal/shadow"
)

// Status reports the measurability of a benchmark cell. A cell that did not run
// is never reported as a measured zero.
type Status string

const (
	// StatusMeasured means the cell ran against the corpus and its metrics are
	// measured values (possibly with small sample counts).
	StatusMeasured Status = "measured"
	// StatusUnavailable means the cell's provider/runtime was not reachable, so
	// no live measurement was performed. Metrics are reported as not measurable.
	StatusUnavailable Status = "unavailable"
	// StatusNotExercised means the cell's entry condition was not met (for
	// example an unmet prerequisite), so it was not run at all.
	StatusNotExercised Status = "not_exercised"
)

// Outcome classifies one cell x case observation. It is exhaustive and never
// collapses malformed/timeout/failure/indeterminate/invalid results into a
// single bucket.
type Outcome string

const (
	OutcomeAgreement     Outcome = "agreement"
	OutcomeDisagreement  Outcome = "disagreement"
	OutcomeUnlabeled     Outcome = "unlabeled"
	OutcomeInvalidChoice Outcome = "invalid_choice"
	OutcomeIndeterminate Outcome = "indeterminate"
	OutcomeTimeout       Outcome = "timeout"
	OutcomeMalformed     Outcome = "malformed"
	OutcomeFailure       Outcome = "failure"
	OutcomeExcluded      Outcome = "excluded"
)

// Stat is an explicit numerator/denominator pair. A zero denominator means the
// metric was not measurable; it is never collapsed into a measured zero.
type Stat struct {
	Numerator   int `json:"numerator"`
	Denominator int `json:"denominator"`
}

// Rate returns Numerator/Denominator, or 0 when the denominator is zero. Callers
// must consult Measured to distinguish "0 of 0" from a measured 0.0.
func (s Stat) Rate() float64 {
	if s.Denominator == 0 {
		return 0
	}
	return float64(s.Numerator) / float64(s.Denominator)
}

// Measured reports whether the statistic had a non-empty denominator.
func (s Stat) Measured() bool { return s.Denominator > 0 }

// Record is one deterministic per-cell, per-case observation. LatencyMS is the
// only non-deterministic field (wall-clock).
type Record struct {
	CaseID      string       `json:"case_id"`
	Cell        string       `json:"cell"`
	Outcome     Outcome      `json:"outcome"`
	Choice      string       `json:"choice,omitempty"`
	Expected    string       `json:"expected,omitempty"`
	GroundTruth bool         `json:"ground_truth"`
	Confidence  *float64     `json:"confidence,omitempty"`
	LatencyMS   float64      `json:"latency_ms"`
	Label       shadow.Label `json:"label,omitempty"`
}

// CalibrationBin is one reliability bucket over ground-truth-labeled cases that
// carried a confidence value.
type CalibrationBin struct {
	Bucket      string `json:"bucket"`
	Predictions int    `json:"predictions"`
	Correct     int    `json:"correct"`
}

// CellResult is the deterministic result for one benchmark cell. When Status is
// not measured, the metric stats are zero-valued and must be read as "not
// measurable", not as zero.
type CellResult struct {
	Cell       string `json:"cell"`
	Matrix     string `json:"matrix"`
	Status     Status `json:"status"`
	Reason     string `json:"reason,omitempty"`
	CorpusHash string `json:"corpus_hash"`
	Cases      int    `json:"cases"`

	// Sample is the number of comparable cases evaluated. Excluded is the number
	// of cases gated out as non-comparable. Sample+Excluded == Cases for a
	// measured cell.
	Sample   int `json:"sample"`
	Excluded int `json:"excluded"`

	// ChoiceValid: numerator valid choices, denominator choice-bearing
	// observations (valid + invalid). Errors and indeterminate outcomes are not
	// choice-bearing and are excluded from this denominator.
	ChoiceValid Stat `json:"choice_valid"`
	// Agreement: numerator agreements, denominator ground-truth-labeled
	// comparisons with a definite choice (agreement + disagreement).
	Agreement Stat `json:"agreement"`
	// ConfidenceAvailable: numerator observations exposing a confidence value,
	// denominator Sample.
	ConfidenceAvailable Stat `json:"confidence_available"`
	// Timeout / Failure / Indeterminate: numerator such outcomes, denominator
	// Sample.
	Timeout       Stat `json:"timeout"`
	Failure       Stat `json:"failure"`
	Indeterminate Stat `json:"indeterminate"`

	LatencyP50MS float64 `json:"latency_p50_ms"`
	LatencyP95MS float64 `json:"latency_p95_ms"`

	ConfidenceBuckets map[string]int   `json:"confidence_buckets,omitempty"`
	Calibration       []CalibrationBin `json:"calibration,omitempty"`

	Records []Record `json:"records,omitempty"`

	latencies []time.Duration
}

// Source is a provider-neutral benchmark subject. It reuses the CLEF-007
// observation vocabulary (shadow.Observation / shadow.Label) and adds a
// non-inference availability check.
type Source interface {
	// Name is the stable subject identifier.
	Name() string
	// Available reports whether the subject can currently serve decisions; it
	// must not perform an inference request.
	Available(ctx context.Context) bool
	// Comparable reports whether the case can be evaluated by this subject.
	// Non-comparable cases are excluded from the metrics, not counted as
	// failures.
	Comparable(c Case) bool
	// Observe obtains the subject's outcome for the case. Any error is contained
	// by the harness and can never affect a primary decision.
	Observe(ctx context.Context, c Case) (shadow.Observation, error)
}

// Cell is one benchmark cell: a matrix slot together with the source to run.
// A nil Source marks an untested cell (for example an unmet prerequisite) that
// is recorded as not exercised with its Reason.
type Cell struct {
	ID     string
	Matrix string
	Source Source
	Reason string
}

// Report is the deterministic, machine-readable benchmark result.
type Report struct {
	CorpusVersion string       `json:"corpus_version"`
	CorpusHash    string       `json:"corpus_hash"`
	Cases         int          `json:"cases"`
	LabeledCases  int          `json:"labeled_cases"`
	Cells         []CellResult `json:"cells"`
}

// JSON returns the report as indented JSON. Map fields are encoded with sorted
// keys, so the output is deterministic apart from measured latency values.
func (r Report) JSON() ([]byte, error) { return json.MarshalIndent(r, "", "  ") }

// Run executes the corpus against every cell. Every cell is evaluated against
// the identical corpus and case order.
func Run(ctx context.Context, corpus Corpus, cells []Cell) (Report, error) {
	hash, err := corpus.Hash()
	if err != nil {
		return Report{}, err
	}
	rep := Report{
		CorpusVersion: corpus.Version,
		CorpusHash:    hash,
		Cases:         len(corpus.Cases),
		LabeledCases:  corpus.LabeledCases(),
	}
	for _, cell := range cells {
		rep.Cells = append(rep.Cells, runCell(ctx, corpus, hash, cell))
	}
	return rep, nil
}

func runCell(ctx context.Context, corpus Corpus, hash string, cell Cell) CellResult {
	res := CellResult{
		Cell:       cell.ID,
		Matrix:     cell.Matrix,
		CorpusHash: hash,
		Cases:      len(corpus.Cases),
	}

	if cell.Source == nil {
		res.Status = StatusNotExercised
		res.Reason = nonEmpty(cell.Reason, "cell not exercised")
		return res
	}
	if !cell.Source.Available(ctx) {
		res.Status = StatusUnavailable
		res.Reason = nonEmpty(cell.Reason, "provider/runtime not available; live measurement not exercised")
		return res
	}

	res.Status = StatusMeasured
	res.ConfidenceBuckets = newConfidenceBuckets()
	res.Calibration = newCalibrationBins()

	var (
		choiceBearing, validChoices       int
		labeled, agreements               int
		withConfidence                    int
		timeouts, failures, indeterminate int
	)

	for _, c := range corpus.Cases {
		if !cell.Source.Comparable(c) {
			res.Excluded++
			res.Records = append(res.Records, Record{
				CaseID: c.ID, Cell: cell.ID, Outcome: OutcomeExcluded,
				Expected: c.Expected, GroundTruth: c.GroundTruth,
				Label: shadow.LabelSkippedNotComparable,
			})
			continue
		}

		res.Sample++
		obs, err, elapsed := observeSafely(ctx, cell.Source, c)
		res.latencies = append(res.latencies, elapsed)

		rec := Record{
			CaseID: c.ID, Cell: cell.ID,
			Expected: c.Expected, GroundTruth: c.GroundTruth,
			LatencyMS: millis(elapsed),
		}

		switch {
		case err != nil:
			switch {
			case isTimeout(err):
				rec.Outcome, rec.Label = OutcomeTimeout, shadow.LabelTimeout
				timeouts++
			case isMalformed(err):
				rec.Outcome, rec.Label = OutcomeMalformed, shadow.LabelMalformed
				failures++
			default:
				rec.Outcome, rec.Label = OutcomeFailure, shadow.LabelError
				failures++
			}
		case obs.Confidence != nil && (*obs.Confidence < 0 || *obs.Confidence > 1):
			rec.Outcome, rec.Label = OutcomeMalformed, shadow.LabelMalformed
			failures++
		case obs.Choice == "":
			rec.Outcome, rec.Label = OutcomeIndeterminate, shadow.LabelIndeterminate
			indeterminate++
		default:
			rec.Choice = obs.Choice
			if obs.Confidence != nil {
				rec.Confidence = obs.Confidence
				withConfidence++
				res.ConfidenceBuckets[confidenceBucket(*obs.Confidence)]++
			}
			choiceBearing++
			if !c.Question.Allows(obs.Choice) {
				rec.Outcome, rec.Label = OutcomeInvalidChoice, shadow.LabelInvalidChoice
			} else {
				validChoices++
				if c.GroundTruth {
					labeled++
					correct := obs.Choice == c.Expected
					if correct {
						agreements++
						rec.Outcome, rec.Label = OutcomeAgreement, shadow.LabelAgreement
					} else {
						rec.Outcome, rec.Label = OutcomeDisagreement, shadow.LabelDisagreement
					}
					if obs.Confidence != nil {
						addCalibration(res.Calibration, *obs.Confidence, correct)
					}
				} else {
					rec.Outcome = OutcomeUnlabeled
				}
			}
		}

		res.Records = append(res.Records, rec)
	}

	res.ChoiceValid = Stat{Numerator: validChoices, Denominator: choiceBearing}
	res.Agreement = Stat{Numerator: agreements, Denominator: labeled}
	res.ConfidenceAvailable = Stat{Numerator: withConfidence, Denominator: res.Sample}
	res.Timeout = Stat{Numerator: timeouts, Denominator: res.Sample}
	res.Failure = Stat{Numerator: failures, Denominator: res.Sample}
	res.Indeterminate = Stat{Numerator: indeterminate, Denominator: res.Sample}
	res.LatencyP50MS = percentile(res.latencies, 0.50)
	res.LatencyP95MS = percentile(res.latencies, 0.95)

	return res
}

// observeSafely invokes a source with panic containment so a misbehaving
// benchmark subject can never crash the harness or reach a primary decision.
func observeSafely(ctx context.Context, src Source, c Case) (obs shadow.Observation, err error, elapsed time.Duration) {
	start := time.Now()
	defer func() {
		if rec := recover(); rec != nil {
			obs = shadow.Observation{}
			err = fmt.Errorf("benchmark source %s panicked: %v", src.Name(), rec)
		}
		elapsed = time.Since(start)
	}()
	obs, err = src.Observe(ctx, c)
	return obs, err, 0
}

// isTimeout mirrors the CLEF-007 shadow containment classification: a context
// deadline or cancellation, or any error whose Timeout() reports true, is a
// timeout.
func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	var t interface{ Timeout() bool }
	return errors.As(err, &t) && t.Timeout()
}

// isMalformed reports whether err unwraps to a malformed-response sentinel.
func isMalformed(err error) bool {
	return err != nil && errors.Is(err, decision.ErrMalformedResponse)
}

func nonEmpty(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}

func millis(d time.Duration) float64 { return float64(d.Microseconds()) / 1000.0 }

// percentile returns the nearest-rank percentile of xs (0 < p <= 1).
func percentile(xs []time.Duration, p float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sorted := append([]time.Duration(nil), xs...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	n := len(sorted)
	rk := int(p*float64(n) + 0.999999)
	if rk < 1 {
		rk = 1
	}
	if rk > n {
		rk = n
	}
	return millis(sorted[rk-1])
}

// confidenceBuckets are the fixed 0.2-step reliability buckets over [0,1].
var confidenceBuckets = []string{"0.0-0.2", "0.2-0.4", "0.4-0.6", "0.6-0.8", "0.8-1.0"}

func newConfidenceBuckets() map[string]int {
	m := make(map[string]int, len(confidenceBuckets))
	for _, b := range confidenceBuckets {
		m[b] = 0
	}
	return m
}

func confidenceBucket(v float64) string {
	switch {
	case v < 0.2:
		return "0.0-0.2"
	case v < 0.4:
		return "0.2-0.4"
	case v < 0.6:
		return "0.4-0.6"
	case v < 0.8:
		return "0.6-0.8"
	default:
		return "0.8-1.0"
	}
}

func newCalibrationBins() []CalibrationBin {
	bins := make([]CalibrationBin, 0, len(confidenceBuckets))
	for _, b := range confidenceBuckets {
		bins = append(bins, CalibrationBin{Bucket: b})
	}
	return bins
}

// addCalibration records one prediction in the bin for confidence v. The bin for
// v is located by bucket key so ordering is stable.
func addCalibration(bins []CalibrationBin, v float64, correct bool) {
	key := confidenceBucket(v)
	for i := range bins {
		if bins[i].Bucket == key {
			bins[i].Predictions++
			if correct {
				bins[i].Correct++
			}
			return
		}
	}
}
