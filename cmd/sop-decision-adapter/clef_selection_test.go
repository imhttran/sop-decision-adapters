package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/imhttran/sop-decision-adapters/decision"
	"github.com/imhttran/sop-decision-adapters/internal/providers/clef"
)

// TestNewProviderClefOffByDefault proves that selecting the clef provider
// through the in-process selection seam constructs a Clef provider but does NOT
// enable it unless CLEF_ENABLED is explicitly set.
func TestNewProviderClefOffByDefault(t *testing.T) {
	t.Setenv("CLEF_ENABLED", "")

	p, err := newProvider("clef", "", "", time.Second)
	if err != nil {
		t.Fatalf("newProvider(clef): unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("newProvider(clef): got nil provider")
	}
	if got := p.Name(); got != clef.ProviderName {
		t.Fatalf("provider name = %q, want %q", got, clef.ProviderName)
	}

	if p.Available(context.Background()) {
		t.Fatal("clef provider reports Available=true without CLEF_ENABLED; must be OFF by default")
	}

	_, err = p.Decide(context.Background(), decision.DecisionRequest{})
	if err == nil {
		t.Fatal("Decide on disabled clef provider: expected error, got nil")
	}
	if !errors.Is(err, decision.ErrUnavailable) {
		t.Fatalf("Decide error = %v, want ErrUnavailable", err)
	}
}

// TestNewProviderClefMixedCase proves name resolution matches the existing
// lower/trim handling.
func TestNewProviderClefMixedCase(t *testing.T) {
	t.Setenv("CLEF_ENABLED", "")

	p, err := newProvider("  Clef ", "", "", time.Second)
	if err != nil {
		t.Fatalf("newProvider(Clef): unexpected error: %v", err)
	}
	if p == nil || p.Name() != clef.ProviderName {
		t.Fatalf("mixed-case clef did not resolve to the Clef provider")
	}
}

// TestNewProviderClefExplicitlyEnabled proves that with CLEF_ENABLED set, the
// clef provider is selected and constructed with Enable=true (off-by-default is
// opt-in only).
func TestNewProviderClefExplicitlyEnabled(t *testing.T) {
	t.Setenv("CLEF_ENABLED", "1")
	t.Setenv("CLEF_MODEL", "clef-flash")
	t.Setenv("CLEF_TIMEOUT", "45s")

	p, err := newProvider("clef", "http://example.invalid:11434", "clef-override", 5*time.Second)
	if err != nil {
		t.Fatalf("newProvider(clef): unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("newProvider(clef): got nil provider")
	}
	if got := p.Name(); got != clef.ProviderName {
		t.Fatalf("provider name = %q, want %q", got, clef.ProviderName)
	}

	// The provider is explicitly enabled: Decide must not fail with ErrDisabled.
	// It may still fail for transport reasons (no live backend here), but the
	// failure must not be the disabled gate.
	_, err = p.Decide(context.Background(), decision.DecisionRequest{})
	if errors.Is(err, clef.ErrDisabled) {
		t.Fatalf("enabled clef provider returned ErrDisabled: %v", err)
	}
}

// TestNewProviderUnknownFailsClosed proves unknown providers fail closed rather
// than falling through to any default.
func TestNewProviderUnknownFailsClosed(t *testing.T) {
	p, err := newProvider("not-a-provider", "", "", time.Second)
	if err == nil {
		t.Fatal("unknown provider: expected error, got nil")
	}
	if p != nil {
		t.Fatalf("unknown provider: expected nil provider, got %v", p)
	}
	if !strings.Contains(err.Error(), "unsupported provider") {
		t.Fatalf("unknown provider error = %v, want unsupported provider", err)
	}
}

// TestAvailableProvidersIncludesClef pins the advertised selection list.
func TestAvailableProvidersIncludesClef(t *testing.T) {
	if !strings.Contains(availableProviders, "clef") {
		t.Fatalf("availableProviders = %q, want it to include \"clef\"", availableProviders)
	}
	if !strings.Contains(availableProviders, "nimble") || !strings.Contains(availableProviders, "julia") {
		t.Fatalf("availableProviders = %q, want it to still include nimble and julia", availableProviders)
	}
}

// TestNewProviderNimbleAndJuliaUnchanged proves the existing selection is
// preserved alongside the new clef case.
func TestNewProviderNimbleAndJuliaUnchanged(t *testing.T) {
	np, err := newProvider("nimble", "", "", time.Second)
	if err != nil || np == nil || np.Name() != "nimble" {
		t.Fatalf("newProvider(nimble) = %v, %v; want nimble provider", np, err)
	}

	jp, err := newProvider("julia", "", "", time.Second)
	if err != nil || jp == nil || jp.Name() != "julia" {
		t.Fatalf("newProvider(julia) = %v, %v; want julia provider", jp, err)
	}
}
