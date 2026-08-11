package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type nodeVersionDataSourceModel struct {
	Node    types.String `tfsdk:"node"`
	Release types.String `tfsdk:"release"`
	Repoid  types.String `tfsdk:"repoid"`
	Version types.String `tfsdk:"version"`
}
