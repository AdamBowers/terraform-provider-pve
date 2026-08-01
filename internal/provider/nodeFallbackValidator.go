package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type nodeFallbackValidator struct{}

func NewNodeFallbackValidator() *nodeFallbackValidator {
	return &nodeFallbackValidator{}
}

func (v *nodeFallbackValidator) Description(_ context.Context) string {
	return "Ensures nested node configuration falls back to root settings when missing."
}

func (v *nodeFallbackValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *nodeFallbackValidator) ValidateProvider(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	validateCredentialFallback(ctx, req, resp)
}

func validateCredentialFallback(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	var diags diag.Diagnostics

	var rootCredential types.Object
	diags = req.Config.GetAttribute(ctx, path.Root("credential"), &rootCredential)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var nodes []nodeModel
	diags = req.Config.GetAttribute(ctx, path.Root("nodes"), &nodes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for i, node := range nodes {
		nodePath := path.Root("nodes").AtListIndex(i)

		if rootCredential.IsNull() && node.Credential.IsNull() {
			resp.Diagnostics.AddAttributeError(
				nodePath.AtName("credential"),
				"Invalid Configuration",
				"The 'credential' block must be set in either the root provider configuration or explicitly on each node definition.",
			)
		}
	}
}
