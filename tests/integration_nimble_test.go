package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/imhttran/sop-decision-adapters/decision"
	"github.com/imhttran/sop-decision-adapters/internal/providers/nimble"
)

// TestNimbleLiveIntegration validates the decision contract against a real
// Ollama + Nimble installation.
//
// It is skipped unless NIMBLE_INTEGRATION_TEST is set, so the default
// `go test ./...` run stays fully offline. Run it with:
//
//	NIMBLE_INTEGRATION_TEST=1 go test ./...
//
// The test asserts the contract shape (all answers present, result types match
// the requested types, probabilities/confidence in range). It deliberately does
// NOT assert exact judgments such as risk == HIGH, because model decisions may
// change over time.
func TestNimbleLiveIntegration(t *testing.T) {
	if os.Getenv("NIMBLE_INTEGRATION_TEST") == "" {
		t.Skip("set NIMBLE_INTEGRATION_TEST=1 to run the live Nimble integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	p := nimble.NewFromEnv()
	if !p.Available(ctx) {
		t.Fatalf("Nimble model %q is not available at OLLAMA_BASE_URL", os.Getenv("OLLAMA_BASE_URL"))
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
			for choice, p := range answer.Probabilities {
				if p < 0 || p > 1 {
					t.Errorf("question %q: probability %v for %q out of range [0,1]", q.ID, p, choice)
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
