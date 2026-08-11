package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type nodeNetworkDataSourceModel struct {
	Node     types.String            `tfsdk:"node"`     // INPUT: Target host to inspect
	Networks []nodeNetworkIfaceModel `tfsdk:"networks"` // OUTPUT
}

type nodeNetworkIfaceModel struct {
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"` // bridge | bond | eth | alias | vlan | fabric | OVSBridge | OVSBond | OVSPort | OVSIntPort | vnet | unknown
	Active    types.Bool   `tfsdk:"active"`
	Autostart types.Bool   `tfsdk:"autostart"`
	CIDR      types.String `tfsdk:"cidr"`
	Gateway   types.String `tfsdk:"gateway"`
	CIDR6     types.String `tfsdk:"cidr6"`
	Gateway6  types.String `tfsdk:"gateway6"`
	Comments  types.String `tfsdk:"comments"`
	Comments6 types.String `tfsdk:"comments6"`
}
