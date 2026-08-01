package provider

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type nodeValidator struct{}

func NewNodeValidator() *nodeValidator {
	return &nodeValidator{}
}

func (v *nodeValidator) Description(_ context.Context) string {
	return "Ensures nested node configuration falls back to root settings when missing."
}

func (v *nodeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *nodeValidator) ValidateProvider(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	checks := []func(context.Context, provider.ValidateConfigRequest, *provider.ValidateConfigResponse){
		validateNodeCredentialFallback,
		validateNodeResolvable,
		validateNodeDuplicates,
	}
	for _, check := range checks {
		check(ctx, req, resp)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func validateNodeCredentialFallback(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
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

func validateNodeDuplicates(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	var diags diag.Diagnostics

	var nodes []nodeModel
	diags = req.Config.GetAttribute(ctx, path.Root("nodes"), &nodes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var wg sync.WaitGroup
	resolvedIPs := make([][]net.IP, len(nodes))
	seenHosts := make(map[string]int)

	for i, node := range nodes {
		if node.Host.IsNull() || node.Host.IsUnknown() {
			continue
		}

		hostStr := strings.TrimSpace(strings.ToLower(node.Host.ValueString()))

		// Synchronously block string-matching duplicates instantly
		if firstIndex, duplicate := seenHosts[hostStr]; duplicate {
			nodePath := path.Root("nodes").AtListIndex(i).AtName("host")
			resp.Diagnostics.AddAttributeError(
				nodePath,
				"Duplicate Node Target",
				fmt.Sprintf("The host '%s' is defined multiple times (found at index %d and %d).", hostStr, firstIndex, i),
			)
			continue
		}
		seenHosts[hostStr] = i

		// Raw IP entries require no network overhead; record them right now
		if parsedIP := net.ParseIP(hostStr); parsedIP != nil {
			resolvedIPs[i] = []net.IP{parsedIP}
			continue
		}

		// Asynchronously dispatch hostnames to background network threads
		wg.Add(1)
		go func(index int, hostname string) {
			defer wg.Done()

			dnsCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			ips, err := net.DefaultResolver.LookupIP(dnsCtx, "ip", hostname)
			if err != nil {
				return
			}

			resolvedIPs[index] = ips
		}(i, hostStr)
	}
	wg.Wait()

	seenIPs := make(map[string]int)

	for i, node := range nodes {
		if node.Host.IsNull() || node.Host.IsUnknown() {
			continue
		}

		hostStr := strings.TrimSpace(strings.ToLower(node.Host.ValueString()))
		nodePath := path.Root("nodes").AtListIndex(i).AtName("host")

		for _, ip := range resolvedIPs[i] {
			// Convert to string here since slices cannot be used as map keys
			ipStr := ip.String()

			if firstIndex, duplicate := seenIPs[ipStr]; duplicate {
				duplicateHost := strings.TrimSpace(strings.ToLower(nodes[firstIndex].Host.ValueString()))
				resp.Diagnostics.AddAttributeError(
					nodePath,
					"Duplicate Node Networking IP",
					fmt.Sprintf("The node host '%s' at index %d resolves to the IP address '%s', which is already in use by the node host '%s' at index %d.", hostStr, i, ipStr, duplicateHost, firstIndex),
				)
				break
			}
			seenIPs[ipStr] = i
		}
	}
}

func validateNodeResolvable(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	var diags diag.Diagnostics

	var nodes []nodeModel
	diags = req.Config.GetAttribute(ctx, path.Root("nodes"), &nodes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var wg sync.WaitGroup
	lookupErrors := make([]struct {
		index   int
		hostStr string
		err     error
	}, len(nodes))
	for i, node := range nodes {
		if node.Host.IsNull() || node.Host.IsUnknown() {
			continue
		}

		hostStr := strings.TrimSpace(strings.ToLower(node.Host.ValueString()))
		if parsedIP := net.ParseIP(hostStr); parsedIP != nil {
			continue
		}

		wg.Add(1)
		go func(index int, hostname string) {
			defer wg.Done()

			dnsCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			_, err := net.DefaultResolver.LookupIP(dnsCtx, "ip", hostname)
			if err != nil {
				lookupErrors[index].index = index
				lookupErrors[index].hostStr = hostname
				lookupErrors[index].err = err
			}
		}(i, hostStr)
	}
	wg.Wait()

	for _, lookup := range lookupErrors {
		// ADDRESSING FIELDS: Read from the struct fields.
		// If lookup.err is nil, this index didn't encounter a DNS error, so we skip it.
		if lookup.err != nil {
			nodePath := path.Root("nodes").AtListIndex(lookup.index).AtName("host")

			resp.Diagnostics.AddAttributeError(
				nodePath,
				"Node not resolvable",
				fmt.Sprintf("The node host '%s' at index %d failed resolution due to error:\n\t%s",
					lookup.hostStr,
					lookup.index,
					lookup.err.Error(),
				),
			)
		}
	}
}
