package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

type providerValidateCheck func(context.Context, provider.ValidateConfigRequest, *provider.ValidateConfigResponse)
type providerValidateSet []providerValidateCheck
type providerValidator []providerValidateSet

func NewProvidereValidator() *providerValidator {
	nodeChecks := providerValidateSet{
		validateNodeAttributeFallback,
		validateNoDuplicateNodeNames,
		validateNoDuplicateNodeTargetValues,
	}
	return &providerValidator{
		nodeChecks,
	}
}
func (v *providerValidator) Description(_ context.Context) string {
	return "Validates provider configuration."
}
func (v *providerValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (v *providerValidator) ValidateProvider(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	for _, set := range *v {
		for _, check := range set {
			check(ctx, req, resp)
		}
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func validateNodeAttributeFallback(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	var diags diag.Diagnostics

	var config pveProviderModel
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var nodes []nodeModel
	diags = config.Nodes.ElementsAs(ctx, &nodes, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for idx, node := range nodes {
		nodePath := path.Root("nodes").AtListIndex(idx)

		if config.Credential.IsNull() && node.Credential.IsNull() {
			resp.Diagnostics.AddAttributeError(
				nodePath.AtName("credential"),
				"Invalid Configuration",
				"The 'credential' block must be set in either the root provider configuration or explicitly on each node definition.",
			)
		}
	}
}

func validateNoDuplicateNodeNames(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	var diags diag.Diagnostics

	var nodes []nodeModel
	diags = req.Config.GetAttribute(ctx, path.Root("nodes"), &nodes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	seenNodes := make(map[string]int) // map with key being the node name and value being the index of the first node seen with the node name.
	for i, node := range nodes {
		nodeName := strings.TrimSpace(strings.ToLower(node.Name.ValueString()))
		nodePath := path.Root("nodes").AtListIndex(i)

		if firstIndex, duplicate := seenNodes[nodeName]; duplicate {
			resp.Diagnostics.AddAttributeError(
				nodePath.AtName("name"),
				"Duplicate Node",
				fmt.Sprintf("The node named '%s' is defined multiple times (found at index %d and %d).", nodeName, firstIndex, i),
			)
			continue
		} else {
			seenNodes[nodeName] = i
		}
	}
}

func validateNoDuplicateNodeTargetValues(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	var diags diag.Diagnostics

	var nodes []nodeModel
	diags = req.Config.GetAttribute(ctx, path.Root("nodes"), &nodes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	seenNodeTargets := make(map[string]int, len(nodes)) // key is target value. value is first index
	for i, node := range nodes {
		nodeName := strings.TrimSpace(strings.ToLower(node.Name.ValueString()))
		nodeTarget := strings.TrimSpace(strings.ToLower(node.Target.ValueString()))
		nodePath := path.Root("nodes").AtListIndex(i)
		if firstIndex, duplicate := seenNodeTargets[nodeTarget]; duplicate {
			resp.Diagnostics.AddAttributeError(
				nodePath.AtName("target"),
				"Duplicate Node Target",
				fmt.Sprintf("The target '%s' on node named '%s' at index %d has been defined multiple times (first seen at index %d).", nodeTarget, nodeName, i, firstIndex),
			)
			continue
		} else {
			seenNodeTargets[nodeTarget] = i
		}
	}
}
