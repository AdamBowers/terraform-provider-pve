package clientmanager

import (
	"github.com/AdamBowers/terraform-provider-pve/v2/internal/diagnostic"
	"github.com/luthermonson/go-proxmox"
)

type clientValidatorCheck func(*proxmox.Client, *diagnostic.Diagnostics)

func validateClient(client *proxmox.Client, checks ...clientValidatorCheck) *diagnostic.Diagnostics {
	diags := diagnostic.New()

	for _, check := range checks {
		check(client, &diags)
	}

	return &diags
}

func defaultClientValidator(client *proxmox.Client) *diagnostic.Diagnostics {
	return validateClient(
		client,
		checkClientNotNil,
	)
}

func checkClientNotNil(client *proxmox.Client, diags *diagnostic.Diagnostics) {
	if client != nil {
		return
	}

	diags.Append(
		diagnostic.Error(
			"Nil Value Reference",
			"provided proxmox client is nil",
			diagnostic.WithoutStacktrace(),
		),
	)
}
