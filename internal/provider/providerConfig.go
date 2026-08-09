package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	proxmox "github.com/luthermonson/go-proxmox"
)

func resolveNodePort(nodePort basetypes.Int32Value, globalPort basetypes.Int32Value) int {
	if !nodePort.IsNull() && !nodePort.IsUnknown() {
		return int(nodePort.ValueInt32())
	}
	if !globalPort.IsNull() && !globalPort.IsUnknown() {
		return int(globalPort.ValueInt32())
	}
	return 8006 // Default Proxmox VE API Port fallback
}

func resolveNodeIgnoreSSL(nodeSSL basetypes.BoolValue, globalSSL basetypes.BoolValue) bool {
	if !nodeSSL.IsNull() && !nodeSSL.IsUnknown() {
		return nodeSSL.ValueBool()
	}
	if !globalSSL.IsNull() && !globalSSL.IsUnknown() {
		return globalSSL.ValueBool()
	}
	return false // Secure by default
}

func resolveNodeCredential(ctx context.Context, nodeCred basetypes.ObjectValue, globalCred basetypes.ObjectValue) (credentialModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var chosenCred credentialModel

	// If the individual node has credentials defined, unpack and use them
	if !nodeCred.IsNull() && !nodeCred.IsUnknown() {
		diags.Append(nodeCred.As(ctx, &chosenCred, basetypes.ObjectAsOptions{})...)
		return chosenCred, diags
	}

	// Fallback to global credentials if validly initialized
	if !globalCred.IsNull() && !globalCred.IsUnknown() {
		diags.Append(globalCred.As(ctx, &chosenCred, basetypes.ObjectAsOptions{})...)
		return chosenCred, diags
	}

	// Neither local nor global credentials were found
	diags.AddError(
		"Missing Required Configuration",
		"Credentials must be specified. Define the 'credential' block globally at the provider root, or explicitly inside the 'nodes' configuration list.",
	)
	return chosenCred, diags
}

func resolveNodeConfig(ctx context.Context, node nodeModel, config pveProviderModel) (ConfiguredNode, diag.Diagnostics) {
	var diags diag.Diagnostics

	port := resolveNodePort(node.Port, config.Port)
	ignoreSSL := resolveNodeIgnoreSSL(node.IgnoreSSL, config.IgnoreSSL)
	cred, diags := resolveNodeCredential(ctx, node.Credential, config.Credential)
	if diags.HasError() {
		return ConfiguredNode{}, diags
	}

	return ConfiguredNode{
		Name:        node.Name.ValueString(),
		Target:      node.Target.ValueString(),
		Port:        port,
		IgnoreSSL:   ignoreSSL,
		Username:    cred.Username.ValueString(),
		Password:    cred.Password.ValueString(),
		Otp:         cred.Otp.ValueString(),
		TokenName:   cred.TokenName.ValueString(),
		TokenSecret: cred.TokenSecret.ValueString(),
	}, diags
}

type ConfiguredNode struct {
	Name        string
	Target      string
	Port        int
	IgnoreSSL   bool
	Username    string
	Password    string
	Otp         string
	TokenName   string
	TokenSecret string
}

func (cn ConfiguredNode) GetConnectionString() string {
	return fmt.Sprintf("https://%s:%d/api2/json", cn.Target, cn.Port)
}

func (cn ConfiguredNode) GetClientCredentialOpt() proxmox.Option {
	if cn.TokenName != "" && cn.TokenSecret != "" {
		tokenID := fmt.Sprintf("%s!%s", cn.Username, cn.TokenName)
		return proxmox.WithAPIToken(tokenID, cn.TokenSecret)
	} else {
		creds := proxmox.Credentials{
			Username: cn.Username,
			Password: cn.Password,
		}
		return proxmox.WithCredentials(&creds)
	}
}
func (cn ConfiguredNode) NewClient() (*proxmox.Client, error) {
	opts := make([]proxmox.Option, 0, 3)

	opts = append(opts, cn.GetClientCredentialOpt())
	if cn.IgnoreSSL {
		opts = append(opts, proxmox.WithInsecureSkipVerify())
	}
	opts = append(opts, proxmox.WithTimeout(30*time.Second))

	nodeUrl := cn.GetConnectionString()
	client := proxmox.NewClient(nodeUrl, opts...)

	return client, nil
}

func testClientConnection(ctx context.Context, client *proxmox.Client) error {
	connCtx, connCancel := context.WithTimeout(ctx, 5*time.Second)
	defer connCancel()

	_, err := client.Version(connCtx)
	return err
}
