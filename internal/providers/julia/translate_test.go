package julia

import (
	"errors"
	"testing"

	"github.com/imhttran/sop-decision-adapters/decision"
)

// fptr returns a pointer to v, for constructing optional answer fields.
func fptr(v float64) *float64 { return &v }

// exampleRequest is the shared risk-evaluation fixture used across the Julia
// tests: a choice question (LOW/MEDIUM/HIGH), a choice question
// (SMALL/MEDIUM/LARGE), and a boolean question.
func exampleRequest() decision.DecisionRequest {
	return decision.DecisionRequest{
		State: "deploying a schema migration during business hours",
		Questions: []decision.Question{
			{
				ID:           "risk",
				Type:         decision.QuestionChoice,
				Criteria:     "Overall risk of the change",
				Choices:      []string{"LOW", "MEDIUM", "HIGH"},
				Instructions: "Assess deployment risk.",
			},
			{
				ID:       "execution_model",
				Type:     decision.QuestionChoice,
				Criteria: "Preferred execution model",
				Choices:  []string{"SMALL", "MEDIUM", "LARGE"},
			},
			{
				ID:           "approval_required",
				Type:         decision.QuestionBoolean,
				Instructions: "Is human approval required?",
			},
		},
	}
}

// okOutputs is a valid raw Outputs for exampleRequest.
func okOutputs() Outputs {
	return Outputs{
		Model: "julia",
		Answers: map[string]QuestionOutput{
			"risk": {
				Choice:        "HIGH",
				Confidence:    fptr(0.83),
				Probabilities: map[string]float64{"HIGH": 0.95, "MEDIUM": 0.04, "LOW": 0.01},
			},
			"execution_model": {
				Choice:     "LARGE",
				Confidence: fptr(0.84),
			},
			"approval_required": {
				Probability: fptr(0.997),
			},
		},
	}
}

func TestBuildInputs(t *testing.T) {
	tests := map[string]struct {
		req      decision.DecisionRequest
		wantText map[string]string
	}{
		"instructions preferred": {
			req: decision.DecisionRequest{
				State: "s",
				Questions: []decision.Question{{
					ID:           "risk",
					Type:         decision.QuestionChoice,
					Criteria:     "criteria text",
					Instructions: "instructions text",
					Choices:      []string{"LOW", "HIGH"},
				}},
			},
			wantText: map[string]string{"risk": "instructions text"},
		},
		"criteria fallback": {
			req: decision.DecisionRequest{
				State: "s",
				Questions: []decision.Question{{
					ID:       "risk",
					Type:     decision.QuestionChoice,
					Criteria: "criteria text",
					Choices:  []string{"LOW", "HIGH"},
				}},
			},
			wantText: map[string]string{"risk": "criteria text"},
		},
		"id fallback": {
			req: decision.DecisionRequest{
				State: "s",
				Questions: []decision.Question{{
					ID:   "approval_required",
					Type: decision.QuestionBoolean,
				}},
			},
			wantText: map[string]string{"approval_required": "Decide approval_required."},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := BuildInputs(tt.req)
			if err != nil {
				t.Fatalf("BuildInputs() error = %v", err)
			}
			if got.State != tt.req.State {
				t.Errorf("State = %q, want %q", got.State, tt.req.State)
			}
			if len(got.Questions) != len(tt.req.Questions) {
				t.Fatalf("len(Questions) = %d, want %d", len(got.Questions), len(tt.req.Questions))
			}
			for _, qi := range got.Questions {
				want, ok := tt.wantText[qi.ID]
				if !ok {
					continue
				}
				if qi.Text != want {
					t.Errorf("question %q Text = %q, want %q", qi.ID, qi.Text, want)
				}
			}
		})
	}
}

func TestBuildInputsCopiesChoiceLabels(t *testing.T) {
	got, err := BuildInputs(exampleRequest())
	if err != nil {
		t.Fatalf("BuildInputs() error = %v", err)
	}
	byID := map[string]QuestionInput{}
	for _, qi := range got.Questions {
		byID[qi.ID] = qi
	}
	if labels := byID["risk"].Labels; len(labels) != 3 {
		t.Fatalf("risk Labels = %v, want 3 labels", labels)
	}
	if byID["risk"].Type != decision.QuestionChoice {
		t.Errorf("risk Type = %q, want choice", byID["risk"].Type)
	}
	if labels := byID["approval_required"].Labels; len(labels) != 0 {
		t.Errorf("boolean Labels = %v, want none", labels)
	}
}

func TestBuildInputsInvalidRequest(t *testing.T) {
	if _, err := BuildInputs(decision.DecisionRequest{}); !errors.Is(err, decision.ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
}

func TestNormalizeOutputsHappyPath(t *testing.T) {
	got, err := NormalizeOutputs(okOutputs(), exampleRequest())
	if err != nil {
		t.Fatalf("NormalizeOutputs() error = %v", err)
	}
	if got.Provider != ProviderName {
		t.Errorf("Provider = %q, want %q", got.Provider, ProviderName)
	}
	if got.Model != "julia" {
		t.Errorf("Model = %q, want julia", got.Model)
	}
	if len(got.Answers) != 3 {
		t.Fatalf("len(Answers) = %d, want 3", len(got.Answers))
	}
	if got.Answers["risk"].Choice != "HIGH" {
		t.Errorf("risk = %+v", got.Answers["risk"])
	}
	if got.Answers["execution_model"].Choice != "LARGE" {
		t.Errorf("execution_model = %+v", got.Answers["execution_model"])
	}
	if got.Answers["approval_required"].Probability != 0.997 {
		t.Errorf("approval_required = %+v", got.Answers["approval_required"])
	}
	if err := got.Validate(); err != nil {
		t.Errorf("result.Validate() = %v", err)
	}
}

func TestNormalizeOutputsRejections(t *testing.T) {
	base := exampleRequest()

	tests := map[string]struct {
		mutate func(*Outputs)
	}{
		"missing answer": {
			mutate: func(o *Outputs) { delete(o.Answers, "risk") },
		},
		"choice outside allowed": {
			mutate: func(o *Outputs) { o.Answers["risk"] = QuestionOutput{Choice: "CRITICAL"} },
		},
		"incompatible type (choice answer without choice)": {
			mutate: func(o *Outputs) {
				o.Answers["risk"] = QuestionOutput{Probability: fptr(0.5)}
			},
		},
		"unknown probability key": {
			mutate: func(o *Outputs) {
				o.Answers["risk"] = QuestionOutput{Choice: "HIGH", Probabilities: map[string]float64{"CRITICAL": 0.5}}
			},
		},
		"probability below zero": {
			mutate: func(o *Outputs) {
				o.Answers["risk"] = QuestionOutput{Choice: "HIGH", Probabilities: map[string]float64{"HIGH": -0.1}}
			},
		},
		"probability above one": {
			mutate: func(o *Outputs) {
				o.Answers["risk"] = QuestionOutput{Choice: "HIGH", Probabilities: map[string]float64{"HIGH": 1.1}}
			},
		},
		"confidence out of range": {
			mutate: func(o *Outputs) { o.Answers["risk"] = QuestionOutput{Choice: "HIGH", Confidence: fptr(1.5)} },
		},
		"boolean with no probability": {
			mutate: func(o *Outputs) { o.Answers["approval_required"] = QuestionOutput{} },
		},
		"score with no value": {
			mutate: func(o *Outputs) { o.Answers["risk"] = QuestionOutput{} },
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			out := okOutputs()
			tt.mutate(&out)
			_, err := NormalizeOutputs(out, base)
			if !errors.Is(err, decision.ErrMalformedResponse) {
				t.Fatalf("error = %v, want ErrMalformedResponse", err)
			}
		})
	}
}

func TestNormalizeOutputsScoreQuestion(t *testing.T) {
	req := decision.DecisionRequest{
		State: "s",
		Questions: []decision.Question{{
			ID:       "confidence_score",
			Type:     decision.QuestionScore,
			Criteria: "Score the change",
		}},
	}
	out := Outputs{
		Model: "julia",
		Answers: map[string]QuestionOutput{
			"confidence_score": {Score: fptr(7.5), Confidence: fptr(0.9)},
		},
	}
	got, err := NormalizeOutputs(out, req)
	if err != nil {
		t.Fatalf("NormalizeOutputs() error = %v", err)
	}
	if got.Answers["confidence_score"].Score != 7.5 {
		t.Errorf("score = %+v", got.Answers["confidence_score"])
	}
}

func TestNormalizeOutputsInvalidRequest(t *testing.T) {
	if _, err := NormalizeOutputs(Outputs{}, decision.DecisionRequest{}); !errors.Is(err, decision.ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
}

func TestNormalizeOutputsDoesNotRequireSum(t *testing.T) {
	// Probabilities that do not sum to 1.0 must still be accepted.
	out := okOutputs()
	out.Answers["risk"] = QuestionOutput{
		Choice:        "HIGH",
		Probabilities: map[string]float64{"HIGH": 0.5, "LOW": 0.1},
	}
	if _, err := NormalizeOutputs(out, exampleRequest()); err != nil {
		t.Fatalf("NormalizeOutputs() error = %v, want success for non-normalized probabilities", err)
	}
}
