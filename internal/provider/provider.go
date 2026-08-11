package provider

import (
	"context"
	"fmt"
	"regexp"
	"sync"

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
				stringvalidator.ConflictsWith(
					path.MatchRelative().AtParent().AtName("token_name"),
					path.MatchRelative().AtParent().AtName("token_secret"),
				),
			},
		},
		"otp": schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "OTP to be used for all nodes unless defined otherwise.",
			Validators: []validator.String{
				stringvalidator.ConflictsWith(
					path.MatchRelative().AtParent().AtName("token_name"),
					path.MatchRelative().AtParent().AtName("token_secret"),
				),
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
				stringvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("password"),
				),
				stringvalidator.ConflictsWith(
					path.MatchRelative().AtParent().AtName("password"),
					path.MatchRelative().AtParent().AtName("otp"),
				),
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
				stringvalidator.ConflictsWith(
					path.MatchRelative().AtParent().AtName("password"),
					path.MatchRelative().AtParent().AtName("otp"),
				),
				stringvalidator.AlsoRequires(
					path.MatchRelative().AtParent().AtName("token_name"),
				),
			},
		},
	}
	schemaNode = map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "name of proxmox node.",
		},
		"target": schema.StringAttribute{
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
	resp.TypeName = "pve"
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
	var diags diag.Diagnostics

	var config pveProviderModel
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
	if !config.Credential.IsNull() && !config.Credential.IsUnknown() {
		diags = config.Credential.As(ctx, &rootCred, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	manager := make(ProviderClientManager)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, node := range nodes {
		nodePath := path.Root("nodes").AtListIndex(i)
		nodeConf, diags := resolveNodeConfig(ctx, node, config)
		if diags.HasError() {
			for _, d := range diags {
				resp.Diagnostics.AddAttributeError(nodePath, d.Summary(), d.Detail())
			}
			continue
		}

		client, err := nodeConf.NewClient()
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				nodePath,
				"Client Initialization Failed",
				fmt.Sprintf("Failed to construct the Proxmox client instance for '%s': %s", nodeConf.Name, err),
			)
			continue
		}
		manager[nodeConf.Name] = client

		wg.Add(1)
		go func(c *proxmox.Client, nc ConfiguredNode, np path.Path) {
			defer wg.Done()

			testErr := testClientConnection(ctx, c)
			if testErr != nil {
				mu.Lock()
				resp.Diagnostics.AddAttributeError(
					np,
					"Proxmox Connection Failed",
					fmt.Sprintf("Successfully initialized client, but failed to connect to node '%s' (%s): %s", nc.Name, nc.GetConnectionString(), testErr),
				)
				mu.Unlock()
				return
			}
		}(client, nodeConf, nodePath)
	}

	wg.Wait()

	if resp.Diagnostics.HasError() {
		return
	}

	resp.DataSourceData = manager
	resp.ResourceData = manager
}
func (p *pveProvider) ConfigValidators(ctx context.Context) []provider.ConfigValidator {
	return []provider.ConfigValidator{
		NewProviderValidator(),
	}
}
func (p *pveProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewNodeNetworkDataSource,
		NewNodeVersionDataSource,
		NewUsersDataSource,
	}
}
func (p *pveProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// NewUserResource,
	}
}
