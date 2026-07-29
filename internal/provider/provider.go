package provider

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_              provider.Provider = &pveProvider{}
	rgxUsername                      = regexp.MustCompile(`^[A-Za-z0-9_\.\-]+@[A-Za-z0-9_\.\-]+$`)
	rgxTokenName                     = regexp.MustCompile(`^[A-Za-z0-9_\.\-]+$`)
	rgxTokenSecret                   = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &pveProvider{
			version: version,
		}
	}
}

type pveProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *pveProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "proxmox-tf"
	resp.Version = p.version
}
func (p *pveProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:    true,
				Description: "FQDN of proxmox node to connect to for datacentre management.",
			},
			"port": schema.Int32Attribute{
				Optional:    true,
				Description: "Port to connect to node on. Default is 8006.",
				Validators: []validator.Int32{
					int32validator.AtMost(65535),
					int32validator.AtLeast(1),
				},
			},
			"insecure": schema.BoolAttribute{
				Optional:    true,
				Description: "Connect over HTTP instead of HTTPS. Default is False.",
			},
			"ignore_ssl": schema.BoolAttribute{
				Optional:    true,
				Description: "Don't validate SSL certificate. Default is False.",
			},
		},
		Blocks: map[string]schema.Block{
			"credential": schema.SingleNestedBlock{
				Description: "Credentials to used to connect to proxmox virtual environment node(s).",
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(
						path.MatchRoot("credential").AtName("basic"),
						path.MatchRoot("credential").AtName("token"),
					),
				},
				Blocks: map[string]schema.Block{
					"basic": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Required: true,
								Description: "Default username to be used for all nodes unless defined otherwise. " +
									"Format as '<userid>@<realm>'.",
								Validators: []validator.String{
									stringvalidator.LengthAtLeast(3),
									stringvalidator.RegexMatches(
										rgxUsername,
										"username provided is an invalid pattern",
									),
								},
							},
							"password": schema.StringAttribute{
								Required:    true,
								Sensitive:   true,
								Description: "Default password to be used for all nodes unless defined otherwise.",
							},
							"otp": schema.StringAttribute{
								Optional:    true,
								Sensitive:   true,
								Description: "OTP to be used for all nodes unless defined otherwise.",
							},
						},
					},
					"token": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Required: true,
								Description: "username API token is associated with. " +
									"Format as '<userid>@<realm>'.",
								Validators: []validator.String{
									stringvalidator.LengthAtLeast(3),
									stringvalidator.RegexMatches(
										rgxUsername,
										"username provided is an invalid pattern",
									),
								},
							},
							"name": schema.StringAttribute{
								Required:    true,
								Description: "API Token Name/ID.",
								Validators: []validator.String{
									stringvalidator.LengthBetween(2, 64),
									stringvalidator.RegexMatches(
										rgxTokenName,
										"name provided is an invalid pattern",
									),
								},
							},
							"secret": schema.StringAttribute{
								Required:    true,
								Sensitive:   true,
								Description: "API token secret.",
								Validators: []validator.String{
									stringvalidator.LengthBetween(36, 36),
									stringvalidator.RegexMatches(
										rgxTokenSecret,
										"secret provided is an invalid pattern",
									),
								},
							},
						},
					},
				},
			},
		},
	}
}
func (p *pveProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}
func (p *pveProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
func (p *pveProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
