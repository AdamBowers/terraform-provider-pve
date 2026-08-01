package provider

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	proxmox "github.com/luthermonson/go-proxmox"
)

var (
	_                provider.Provider                     = &pveProvider{}
	_                provider.ProviderWithConfigValidators = &pveProvider{}
	rgxUsername                                            = regexp.MustCompile(`^[A-Za-z0-9_\.\-]+@[A-Za-z0-9_\.\-]+$`)
	rgxTokenName                                           = regexp.MustCompile(`^[A-Za-z0-9_\.\-]+$`)
	rgxTokenSecret                                         = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	schemaCredential                                       = map[string]schema.Attribute{
		"username": schema.StringAttribute{
			Required:    true,
			Description: "Username to be used for all nodes unless defined otherwise. Format as '<userid>@<realm>'.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(3),
				stringvalidator.RegexMatches(rgxUsername, "username provided is an invalid pattern"),
			},
		},
		"password": schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "Default password to be used for all nodes unless defined otherwise.",
			Validators: []validator.String{
				stringvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("token_name"),
				),
			},
		},
		"otp": schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "OTP to be used for all nodes unless defined otherwise.",
			Validators: []validator.String{
				stringvalidator.AlsoRequires(
					path.MatchRelative().AtParent().AtName("password"),
				),
			},
		},
		"token_name": schema.StringAttribute{
			Optional:    true,
			Description: "API Token Name/ID.",
			Validators: []validator.String{
				stringvalidator.LengthBetween(2, 64),
				stringvalidator.RegexMatches(rgxTokenName, "name provided is an invalid pattern"),
				stringvalidator.AlsoRequires(
					path.MatchRelative().AtParent().AtName("token_secret"),
				),
			},
		},
		"token_secret": schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "API token secret.",
			Validators: []validator.String{
				stringvalidator.LengthBetween(36, 36),
				stringvalidator.RegexMatches(rgxTokenSecret, "secret provided is an invalid pattern"),
				stringvalidator.AlsoRequires(
					path.MatchRelative().AtParent().AtName("token_name"),
				),
			},
		},
	}
	schemaNode = map[string]schema.Attribute{
		"host": schema.StringAttribute{
			Required:    true,
			Description: "Hostname or IP of proxmox node.",
		},
		"port": schema.Int32Attribute{
			Optional:    true,
			Description: "Port to connect to node on. Default is 8006.",
			Validators: []validator.Int32{
				int32validator.AtMost(65535),
				int32validator.AtLeast(1),
			},
		},
		"ignore_ssl": schema.BoolAttribute{
			Optional:    true,
			Description: "Don't validate SSL certificate. Default is False.",
		},
		"credential": schema.SingleNestedAttribute{
			Optional:    true,
			Description: "Credentials to used to connect to proxmox virtual environment node(s).",
			Attributes:  schemaCredential,
		},
	}
)

type ProviderClientManager map[string]*proxmox.Client

type ConfiguredNode struct {
	Host        string
	Port        int32
	IgnoreSSL   bool
	Username    string
	Password    string
	Otp         string
	TokenName   string
	TokenSecret string
}

func (cn ConfiguredNode) GetCredentialMethod() proxmox.Option {
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

type pveProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &pveProvider{
			version: version,
		}
	}
}

func (p *pveProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "proxmox-tf"
	resp.Version = p.version
}
func (p *pveProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"nodes": schema.ListNestedAttribute{
				Required:    true,
				Description: "List of proxmox nodes to manage and any overriding configuration.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: schemaNode,
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"port": schema.Int32Attribute{
				Optional:    true,
				Description: "Port to connect to node on. Default is 8006.",
				Validators: []validator.Int32{
					int32validator.AtMost(65535),
					int32validator.AtLeast(1),
				},
			},
			"ignore_ssl": schema.BoolAttribute{
				Optional:    true,
				Description: "Don't validate SSL certificate. Default is False.",
			},
			"credential": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Credentials to used to connect to proxmox virtual environment node(s).",
				Attributes:  schemaCredential,
			},
		},
	}
}
func (p *pveProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config pveProviderModel
	var diags diag.Diagnostics

	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Nodes.IsUnknown() || config.Credential.IsUnknown() {
		return
	}

	var nodes []nodeModel
	diags = config.Nodes.ElementsAs(ctx, &nodes, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var rootCred credentialModel
	var hasRootCred bool
	if !config.Credential.IsNull() && !config.Credential.IsUnknown() {
		diags = config.Credential.As(ctx, &rootCred, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		hasRootCred = true
	} else {
		hasRootCred = false
	}

	configuredNodes := make([]ConfiguredNode, 0, len(nodes))
	for _, node := range nodes {
		port := int32(8006)
		if !node.Port.IsNull() && !node.Port.IsUnknown() {
			port = node.Port.ValueInt32()
		} else if !config.Port.IsNull() && !config.Port.IsUnknown() {
			port = config.Port.ValueInt32()
		}

		ignoreSSL := false
		if !node.IgnoreSSL.IsNull() && !node.IgnoreSSL.IsUnknown() {
			ignoreSSL = node.IgnoreSSL.ValueBool()
		} else if !config.IgnoreSSL.IsNull() && !config.IgnoreSSL.IsUnknown() {
			ignoreSSL = config.IgnoreSSL.ValueBool()
		}

		var nodeCred credentialModel
		if !node.Credential.IsNull() && !node.Credential.IsUnknown() {
			diags = node.Credential.As(ctx, &nodeCred, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else if hasRootCred {
			nodeCred = rootCred
		} else {
			resp.Diagnostics.AddError(
				"Missing Required Configuration",
				"Credentials must be specified. Define the 'credential' block globally at the provider root, or explicitly inside the 'nodes' configuration list.",
			)
			return
		}

		configuredNodes = append(configuredNodes, ConfiguredNode{
			Host:        node.Host.ValueString(),
			Port:        port,
			IgnoreSSL:   ignoreSSL,
			Username:    nodeCred.Username.ValueString(),
			Password:    nodeCred.Password.ValueString(),
			Otp:         nodeCred.Otp.ValueString(),
			TokenName:   nodeCred.TokenName.ValueString(),
			TokenSecret: nodeCred.TokenSecret.ValueString(),
		})

		manager := make(ProviderClientManager)
		for _, confNode := range configuredNodes {
			nodeUrl := fmt.Sprintf("https://%s:%d/api2/json", confNode.Host, confNode.Port)

			opts := make([]proxmox.Option, 0, 3)
			opts = append(opts, confNode.GetCredentialMethod())
			if confNode.IgnoreSSL {
				opts = append(opts, proxmox.WithInsecureSkipVerify())
			}
			opts = append(opts, proxmox.WithTimeout(30*time.Second))

			client := proxmox.NewClient(nodeUrl)

			manager[confNode.Host] = client
		}

		resp.DataSourceData = manager
		resp.ResourceData = manager
	}
}
func (p *pveProvider) ConfigValidators(ctx context.Context) []provider.ConfigValidator {
	return []provider.ConfigValidator{
		NewNodeValidator(),
	}
}
func (p *pveProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
func (p *pveProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
