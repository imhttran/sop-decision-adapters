package benchmark

import (
	"context"
	"fmt"

	"github.com/imhttran/sop-decision-adapters/decision"
	"github.com/imhttran/sop-decision-adapters/internal/providers/clef"
	"github.com/imhttran/sop-decision-adapters/internal/providers/nimble"
	"github.com/imhttran/sop-decision-adapters/internal/shadow"
)

// ProviderSource adapts any decision.Provider to the benchmark Source contract.
// It reaches providers only through decision.Provider: no provider-specific
// transport, wire shape, or policy is referenced, so provider-specific transport
// stays behind the existing provider boundary. This is the only file in the
// package that is aware of concrete providers.
type ProviderSource struct {
	Provider decision.Provider
}

// Name reports the underlying provider name.
func (s ProviderSource) Name() string { return s.Provider.Name() }

// Available reports whether the underlying provider is currently reachable. It
// performs no inference.
func (s ProviderSource) Available(ctx context.Context) bool { return s.Provider.Available(ctx) }

// Comparable reports whether the case can be evaluated: a choice question with a
// non-empty allowed-choice set. This is the same comparability predicate used by
// the CLEF-007 harness.
func (s ProviderSource) Comparable(c Case) bool {
	return c.Question.Type == decision.QuestionChoice && len(c.Question.Choices) > 0
}

// Observe issues one provider-neutral decision request for the case and returns
// the normalized observation. An omitted answer is reported as a malformed
// response rather than silently treated as a choice.
func (s ProviderSource) Observe(ctx context.Context, c Case) (shadow.Observation, error) {
	req := decision.DecisionRequest{State: c.State, Questions: []decision.Question{c.Question}}
	res, err := s.Provider.Decide(ctx, req)
	if err != nil {
		return shadow.Observation{}, err
	}
	ans, ok := res.Answers[c.Question.ID]
	if !ok {
		return shadow.Observation{}, fmt.Errorf("%w: provider omitted an answer for question %q", decision.ErrMalformedResponse, c.Question.ID)
	}
	return shadow.Observation{Choice: ans.Choice, Confidence: ans.Confidence}, nil
}

// DefaultCells returns the CLEF-008 benchmark matrix built from the environment.
//
// The primary matrix selects the Clef and Nimble cells through the existing
// provider-neutral seam. The conditional quantization/fidelity cells (oMLX
// clef-4bit, Clef 8-bit) are recorded as not exercised because their
// prerequisites are unmet (CLEF-002 did not prove a fidelity-preserving Clef
// decision transport for oMLX).
func DefaultCells() []Cell {
	return []Cell{
		{ID: "clef", Matrix: "primary", Source: ProviderSource{Provider: clef.NewFromEnv()}},
		{ID: "nimble", Matrix: "primary", Source: ProviderSource{Provider: nimble.NewFromEnv()}},
		{
			ID:     "omlx-clef-4bit",
			Matrix: "conditional",
			Reason: "CLEF-002 did not prove a fidelity-preserving Clef decision transport for oMLX (/v1/systemone returns HTTP 404); prerequisite unmet, not exercised",
		},
		{
			ID:     "clef-8bit",
			Matrix: "conditional",
			Reason: "conditional quantization/fidelity cell exercised only against a measurable primary baseline; not exercised",
		},
	}
}
