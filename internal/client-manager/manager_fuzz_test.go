package clientmanager

import (
	"testing"

	"github.com/AdamBowers/terraform-provider-pve/v2/internal/diagnostic"
	"github.com/luthermonson/go-proxmox"
)

func boundedName(name string) string {
	const maxNameLength = 256

	if len(name) > maxNameLength {
		return name[:maxNameLength]
	}

	return name
}

func FuzzClientManagerAddAndGet(f *testing.F) {
	for _, seed := range []string{
		"",
		"client",
		" ",
		"\t\n",
		"client/name",
		`client"name`,
		"客户端",
		"😀",
		"\x00",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, name string) {
		name = boundedName(name)

		cm := ClientManager{}
		client := new(proxmox.Client)

		var diags diagnostic.Diagnostics
		if !cm.Add(name, client, &diags) {
			t.Fatalf("Add(%q) returned false: %v", name, diags.Snapshot())
		}

		if diags.Count() != 0 {
			t.Fatalf("Add(%q) produced diagnostics: %v", name, diags.Snapshot())
		}

		var getDiags diagnostic.Diagnostics
		got := cm.Get(name, &getDiags)

		if got != client {
			t.Fatalf("Get(%q) returned %p, want %p", name, got, client)
		}

		if getDiags.Count() != 0 {
			t.Fatalf("Get(%q) produced diagnostics: %v", name, getDiags.Snapshot())
		}
	})
}

func FuzzClientManagerDuplicateAdd(f *testing.F) {
	for _, seed := range []string{
		"",
		"client",
		"client/name",
		"同じ名前",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, name string) {
		name = boundedName(name)

		cm := ClientManager{}
		first := new(proxmox.Client)
		second := new(proxmox.Client)

		var firstDiags diagnostic.Diagnostics
		if !cm.Add(name, first, &firstDiags) {
			t.Fatalf("initial Add(%q) returned false: %v", name, firstDiags.Snapshot())
		}

		var duplicateDiags diagnostic.Diagnostics
		if cm.Add(name, second, &duplicateDiags) {
			t.Fatalf("duplicate Add(%q) returned true", name)
		}

		if duplicateDiags.Count() == 0 {
			t.Fatalf("duplicate Add(%q) produced no diagnostics", name)
		}

		var getDiags diagnostic.Diagnostics
		got := cm.Get(name, &getDiags)

		if got != first {
			t.Fatalf("duplicate Add replaced client: got %p, want %p", got, first)
		}
	})
}

func FuzzClientManagerSet(f *testing.F) {
	for _, seed := range []string{
		"",
		"client",
		" ",
		"client/name",
		"客户端",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, name string) {
		name = boundedName(name)

		cm := ClientManager{}
		first := new(proxmox.Client)
		second := new(proxmox.Client)

		var diags diagnostic.Diagnostics
		cm.Set(name, first, &diags)
		cm.Set(name, second, &diags)

		if diags.Count() != 0 {
			t.Fatalf("Set(%q) produced diagnostics: %v", name, diags.Snapshot())
		}

		var getDiags diagnostic.Diagnostics
		got := cm.Get(name, &getDiags)

		if got != second {
			t.Fatalf("Set(%q) did not overwrite client: got %p, want %p",
				name, got, second)
		}
	})
}

func FuzzClientManagerMissingLookup(f *testing.F) {
	f.Add("missing")
	f.Add("")
	f.Add(" ")
	f.Add("client/name")
	f.Add("不存在")

	f.Fuzz(func(t *testing.T, name string) {
		name = boundedName(name)

		cm := ClientManager{}

		var getDiags diagnostic.Diagnostics
		if got := cm.Get(name, &getDiags); got != nil {
			t.Fatalf("Get(%q) returned %p, want nil", name, got)
		}

		if getDiags.Count() == 0 {
			t.Fatalf("Get(%q) produced no missing-client diagnostic", name)
		}

		var tryDiags diagnostic.Diagnostics
		if got := cm.TryGet(name, &tryDiags); got != nil {
			t.Fatalf("TryGet(%q) returned %p, want nil", name, got)
		}

		if tryDiags.Count() == 0 {
			t.Fatalf("TryGet(%q) produced no missing-client diagnostic", name)
		}
	})
}
