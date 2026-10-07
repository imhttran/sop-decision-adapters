// Package benchmark implements the CLEF-008 adapter-side decision-provider
// benchmark: a deterministic, provider-neutral evaluation harness plus a fixed
// benchmark corpus.
//
// The package is evaluation tooling only. It is strictly off the production
// decision path: no production package imports it, it holds no reference to any
// SOP policy writer, and no result it produces can influence routing,
// governance, or approval decisions. Every comparison runs in isolation and is
// recorded; a shadow/source failure can never alter a primary decision.
//
// The harness reuses the CLEF-007 provider-neutral observation vocabulary
// (shadow.Observation, shadow.Reference, shadow.Label) and mirrors that harness's
// containment semantics. It introduces no provider-specific policy logic: the
// only provider-aware file is cells.go, which selects providers strictly through
// the decision.Provider seam.
package benchmark

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	_ "embed"

	"github.com/imhttran/sop-decision-adapters/decision"
)

//go:embed corpus.json
var embeddedCorpus []byte

// Corpus is the deterministic benchmark corpus: a fixed, versioned set of
// choice-decision cases with stable IDs and optional labeled ground truth.
type Corpus struct {
	// Version is the corpus version identifier. It is required and is part of
	// the corpus hash.
	Version string `json:"version"`

	// Cases are the benchmark cases, in evaluation order.
	Cases []Case `json:"cases"`
}

// Case is one benchmark case: a single provider-neutral choice question over a
// state, with optional labeled ground truth used for agreement and calibration.
type Case struct {
	// ID is the stable case identifier. It is required and unique.
	ID string `json:"id"`

	// State is the free-form situation the decision is made against.
	State string `json:"state"`

	// Question is the provider-neutral choice question under evaluation.
	Question decision.Question `json:"question"`

	// Expected is the labeled ground-truth choice. It is present only when
	// GroundTruth is true and must then be one of Question.Choices.
	Expected string `json:"expected,omitempty"`

	// GroundTruth reports whether Expected is a labeled ground-truth outcome.
	// Unlabeled cases carry no Expected and are excluded from agreement and
	// calibration, but still contribute to shape/quality metrics.
	GroundTruth bool `json:"ground_truth"`
}

// EmbeddedCorpus returns a copy of the embedded corpus bytes.
func EmbeddedCorpus() []byte {
	out := make([]byte, len(embeddedCorpus))
	copy(out, embeddedCorpus)
	return out
}

// DefaultCorpus loads the embedded corpus.
func DefaultCorpus() (Corpus, error) { return LoadCorpus(embeddedCorpus) }

// LoadCorpus parses and validates a corpus. A malformed corpus is rejected
// explicitly: the returned error is non-nil and no partially-valid corpus is
// returned.
func LoadCorpus(data []byte) (Corpus, error) {
	var c Corpus
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return Corpus{}, fmt.Errorf("decode corpus: %w", err)
	}
	if err := c.Validate(); err != nil {
		return Corpus{}, err
	}
	return c, nil
}

// Validate reports whether the corpus is well-formed. It rejects empty
// versions, empty case lists, missing or duplicate case IDs, non-choice
// questions, and inconsistent ground-truth labels.
func (c Corpus) Validate() error {
	if strings.TrimSpace(c.Version) == "" {
		return errors.New("corpus: version is required")
	}
	if len(c.Cases) == 0 {
		return errors.New("corpus: at least one case is required")
	}

	seen := make(map[string]struct{}, len(c.Cases))
	for i, cs := range c.Cases {
		if strings.TrimSpace(cs.ID) == "" {
			return fmt.Errorf("corpus: case %d: id is required", i)
		}
		if _, dup := seen[cs.ID]; dup {
			return fmt.Errorf("corpus: duplicate case id %q", cs.ID)
		}
		seen[cs.ID] = struct{}{}

		if strings.TrimSpace(cs.State) == "" {
			return fmt.Errorf("corpus: case %q: state is required", cs.ID)
		}
		if err := cs.Question.Validate(); err != nil {
			return fmt.Errorf("corpus: case %q: %w", cs.ID, err)
		}
		if cs.Question.Type != decision.QuestionChoice {
			return fmt.Errorf("corpus: case %q: the benchmark corpus supports choice questions only", cs.ID)
		}

		switch {
		case cs.GroundTruth && strings.TrimSpace(cs.Expected) == "":
			return fmt.Errorf("corpus: case %q: ground_truth requires a non-empty expected choice", cs.ID)
		case cs.GroundTruth && !cs.Question.Allows(cs.Expected):
			return fmt.Errorf("corpus: case %q: expected choice %q is not in the allowed set", cs.ID, cs.Expected)
		case !cs.GroundTruth && cs.Expected != "":
			return fmt.Errorf("corpus: case %q: an unlabeled case must not carry an expected choice", cs.ID)
		}
	}
	return nil
}

// Hash returns the deterministic SHA-256 (hex) of the canonical corpus
// encoding. It is stable across processes and independent of source
// whitespace.
func (c Corpus) Hash() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("hash corpus: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// LabeledCases returns the number of cases carrying labeled ground truth.
func (c Corpus) LabeledCases() int {
	n := 0
	for _, cs := range c.Cases {
		if cs.GroundTruth {
			n++
		}
	}
	return n
}
