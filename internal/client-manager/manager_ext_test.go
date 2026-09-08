package clientmanager_test

import (
	"fmt"
	"sync"
	"testing"

	clientmanager "github.com/AdamBowers/terraform-provider-pve/v2/internal/client-manager"
	"github.com/AdamBowers/terraform-provider-pve/v2/internal/diagnostic"
	"github.com/luthermonson/go-proxmox"
)

func newTestClient() *proxmox.Client {
	return &proxmox.Client{}
}

func newTestDiagnostics() *diagnostic.Diagnostics {
	diags := diagnostic.New()
	return &diags
}

func TestClientManager_New_ReturnsUsableManager(t *testing.T) {
	cm := clientmanager.New()

	if cm == nil {
		t.Fatal("New returned nil")
	}

	client := newTestClient()
	diags := newTestDiagnostics()

	if ok := cm.Add("primary", client, diags); !ok {
		t.Fatal("Add returned false")
	}

	if got := cm.Get("primary", diags); got != client {
		t.Fatal("Get returned a different client")
	}

	if got := len(diags.Snapshot()); got != 0 {
		t.Fatalf("diagnostic count = %d, want 0", got)
	}
}

func TestClientManager_Get_ReturnsExistingClient(t *testing.T) {
	cm := clientmanager.New()
	client := newTestClient()
	diags := newTestDiagnostics()

	if ok := cm.Add("primary", client, diags); !ok {
		t.Fatal("Add returned false")
	}

	got := cm.Get("primary", diags)

	if got != client {
		t.Fatal("Get returned a different client")
	}

	if got := len(diags.Snapshot()); got != 0 {
		t.Fatalf("diagnostic count = %d, want 0", got)
	}
}

func TestClientManager_Get_ReturnsErrorWhenClientDoesNotExist(t *testing.T) {
	cm := clientmanager.New()
	diags := newTestDiagnostics()

	got := cm.Get("missing", diags)

	if got != nil {
		t.Fatalf("Get returned %#v, want nil", got)
	}

	results := diags.Snapshot()

	if len(results) != 1 {
		t.Fatalf("diagnostic count = %d, want 1", len(results))
	}

	if !diags.AnySeverity(diagnostic.SEVERITY_Error) {
		t.Fatal("expected an error diagnostic")
	}

	if diags.AnySeverity(diagnostic.SEVERITY_Warning) {
		t.Fatal("did not expect a warning diagnostic")
	}
}

func TestClientManager_TryGet_ReturnsExistingClient(t *testing.T) {
	cm := clientmanager.New()
	client := newTestClient()
	diags := newTestDiagnostics()

	if ok := cm.Add("primary", client, diags); !ok {
		t.Fatal("Add returned false")
	}

	got := cm.TryGet("primary", diags)

	if got != client {
		t.Fatal("TryGet returned a different client")
	}

	if got := len(diags.Snapshot()); got != 0 {
		t.Fatalf("diagnostic count = %d, want 0", got)
	}
}

func TestClientManager_TryGet_ReturnsWarningWhenClientDoesNotExist(t *testing.T) {
	cm := clientmanager.New()
	diags := newTestDiagnostics()

	got := cm.TryGet("missing", diags)

	if got != nil {
		t.Fatalf("TryGet returned %#v, want nil", got)
	}

	results := diags.Snapshot()

	if len(results) != 1 {
		t.Fatalf("diagnostic count = %d, want 1", len(results))
	}

	if !diags.AnySeverity(diagnostic.SEVERITY_Warning) {
		t.Fatal("expected a warning diagnostic")
	}

	if diags.AnySeverity(diagnostic.SEVERITY_Error) {
		t.Fatal("did not expect an error diagnostic")
	}
}

func TestClientManager_Add_AddsValidClient(t *testing.T) {
	cm := clientmanager.New()
	client := newTestClient()
	addDiags := newTestDiagnostics()

	if ok := cm.Add("primary", client, addDiags); !ok {
		t.Fatal("Add returned false")
	}

	if got := len(addDiags.Snapshot()); got != 0 {
		t.Fatalf("diagnostic count = %d, want 0", got)
	}

	getDiags := newTestDiagnostics()
	if got := cm.Get("primary", getDiags); got != client {
		t.Fatal("Add did not make the client retrievable")
	}

	if got := len(getDiags.Snapshot()); got != 0 {
		t.Fatalf("Get diagnostic count = %d, want 0", got)
	}
}

func TestClientManager_Add_RejectsNilClient(t *testing.T) {
	cm := clientmanager.New()
	diags := newTestDiagnostics()

	if ok := cm.Add("primary", nil, diags); ok {
		t.Fatal("Add accepted a nil client")
	}

	if got := len(diags.Snapshot()); got != 1 {
		t.Fatalf("diagnostic count = %d, want 1", got)
	}

	if !diags.AnySeverity(diagnostic.SEVERITY_Error) {
		t.Fatal("expected an error diagnostic")
	}

	getDiags := newTestDiagnostics()
	if got := cm.Get("primary", getDiags); got != nil {
		t.Fatalf("nil client was stored: %#v", got)
	}
}

func TestClientManager_Add_RejectsDuplicateNameWithoutReplacingClient(t *testing.T) {
	cm := clientmanager.New()
	first := newTestClient()
	second := newTestClient()

	initialDiags := newTestDiagnostics()
	if ok := cm.Add("primary", first, initialDiags); !ok {
		t.Fatal("initial Add returned false")
	}

	duplicateDiags := newTestDiagnostics()
	if ok := cm.Add("primary", second, duplicateDiags); ok {
		t.Fatal("duplicate Add returned true")
	}

	if got := len(duplicateDiags.Snapshot()); got != 1 {
		t.Fatalf("diagnostic count = %d, want 1", got)
	}

	if !duplicateDiags.AnySeverity(diagnostic.SEVERITY_Error) {
		t.Fatal("expected an error diagnostic")
	}

	getDiags := newTestDiagnostics()
	if got := cm.Get("primary", getDiags); got != first {
		t.Fatal("duplicate Add replaced the original client")
	}
}

func TestClientManager_Set_StoresValidClient(t *testing.T) {
	cm := clientmanager.New()
	client := newTestClient()
	setDiags := newTestDiagnostics()

	cm.Set("primary", client, setDiags)

	if got := len(setDiags.Snapshot()); got != 0 {
		t.Fatalf("diagnostic count = %d, want 0", got)
	}

	getDiags := newTestDiagnostics()
	if got := cm.Get("primary", getDiags); got != client {
		t.Fatal("Set did not make the client retrievable")
	}

	if got := len(getDiags.Snapshot()); got != 0 {
		t.Fatalf("Get diagnostic count = %d, want 0", got)
	}
}

func TestClientManager_Set_OverwritesExistingClient(t *testing.T) {
	cm := clientmanager.New()
	first := newTestClient()
	second := newTestClient()

	addDiags := newTestDiagnostics()
	if ok := cm.Add("primary", first, addDiags); !ok {
		t.Fatal("initial Add returned false")
	}

	setDiags := newTestDiagnostics()
	cm.Set("primary", second, setDiags)

	if got := len(setDiags.Snapshot()); got != 0 {
		t.Fatalf("diagnostic count = %d, want 0", got)
	}

	getDiags := newTestDiagnostics()
	if got := cm.Get("primary", getDiags); got != second {
		t.Fatal("Set did not overwrite the existing client")
	}
}

func TestClientManager_Set_RejectsNilClientWithoutReplacingExistingClient(t *testing.T) {
	cm := clientmanager.New()
	client := newTestClient()

	initialDiags := newTestDiagnostics()
	if ok := cm.Add("primary", client, initialDiags); !ok {
		t.Fatal("initial Add returned false")
	}

	setDiags := newTestDiagnostics()
	cm.Set("primary", nil, setDiags)

	if got := len(setDiags.Snapshot()); got != 1 {
		t.Fatalf("diagnostic count = %d, want 1", got)
	}

	if !setDiags.AnySeverity(diagnostic.SEVERITY_Error) {
		t.Fatal("expected an error diagnostic")
	}

	getDiags := newTestDiagnostics()
	if got := cm.Get("primary", getDiags); got != client {
		t.Fatal("invalid Set replaced the existing client")
	}
}

func TestClientManager_ZeroValue_IsUsable(t *testing.T) {
	var cm clientmanager.ClientManager
	client := newTestClient()
	addDiags := newTestDiagnostics()

	if ok := cm.Add("primary", client, addDiags); !ok {
		t.Fatal("Add returned false for zero-value ClientManager")
	}

	getDiags := newTestDiagnostics()
	if got := cm.Get("primary", getDiags); got != client {
		t.Fatal("Get returned a different client")
	}

	if got := len(addDiags.Snapshot()); got != 0 {
		t.Fatalf("Add diagnostic count = %d, want 0", got)
	}

	if got := len(getDiags.Snapshot()); got != 0 {
		t.Fatalf("Get diagnostic count = %d, want 0", got)
	}
}

func TestClientManager_PublicMethods_SupportNames(t *testing.T) {
	names := []string{
		"",
		" ",
		"primary",
		"unicode-名前",
		`quoted"name`,
		"name/with/slashes",
		"name\nwith\nnewlines",
	}

	for _, name := range names {
		t.Run(fmt.Sprintf("%q", name), func(t *testing.T) {
			cm := clientmanager.New()
			client := newTestClient()
			addDiags := newTestDiagnostics()

			if ok := cm.Add(name, client, addDiags); !ok {
				t.Fatal("Add returned false")
			}

			getDiags := newTestDiagnostics()
			if got := cm.Get(name, getDiags); got != client {
				t.Fatal("Get returned a different client")
			}

			if got := len(addDiags.Snapshot()); got != 0 {
				t.Fatalf("Add diagnostic count = %d, want 0", got)
			}

			if got := len(getDiags.Snapshot()); got != 0 {
				t.Fatalf("Get diagnostic count = %d, want 0", got)
			}
		})
	}
}

func TestClientManager_PublicMethods_AreSafeForConcurrentUse(t *testing.T) {
	var cm clientmanager.ClientManager

	const (
		workers    = 32
		iterations = 1000
	)

	var wg sync.WaitGroup
	wg.Add(workers)

	for worker := 0; worker < workers; worker++ {
		go func(worker int) {
			defer wg.Done()

			for iteration := 0; iteration < iterations; iteration++ {
				name := fmt.Sprintf("client-%d", (worker+iteration)%8)
				client := newTestClient()

				switch iteration % 4 {
				case 0:
					diags := newTestDiagnostics()
					cm.Set(name, client, diags)
				case 1:
					diags := newTestDiagnostics()
					cm.Add(name, client, diags)
				case 2:
					diags := newTestDiagnostics()
					_ = cm.Get(name, diags)
				default:
					diags := newTestDiagnostics()
					_ = cm.TryGet(name, diags)
				}
			}
		}(worker)
	}

	wg.Wait()
}
