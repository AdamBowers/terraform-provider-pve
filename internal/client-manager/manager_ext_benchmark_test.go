package clientmanager_test

import (
	"strconv"
	"testing"

	clientmanager "github.com/AdamBowers/terraform-provider-pve/v2/internal/client-manager"
	"github.com/AdamBowers/terraform-provider-pve/v2/internal/diagnostic"
	"github.com/luthermonson/go-proxmox"
)

func benchmarkDiagnostics() *diagnostic.Diagnostics {
	diags := diagnostic.New()
	return &diags
}

func BenchmarkClientManager_New(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = clientmanager.New()
	}
}

func BenchmarkClientManager_Add(b *testing.B) {
	cm := clientmanager.New()
	client := &proxmox.Client{}
	diags := benchmarkDiagnostics()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		name := "client-" + strconv.Itoa(i)

		if ok := cm.Add(name, client, diags); !ok {
			b.Fatalf("Add returned false for name %q", name)
		}
	}
}

func BenchmarkClientManager_Add_Duplicate(b *testing.B) {
	cm := clientmanager.New()
	client := &proxmox.Client{}
	diags := benchmarkDiagnostics()

	if ok := cm.Add("primary", client, diags); !ok {
		b.Fatal("initial Add returned false")
	}

	diags = benchmarkDiagnostics()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if ok := cm.Add("primary", client, diags); ok {
			b.Fatal("Add returned true for duplicate name")
		}

		// Add appends an error diagnostic for every duplicate. Reset the
		// diagnostics so the benchmark does not measure unbounded growth.
		diags = benchmarkDiagnostics()
	}
}

func BenchmarkClientManager_Set(b *testing.B) {
	cm := clientmanager.New()
	client := &proxmox.Client{}
	diags := benchmarkDiagnostics()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cm.Set("primary", client, diags)
	}
}

func BenchmarkClientManager_Get(b *testing.B) {
	cm := clientmanager.New()
	client := &proxmox.Client{}
	setupDiags := benchmarkDiagnostics()

	if ok := cm.Add("primary", client, setupDiags); !ok {
		b.Fatal("Add returned false")
	}

	diags := benchmarkDiagnostics()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if got := cm.Get("primary", diags); got != client {
			b.Fatal("Get returned a different client")
		}
	}
}

func BenchmarkClientManager_TryGet(b *testing.B) {
	cm := clientmanager.New()
	client := &proxmox.Client{}
	setupDiags := benchmarkDiagnostics()

	if ok := cm.Add("primary", client, setupDiags); !ok {
		b.Fatal("Add returned false")
	}

	diags := benchmarkDiagnostics()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if got := cm.TryGet("primary", diags); got != client {
			b.Fatal("TryGet returned a different client")
		}
	}
}

func BenchmarkClientManager_Get_Missing(b *testing.B) {
	cm := clientmanager.New()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		diags := benchmarkDiagnostics()

		if got := cm.Get("missing", diags); got != nil {
			b.Fatal("Get returned a client for a missing name")
		}
	}
}

func BenchmarkClientManager_TryGet_Missing(b *testing.B) {
	cm := clientmanager.New()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		diags := benchmarkDiagnostics()

		if got := cm.TryGet("missing", diags); got != nil {
			b.Fatal("TryGet returned a client for a missing name")
		}
	}
}
