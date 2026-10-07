package tests

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/imhttran/sop-decision-adapters/decision"
	"github.com/imhttran/sop-decision-adapters/internal/providers/julia"
)

// TestJuliaHTTPLiveIntegration exercises the operator-supplied HTTP runtime path:
// Go adapter -> CommandRunner -> tools/julia/http_bridge.py -> Julia HTTP service.
//
// It is opt-in and SKIPS (never fails the default suite) unless
// JULIA_HTTP_INTEGRATION_TEST is set, so `go test ./...` stays fully offline and
// never depends on a running Julia service. Run it with:
//
//	JULIA_HTTP_INTEGRATION_TEST=1 JULIA_URL=http://127.0.0.1:8011 go test ./...
//
// The bridge command is the repository-owned helper resolved relative to the
// module root; JULIA_PYTHON overrides the interpreter. It asserts the CONTRACT
// (answers present, types match, choices allowed, probabilities in range), never
// exact judgments. This is a separate test from the ONNX live test, which it does
// not replace.
func TestJuliaHTTPLiveIntegration(t *testing.T) {
	if os.Getenv("JULIA_HTTP_INTEGRATION_TEST") == "" {
		t.Skip("set JULIA_HTTP_INTEGRATION_TEST=1 (and JULIA_URL) to run the live Julia HTTP integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	bridge := filepath.Join(repoRoot(t), "tools", "julia", "http_bridge.py")
	if _, err := os.Stat(bridge); err != nil {
		t.Skipf("HTTP bridge not found at %s: %v", bridge, err)
	}

	cfg := julia.ConfigFromEnv()
	cfg.InferenceCmd = []string{cfg.Python, bridge}
	cfg.Model = envOr("JULIA_MODEL", "SupersonicLabs/Julia-1")
	provider := julia.New(cfg, nil)

	if !provider.Available(ctx) {
		t.Skipf("HTTP bridge command %q is not resolvable", cfg.InferenceCmd[0])
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

	result, err := provider.Decide(ctx, req)
	if err != nil {
		if errors.Is(err, decision.ErrUnavailable) {
			t.Skipf("Julia HTTP service is not reachable (set JULIA_URL): %v", err)
		}
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

// repoRoot walks up from the working directory to the module root (the directory
// holding go.mod).
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
