package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &nodeVersionDataSource{}
var _ datasource.DataSourceWithConfigure = &nodeVersionDataSource{}

type nodeVersionDataSource struct {
	clients ProviderClientManager
}

func NewNodeVersionDataSource() datasource.DataSource {
	return &nodeVersionDataSource{}
}

func (d *nodeVersionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_version"
}
func (d *nodeVersionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information release, version and repoid for a Proxmox Virtual Environment node.",
		Attributes: map[string]schema.Attribute{
			"node": schema.StringAttribute{
				MarkdownDescription: "The name of the target Proxmox Virtual Environment node to fetch network links from.",
				Required:            true, // We must know which host to probe
			},
			"release": schema.StringAttribute{
				MarkdownDescription: "The current Proxmox VE point release in `x.y` format.",
				Computed:            true,
			},
			"repoid": schema.StringAttribute{
				MarkdownDescription: "The short git revision from which this version was build.",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The full pve-manager package version of this node.",
				Computed:            true,
			},
		},
	}

}
func (d *nodeVersionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(ProviderClientManager)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected ProviderClientManager, got: %T.", req.ProviderData),
		)
		return
	}

	d.clients = clients
}
func (d *nodeVersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state nodeVersionDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetNode := state.Node.ValueString()

	pveClient := d.clients[targetNode]
	if pveClient == nil {
		resp.Diagnostics.AddError(
			"No Proxmox Client Connection Available",
			fmt.Sprintf("Could not establish a connection manager pool entry to check node: %s", targetNode),
		)
		return
	}

	version, err := pveClient.Version(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Locate Proxmox Node",
			fmt.Sprintf("The node '%s' could not be resolved or fetched: %v", targetNode, err),
		)
		return
	}

	state.Release = types.StringValue(version.Release)
	state.Repoid = types.StringValue(version.RepoID)
	state.Version = types.StringValue(version.Version)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
