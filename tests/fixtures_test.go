package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/imhttran/sop-decision-adapters/decision"
)

// wantFixtures maps a single-question fixture to its expected choices. All other
// .json fixtures are still validated as DecisionRequests.
var wantFixtures = map[string][]string{
	"risk.json":            {"LOW", "MEDIUM", "HIGH"},
	"execution_model.json": {"SMALL", "MEDIUM", "LARGE"},
}

func TestFixturesAreValidDecisionRequests(t *testing.T) {
	entries, err := os.ReadDir("fixtures")
	if err != nil {
		t.Fatalf("read fixtures dir: %v", err)
	}

	seen := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		name := entry.Name()
		seen[name] = true

		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("fixtures", name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			var req decision.DecisionRequest
			if err := json.Unmarshal(data, &req); err != nil {
				t.Fatalf("unmarshal fixture: %v", err)
			}
			if err := req.Validate(); err != nil {
				t.Fatalf("fixture is not a valid DecisionRequest: %v", err)
			}
			if want, ok := wantFixtures[name]; ok {
				if len(req.Questions) != 1 {
					t.Fatalf("fixture has %d questions, want 1", len(req.Questions))
				}
				if !slices.Equal(req.Questions[0].Choices, want) {
					t.Errorf("choices = %v, want %v", req.Questions[0].Choices, want)
				}
			}
		})
	}

	for name := range wantFixtures {
		if !seen[name] {
			t.Errorf("expected fixture %q was not found", name)
		}
	}
}

// TestRiskEvaluationFixtureIsMultiQuestion pins the CLI example fixture to the
// PRD's SOP example: one state evaluated against risk, approval_required, and
// execution_model.
func TestRiskEvaluationFixtureIsMultiQuestion(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("fixtures", "risk-evaluation.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var req decision.DecisionRequest
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("fixture is not a valid DecisionRequest: %v", err)
	}
	for _, id := range []string{"risk", "approval_required", "execution_model"} {
		if _, ok := req.Question(id); !ok {
			t.Errorf("missing question %q", id)
		}
	}
}

// TestNoProviderSpecificTermsInFixtures guards against leaking transport
// terminology into fixture files.
func TestNoProviderSpecificTermsInFixtures(t *testing.T) {
	entries, err := os.ReadDir("fixtures")
	if err != nil {
		t.Fatalf("read fixtures dir: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join("fixtures", entry.Name()))
		if err != nil {
			t.Fatalf("read fixture: %v", err)
		}
		if strings.Contains(strings.ToLower(string(data)), "noul") {
			t.Errorf("fixture %q must not mention the transport-specific term", entry.Name())
		}
	}
}
