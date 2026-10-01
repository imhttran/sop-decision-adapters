package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/imhttran/sop-decision-adapters/decision"
	"github.com/imhttran/sop-decision-adapters/internal/providers/julia"
)

// TestJuliaLiveIntegration is the Phase 2.1 acceptance test: it drives the real
// Julia-1-ONNX model through the repository-owned helper and the Julia adapter.
//
// It is skipped unless JULIA_INTEGRATION_TEST is set, so the default
// `go test ./...` run stays fully offline (no Python, ONNX Runtime, model, or
// network). Run it with:
//
//	JULIA_INTEGRATION_TEST=1 JULIA_MODEL_PATH=/path/to/julia-1.onnx go test ./...
//
// The test asserts the CONTRACT (all answers present, answer types match the
// requested types, selected choices are allowed, probabilities/confidence in
// range). It deliberately does NOT assert exact judgments such as risk == HIGH,
// because model decisions may vary. When the runtime is not configured the test
// SKIPS rather than failing the suite.
func TestJuliaLiveIntegration(t *testing.T) {
	if os.Getenv("JULIA_INTEGRATION_TEST") == "" {
		t.Skip("set JULIA_INTEGRATION_TEST=1 (and JULIA_MODEL_PATH) to run the live Julia integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	p := julia.NewFromEnv()
	if !p.Available(ctx) {
		t.Skipf("Julia provider is not available (configure JULIA_MODEL_PATH and the helper dependencies): JULIA_MODEL_PATH=%q",
			os.Getenv("JULIA_MODEL_PATH"))
	}

	req := decision.DecisionRequest{
		State: "A schema migration will execute against production during business hours.",
		Questions: []decision.Question{
			{
				ID:       "risk",
				Type:     decision.QuestionChoice,
				Criteria: "Assess the operational risk of the described change.",
				Choices:  []string{"LOW", "MEDIUM", "HIGH"},
			},
			{
				ID:   "approval_required",
				Type: decision.QuestionBoolean,
			},
			{
				ID:       "execution_model",
				Type:     decision.QuestionChoice,
				Criteria: "Select the execution model.",
				Choices:  []string{"SMALL", "MEDIUM", "LARGE"},
			},
		},
	}

	result, err := p.Decide(ctx, req)
	if err != nil {
		t.Fatalf("Decide() error = %v", err)
	}

	for _, q := range req.Questions {
		answer, ok := result.Answers[q.ID]
		if !ok {
			t.Errorf("missing answer for question %q", q.ID)
			continue
		}
		switch q.Type {
		case decision.QuestionChoice:
			if answer.Type != decision.AnswerChoice {
				t.Errorf("question %q: answer type = %q, want choice", q.ID, answer.Type)
				continue
			}
			if !q.Allows(answer.Choice) {
				t.Errorf("question %q: choice %q is not one of %v", q.ID, answer.Choice, q.Choices)
			}
			for choice, prob := range answer.Probabilities {
				if !q.Allows(choice) {
					t.Errorf("question %q: probability for unknown choice %q", q.ID, choice)
				}
				if prob < 0 || prob > 1 {
					t.Errorf("question %q: probability %v for %q out of range [0,1]", q.ID, prob, choice)
				}
			}
		case decision.QuestionBoolean:
			if answer.Type != decision.AnswerBoolean {
				t.Errorf("question %q: answer type = %q, want boolean", q.ID, answer.Type)
				continue
			}
			if answer.Probability < 0 || answer.Probability > 1 {
				t.Errorf("question %q: probability %v out of range [0,1]", q.ID, answer.Probability)
			}
		case decision.QuestionScore:
			if answer.Type != decision.AnswerScore {
				t.Errorf("question %q: answer type = %q, want score", q.ID, answer.Type)
			}
		}
		if answer.Confidence != nil && (*answer.Confidence < 0 || *answer.Confidence > 1) {
			t.Errorf("question %q: confidence %v out of range [0,1]", q.ID, *answer.Confidence)
		}
	}
}
