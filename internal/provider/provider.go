package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ provider.Provider = &proxmoxProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &proxmoxProvider{
			version: version,
		}
	}
}

type proxmoxProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *proxmoxProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "proxmox-tf"
	resp.Version = p.version
}
func (p *proxmoxProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}
func (p *proxmoxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}
func (p *proxmoxProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
func (p *proxmoxProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
