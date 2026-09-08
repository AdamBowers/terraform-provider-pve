package clientmanager

import (
	"fmt"
	"sync"
	"testing"

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

func TestClientManager_New_InitializesManager(t *testing.T) {
	cm := New()

	if cm == nil {
		t.Fatal("New returned nil")
	}

	if cm.clients == nil {
		t.Fatal("New initialized a nil clients map")
	}

	if got := len(cm.clients); got != 0 {
		t.Fatalf("client count = %d, want 0", got)
	}
}

func TestClientManager_get_ReturnsExistingClient(t *testing.T) {
	cm := New()
	client := newTestClient()

	cm.clients["primary"] = client

	got := cm.get("primary")

	if got != client {
		t.Fatal("get returned the wrong client")
	}
}

func TestClientManager_get_ReturnsNilWhenClientDoesNotExist(t *testing.T) {
	cm := New()

	got := cm.get("missing")

	if got != nil {
		t.Fatalf("get returned %#v, want nil", got)
	}
}

func TestClientManager_get_DoesNotInitializeNilMap(t *testing.T) {
	var cm ClientManager

	got := cm.get("missing")

	if got != nil {
		t.Fatalf("get returned %#v, want nil", got)
	}

	if cm.clients != nil {
		t.Fatal("get initialized the clients map")
	}
}

func TestClientManager_set_StoresClient(t *testing.T) {
	cm := New()
	client := newTestClient()

	cm.set("primary", client)

	if got := cm.clients["primary"]; got != client {
		t.Fatal("set did not store the client")
	}
}

func TestClientManager_set_OverwritesExistingClient(t *testing.T) {
	cm := New()
	first := newTestClient()
	second := newTestClient()

	cm.clients["primary"] = first
	cm.set("primary", second)

	if got := cm.clients["primary"]; got != second {
		t.Fatal("set did not overwrite the existing client")
	}
}

func TestClientManager_set_InitializesNilMap(t *testing.T) {
	var cm ClientManager
	client := newTestClient()

	cm.set("primary", client)

	if cm.clients == nil {
		t.Fatal("set did not initialize the clients map")
	}

	if got := cm.clients["primary"]; got != client {
		t.Fatal("set did not store the client")
	}
}

func TestClientManager_add_AddsClient(t *testing.T) {
	cm := New()
	client := newTestClient()

	if ok := cm.add("primary", client); !ok {
		t.Fatal("add returned false")
	}

	if got := cm.clients["primary"]; got != client {
		t.Fatal("add did not store the client")
	}
}

func TestClientManager_add_RejectsDuplicateName(t *testing.T) {
	cm := New()
	first := newTestClient()
	second := newTestClient()

	cm.clients["primary"] = first

	if ok := cm.add("primary", second); ok {
		t.Fatal("add returned true for a duplicate name")
	}

	if got := cm.clients["primary"]; got != first {
		t.Fatal("duplicate add replaced the existing client")
	}
}

func TestClientManager_add_InitializesNilMap(t *testing.T) {
	var cm ClientManager
	client := newTestClient()

	if ok := cm.add("primary", client); !ok {
		t.Fatal("add returned false")
	}

	if cm.clients == nil {
		t.Fatal("add did not initialize the clients map")
	}

	if got := cm.clients["primary"]; got != client {
		t.Fatal("add did not store the client")
	}
}

func TestClientManager_Get_ReturnsExistingClient(t *testing.T) {
	cm := New()
	client := newTestClient()
	diags := newTestDiagnostics()

	cm.set("primary", client)

	got := cm.Get("primary", diags)

	if got != client {
		t.Fatal("Get returned the wrong client")
	}

	if got := len(diags.Snapshot()); got != 0 {
		t.Fatalf("diagnostic count = %d, want 0", got)
	}
}

func TestClientManager_Get_ReturnsErrorWhenClientDoesNotExist(t *testing.T) {
	cm := New()
	diags := newTestDiagnostics()

	got := cm.Get("missing", diags)

	if got != nil {
		t.Fatalf("Get returned %#v, want nil", got)
	}

	if got := len(diags.Snapshot()); got != 1 {
		t.Fatalf("diagnostic count = %d, want 1", got)
	}

	if !diags.AnySeverity(diagnostic.SEVERITY_Error) {
		t.Fatal("expected an error diagnostic")
	}

	if diags.AnySeverity(diagnostic.SEVERITY_Warning) {
		t.Fatal("did not expect a warning diagnostic")
	}
}

func TestClientManager_TryGet_ReturnsExistingClient(t *testing.T) {
	cm := New()
	client := newTestClient()
	diags := newTestDiagnostics()

	cm.set("primary", client)

	got := cm.TryGet("primary", diags)

	if got != client {
		t.Fatal("TryGet returned the wrong client")
	}

	if got := len(diags.Snapshot()); got != 0 {
		t.Fatalf("diagnostic count = %d, want 0", got)
	}
}

func TestClientManager_TryGet_ReturnsWarningWhenClientDoesNotExist(t *testing.T) {
	cm := New()
	diags := newTestDiagnostics()

	got := cm.TryGet("missing", diags)

	if got != nil {
		t.Fatalf("TryGet returned %#v, want nil", got)
	}

	if got := len(diags.Snapshot()); got != 1 {
		t.Fatalf("diagnostic count = %d, want 1", got)
	}

	if !diags.AnySeverity(diagnostic.SEVERITY_Warning) {
		t.Fatal("expected a warning diagnostic")
	}

	if diags.AnySeverity(diagnostic.SEVERITY_Error) {
		t.Fatal("did not expect an error diagnostic")
	}
}

func TestClientManager_Add_AddsValidClient(t *testing.T) {
	cm := New()
	client := newTestClient()
	diags := newTestDiagnostics()

	if ok := cm.Add("primary", client, diags); !ok {
		t.Fatal("Add returned false")
	}

	if got := cm.get("primary"); got != client {
		t.Fatal("Add did not store the client")
	}

	if got := len(diags.Snapshot()); got != 0 {
		t.Fatalf("diagnostic count = %d, want 0", got)
	}
}

func TestClientManager_Add_RejectsNilClient(t *testing.T) {
	cm := New()
	diags := newTestDiagnostics()

	if ok := cm.Add("primary", nil, diags); ok {
		t.Fatal("Add accepted a nil client")
	}

	if got := cm.get("primary"); got != nil {
		t.Fatalf("Add stored a nil client: %#v", got)
	}

	if got := len(diags.Snapshot()); got != 1 {
		t.Fatalf("diagnostic count = %d, want 1", got)
	}

	if !diags.AnySeverity(diagnostic.SEVERITY_Error) {
		t.Fatal("expected an error diagnostic")
	}
}

func TestClientManager_Add_RejectsDuplicateName(t *testing.T) {
	cm := New()
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

	if got := cm.get("primary"); got != first {
		t.Fatal("duplicate Add replaced the original client")
	}

	if got := len(duplicateDiags.Snapshot()); got != 1 {
		t.Fatalf("diagnostic count = %d, want 1", got)
	}

	if !duplicateDiags.AnySeverity(diagnostic.SEVERITY_Error) {
		t.Fatal("expected an error diagnostic")
	}
}

func TestClientManager_Set_StoresValidClient(t *testing.T) {
	cm := New()
	client := newTestClient()
	diags := newTestDiagnostics()

	cm.Set("primary", client, diags)

	if got := cm.get("primary"); got != client {
		t.Fatal("Set did not store the client")
	}

	if got := len(diags.Snapshot()); got != 0 {
		t.Fatalf("diagnostic count = %d, want 0", got)
	}
}

func TestClientManager_Set_OverwritesExistingClient(t *testing.T) {
	cm := New()
	first := newTestClient()
	second := newTestClient()

	cm.clients["primary"] = first
	diags := newTestDiagnostics()

	cm.Set("primary", second, diags)

	if got := cm.get("primary"); got != second {
		t.Fatal("Set did not overwrite the existing client")
	}

	if got := len(diags.Snapshot()); got != 0 {
		t.Fatalf("diagnostic count = %d, want 0", got)
	}
}

func TestClientManager_Set_RejectsNilClient(t *testing.T) {
	cm := New()
	diags := newTestDiagnostics()

	cm.Set("primary", nil, diags)

	if got := cm.get("primary"); got != nil {
		t.Fatalf("Set stored a nil client: %#v", got)
	}

	if got := len(diags.Snapshot()); got != 1 {
		t.Fatalf("diagnostic count = %d, want 1", got)
	}

	if !diags.AnySeverity(diagnostic.SEVERITY_Error) {
		t.Fatal("expected an error diagnostic")
	}
}

func TestClientManager_ZeroValue_IsUsable(t *testing.T) {
	var cm ClientManager
	client := newTestClient()
	diags := newTestDiagnostics()

	if ok := cm.Add("primary", client, diags); !ok {
		t.Fatal("Add returned false for zero-value ClientManager")
	}

	if got := cm.Get("primary", diags); got != client {
		t.Fatal("Get returned the wrong client")
	}

	if got := len(diags.Snapshot()); got != 0 {
		t.Fatalf("diagnostic count = %d, want 0", got)
	}
}

func TestClientManager_StoresSupportedNames(t *testing.T) {
	cm := New()
	client := newTestClient()

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
			if ok := cm.add(name, client); !ok {
				t.Fatal("add returned false")
			}

			if got := cm.get(name); got != client {
				t.Fatal("get returned the wrong client")
			}
		})
	}
}

func TestClientManager_PublicMethods_AreSafeForConcurrentUse(t *testing.T) {
	var cm ClientManager

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
					cm.set(name, client)
				case 1:
					cm.add(name, client)
				case 2:
					_ = cm.get(name)
				default:
					diags := newTestDiagnostics()
					_ = cm.Get(name, diags)
				}
			}
		}(worker)
	}

	wg.Wait()
}
