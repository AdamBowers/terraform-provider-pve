package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type nodeVersionDataSourceModel struct {
	ID      types.String      `tfsdk:"id"`
	Node    types.String      `tfsdk:"node"`
	Version *nodeVersionModel `tfsdk:"version"`
}

type nodeVersionModel struct {
	Release types.String `tfsdk:"release"`
	Repoid  types.String `tfsdk:"repoid"`
	Version types.String `tfsdk:"version"`
}
