package clientmanager

import (
	"testing"

	"github.com/AdamBowers/terraform-provider-pve/v2/internal/diagnostic"
	"github.com/luthermonson/go-proxmox"
)

func TestValidateClient_RunsAllChecks(t *testing.T) {
	var calls int

	checkOne := func(_ *proxmox.Client, _ *diagnostic.Diagnostics) {
		calls++
	}

	checkTwo := func(_ *proxmox.Client, _ *diagnostic.Diagnostics) {
		calls++
	}

	diags := validateClient(&proxmox.Client{}, checkOne, checkTwo)

	if diags == nil {
		t.Fatal("validateClient returned nil")
	}

	if calls != 2 {
		t.Fatalf("checks called %d times, want 2", calls)
	}

	if got := len(diags.Snapshot()); got != 0 {
		t.Fatalf("diagnostic count = %d, want 0", got)
	}
}

func TestValidateClient_ReturnsDiagnosticsFromChecks(t *testing.T) {
	check := func(_ *proxmox.Client, diags *diagnostic.Diagnostics) {
		diags.Append(
			diagnostic.Error("Test Error", "test diagnostic"),
		)
	}

	diags := validateClient(&proxmox.Client{}, check)

	if got := len(diags.Snapshot()); got != 1 {
		t.Fatalf("diagnostic count = %d, want 1", got)
	}

	if !diags.AnySeverity(diagnostic.SEVERITY_Error) {
		t.Fatal("expected an error diagnostic")
	}
}

func TestDefaultClientValidator_AcceptsNonNilClient(t *testing.T) {
	diags := defaultClientValidator(&proxmox.Client{})

	if got := len(diags.Snapshot()); got != 0 {
		t.Fatalf("diagnostic count = %d, want 0", got)
	}

	if diags.AnySeverity(diagnostic.SEVERITY_Error) {
		t.Fatal("did not expect an error diagnostic")
	}
}

func TestDefaultClientValidator_RejectsNilClient(t *testing.T) {
	diags := defaultClientValidator(nil)

	if got := len(diags.Snapshot()); got != 1 {
		t.Fatalf("diagnostic count = %d, want 1", got)
	}

	if !diags.AnySeverity(diagnostic.SEVERITY_Error) {
		t.Fatal("expected an error diagnostic")
	}
}

func TestCheckClientNotNil_AcceptsNonNilClient(t *testing.T) {
	diags := diagnostic.New()

	checkClientNotNil(&proxmox.Client{}, &diags)

	if got := len(diags.Snapshot()); got != 0 {
		t.Fatalf("diagnostic count = %d, want 0", got)
	}
}

func TestCheckClientNotNil_RejectsNilClient(t *testing.T) {
	diags := diagnostic.New()

	checkClientNotNil(nil, &diags)

	if got := len(diags.Snapshot()); got != 1 {
		t.Fatalf("diagnostic count = %d, want 1", got)
	}

	if !diags.AnySeverity(diagnostic.SEVERITY_Error) {
		t.Fatal("expected an error diagnostic")
	}
}
