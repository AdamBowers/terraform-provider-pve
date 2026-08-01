package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type pveProviderModel struct {
	Nodes      types.List   `tfsdk:"nodes"`
	Port       types.Int32  `tfsdk:"port"`
	IgnoreSSL  types.Bool   `tfsdk:"ignore_ssl"`
	Credential types.Object `tfsdk:"credential"`
}

type nodeModel struct {
	Host       types.String `tfsdk:"host"`
	Port       types.Int32  `tfsdk:"port"`
	IgnoreSSL  types.Bool   `tfsdk:"ignore_ssl"`
	Credential types.Object `tfsdk:"credential"`
}

type credentialModel struct {
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	Otp         types.String `tfsdk:"otp"`
	TokenName   types.String `tfsdk:"token_name"`
	TokenSecret types.String `tfsdk:"token_secret"`
}
