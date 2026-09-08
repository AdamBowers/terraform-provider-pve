package clientmanager

import (
	"testing"

	"github.com/AdamBowers/terraform-provider-pve/v2/internal/diagnostic"
	"github.com/luthermonson/go-proxmox"
)

func benchmarkValidatorClient() *proxmox.Client {
	return &proxmox.Client{}
}

func BenchmarkValidateClient_NoChecks(b *testing.B) {
	client := benchmarkValidatorClient()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		diags := validateClient(client)

		if diags == nil {
			b.Fatal("validateClient returned nil")
		}
	}
}

func BenchmarkValidateClient_WithCheck(b *testing.B) {
	client := benchmarkValidatorClient()
	check := func(_ *proxmox.Client, _ *diagnostic.Diagnostics) {}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		diags := validateClient(client, check)

		if diags == nil {
			b.Fatal("validateClient returned nil")
		}
	}
}

func BenchmarkDefaultClientValidator_ValidClient(b *testing.B) {
	client := benchmarkValidatorClient()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		diags := defaultClientValidator(client)

		if diags == nil {
			b.Fatal("defaultClientValidator returned nil")
		}
	}
}

func BenchmarkDefaultClientValidator_NilClient(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		diags := defaultClientValidator(nil)

		if diags == nil {
			b.Fatal("defaultClientValidator returned nil")
		}
	}
}

func BenchmarkCheckClientNotNil_ValidClient(b *testing.B) {
	client := benchmarkValidatorClient()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		diags := diagnostic.New()
		checkClientNotNil(client, &diags)
	}
}

func BenchmarkCheckClientNotNil_NilClient(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		diags := diagnostic.New()
		checkClientNotNil(nil, &diags)
	}
}
