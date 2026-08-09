package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &nodeNetworkDataSource{}
var _ datasource.DataSourceWithConfigure = &nodeNetworkDataSource{}

func NewNodeNetworkDataSource() datasource.DataSource {
	return &nodeNetworkDataSource{}
}

type nodeNetworkDataSource struct {
	clients ProviderClientManager
}

func (d *nodeNetworkDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_network"
}

func (d *nodeNetworkDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about network interfaces, bridges, bonds, and links configured on a specific Proxmox node.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"node": schema.StringAttribute{
				MarkdownDescription: "The name of the target Proxmox node to fetch network links from.",
				Required:            true, // We must know which host to probe
			},
			"networks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The interface device name (e.g. vmbr0, eth0).",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The interface configuration type (e.g. bridge, eth, bond).",
							Computed:            true,
						},
						"active": schema.BoolAttribute{
							MarkdownDescription: "Indicates whether the network interface is currently operational.",
							Computed:            true,
						},
						"cidr": schema.StringAttribute{
							MarkdownDescription: "The IPv4 or IPv6 network address space assigned to the link.",
							Computed:            true,
						},
						"gateway": schema.StringAttribute{
							MarkdownDescription: "The default gateway routed through this interface.",
							Computed:            true,
						},
						"autostart": schema.BoolAttribute{
							MarkdownDescription: "Whether the interface starts automatically on system boot.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *nodeNetworkDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *nodeNetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state nodeNetworkDataSourceModel

	// Read input configuration from the plan
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetNode := state.Node.ValueString()

	// 1. Locate an active Proxmox client connection matching the requested host
	pveClient, exists := d.clients[targetNode]
	if !exists || pveClient == nil {
		// Fallback lookup: if names don't match perfectly, grab the first available connection client
		for _, client := range d.clients {
			if client != nil {
				pveClient = client
				break
			}
		}
	}

	if pveClient == nil {
		resp.Diagnostics.AddError(
			"No Proxmox Client Connection Available",
			fmt.Sprintf("Could not establish a connection manager pool entry to check node: %s", targetNode),
		)
		return
	}

	// 2. Fetch the concrete Node handle from the API
	node, err := pveClient.Node(ctx, targetNode)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Locate Proxmox Node",
			fmt.Sprintf("The node '%s' could not be resolved or fetched: %v", targetNode, err),
		)
		return
	}

	// 3. Query the networks endpoint on that specific hardware node
	networks, err := node.Networks(ctx) // Leverages go-proxmox's /nodes/{node}/network bindings
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Fetch Node Network Interface Metrics",
			fmt.Sprintf("API communication with node '%s' dropped: %v", targetNode, err),
		)
		return
	}

	// 4. Map the slices into state outputs
	state.ID = types.StringValue(fmt.Sprintf("pve-net-%s-%d", targetNode, time.Now().Unix()))
	state.Networks = make([]nodeNetworkItemModel, 0, len(networks))

	for _, netLink := range networks {
		if netLink == nil {
			continue
		}

		state.Networks = append(state.Networks, nodeNetworkItemModel{
			Name:      types.StringValue(netLink.Iface),
			Type:      types.StringValue(netLink.Type),
			Active:    types.BoolValue(netLink.Active == 1), // maps int 1/0 flags to boolean types
			CIDR:      types.StringValue(netLink.CIDR),
			Gateway:   types.StringValue(netLink.Gateway),
			Autostart: types.BoolValue(netLink.Autostart == 1),
		})
	}

	// Save final values straight into the state file
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
