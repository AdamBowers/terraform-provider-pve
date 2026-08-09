package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type nodeNetworkDataSourceModel struct {
	ID       types.String           `tfsdk:"id"`
	Node     types.String           `tfsdk:"node"`     // INPUT: Target host to inspect
	Networks []nodeNetworkItemModel `tfsdk:"networks"` // OUTPUT
}

type nodeNetworkItemModel struct {
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`    // bridge, eth, bond, vlan, etc.
	Active    types.Bool   `tfsdk:"active"`  // Is the link live?
	CIDR      types.String `tfsdk:"cidr"`    // Assigned IP subnet address if applicable
	Gateway   types.String `tfsdk:"gateway"` // Assigned gateway if applicable
	Autostart types.Bool   `tfsdk:"autostart"`
}
