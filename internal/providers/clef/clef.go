// Package clef adapts the Clef decision model, served through an
// Ollama-compatible backend, to the provider-neutral decision contract.
//
// The request path is:
//
//	decision.Provider -> internal/providers/clef -> POST /v1/systemone -> Ollama -> Clef
//
// /v1/systemone is the CLEF-002-selected transport. The unverified MLX/oMLX
// clef-4bit runtime is NOT wired here and cannot be enabled: only the selected
// transport has an implementation, and Clef is OFF by default.
//
// The adapter only translates and normalizes. It never defines or enforces SOP
// policy; approval gates, execution thresholds, and safety enforcement belong
// to agentic-sop.
//
// Clef-specific concepts (SystemOne wire shapes, the "noul" boolean
// representation, Ollama request/response shapes) are implementation details
// confined to this package and never appear in the public decision package or
// the provider-neutral wire layer.
package clef

import (
	"context"
	"errors"
	"strings"

	"github.com/imhttran/sop-decision-adapters/decision"
)

// ErrDisabled reports that the Clef provider is not explicitly enabled. Clef is
// OFF by default and selection requires explicit operator action.
var ErrDisabled = errors.New("clef: provider is disabled (set CLEF_ENABLED to enable explicitly)")

// Provider implements decision.Provider for Clef.
type Provider struct {
	cfg       Config
	transport transport
}

// New builds a Clef provider. A nil transport defaults to an httpTransport
// pointed at cfg.BaseURL, which is convenient in production and stubbed in
// tests.
//
// New never enables Clef by itself: cfg.Enable (or CLEF_ENABLED via
// NewFromEnv) must be set explicitly by the operator. Construction of a
// disabled provider is allowed so callers can report capability without
// enabling the provider.
func New(cfg Config, tr transport) *Provider {
	cfg = cfg.WithDefaults()
	if tr == nil {
		tr = newHTTPTransport(cfg.BaseURL, cfg.Timeout)
	}
	return &Provider{cfg: cfg, transport: tr}
}

// NewFromEnv builds a Clef provider from environment configuration. Clef is
// disabled unless CLEF_ENABLED is explicitly set.
func NewFromEnv() *Provider {
	return New(ConfigFromEnv(), nil)
}

// Name implements decision.Provider.
func (p *Provider) Name() string { return ProviderName }

// Available implements decision.Provider.
//
// A disabled provider is never available: Clef is OFF by default and transport
// defaults must not implicitly enable it. When enabled, availability
// distinguishes three states without issuing an inference request:
//
//   - backend unreachable (or the request times out)     -> false
//   - backend reachable but the configured model absent  -> false
//   - backend reachable and the configured model present -> true
func (p *Provider) Available(ctx context.Context) bool {
	if !p.cfg.Enabled() {
		return false
	}
	models, err := p.transport.ListModels(ctx)
	if err != nil {
		return false
	}
	return p.hasModel(models)
}

// hasModel reports whether the configured model is present in the reported
// model names. The backend may report names with a tag suffix
// ("clef-flash:latest"), so an exact match or a tag-prefixed match both count.
func (p *Provider) hasModel(models []string) bool {
	want := strings.TrimSpace(p.cfg.Model)
	if want == "" {
		return false
	}
	for _, m := range models {
		m = strings.TrimSpace(m)
		if m == want || strings.HasPrefix(m, want+":") {
			return true
		}
	}
	return false
}

// Decide implements decision.Provider.
//
// It validates the request, translates it into a SystemOne request, performs
// the native POST /v1/systemone call, and normalizes the response back into a
// provider-neutral decision.DecisionResult.
//
// Failures never become successful evidence:
//   - an invalid request or unsupported operation -> ProviderError(KindInvalidRequest)
//   - a disabled provider or unreachable backend -> ProviderError(KindUnavailable)
//   - an undecodable response                     -> ProviderError(KindMalformedResponse)
//   - an indeterminate response                  -> ProviderError(KindMalformedResponse)
func (p *Provider) Decide(ctx context.Context, req decision.DecisionRequest) (decision.DecisionResult, error) {
	if !p.cfg.Enabled() {
		// Clef is OFF by default and selection requires explicit operator
		// action; a disabled adapter must not serve decisions.
		return decision.DecisionResult{}, providerErr(decision.KindUnavailable, ErrDisabled)
	}

	if err := req.Validate(); err != nil {
		return decision.DecisionResult{}, providerErr(decision.KindInvalidRequest, err)
	}

	sysReq, err := BuildSystemOneRequest(p.cfg.Model, req)
	if err != nil {
		return decision.DecisionResult{}, providerErr(decision.KindInvalidRequest, err)
	}

	sysResp, err := p.transport.SystemOne(ctx, sysReq)
	if err != nil {
		return decision.DecisionResult{}, providerErr(classifyTransportError(err), err)
	}

	result, err := NormalizeSystemOneResponse(sysResp, req)
	if err != nil {
		return decision.DecisionResult{}, providerErr(decision.KindMalformedResponse, err)
	}

	return result, nil
}

// providerErr wraps cause in a normalized decision.ProviderError attributed to
// this adapter.
func providerErr(kind decision.ErrorKind, cause error) error {
	return &decision.ProviderError{Provider: ProviderName, Kind: kind, Err: cause}
}

// classifyTransportError maps a transport-level failure to a normalized
// decision.ErrorKind. Connectivity failures become KindUnavailable; responses
// the transport could not decode become KindMalformedResponse; anything else is
// a generic KindProviderFailure.
func classifyTransportError(err error) decision.ErrorKind {
	if err == nil {
		return decision.KindProviderFailure
	}
	if isMalformedResponse(err) {
		return decision.KindMalformedResponse
	}
	if isUnavailable(err) {
		return decision.KindUnavailable
	}
	return decision.KindProviderFailure
}
