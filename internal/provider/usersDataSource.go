package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &usersDataSource{}
var _ datasource.DataSourceWithConfigure = &usersDataSource{}

var (
	rgxUserKeys regexp.Regexp = *regexp.MustCompile(`^[0-9a-zA-Z!=]*$`)
)

type usersDataSource struct {
	clients  ProviderClientManager
	typeName string
}

func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	d.typeName = req.ProviderTypeName + "_users"
	resp.TypeName = d.typeName
}
func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information release, version and repoid for a Proxmox Virtual Environment node.",
		Attributes: map[string]schema.Attribute{
			"node": schema.StringAttribute{ // INPUT - data
				Description: "The name of the target Proxmox Virtual Environment node to fetch user list from.",
				Optional:    true,
			},
			"users": schema.SetNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "userid. name+realm.",
							Computed:    true,
						},
						"realm_type": schema.StringAttribute{
							Description: "Realm type of user.",
							Computed:    true,
						},
						"firstname": schema.StringAttribute{
							Description: "Firstname provided for user.",
							Computed:    true,
						},
						"lastname": schema.StringAttribute{
							Description: "Lastname provided for user.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "Email provided for user.",
							Computed:    true,
						},
						"groups": schema.SetAttribute{
							Description: "Firstname provided for user.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"expire": schema.Int64Attribute{
							Description: "Expirary date configured for user in seconds since epoch." +
								"null if no expirary.",
							Computed: true,
						},
						"enable": schema.BoolAttribute{
							Description: "Email provided for user.",
							Computed:    true,
						},
						"tfa_locked_until": schema.StringAttribute{
							Description: "Timestamp when users 2FA was locked. In seconds for epoch.",
							Computed:    true,
						},
						"totp_locked": schema.BoolAttribute{
							Description: "True if the user is currently locked out of TOTP factor.",
							Computed:    true,
						},
						"keys": schema.StringAttribute{
							Description: "Keys for two factor authentication with yubico.",
							Computed:    true,
							Sensitive:   true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(4096),
								stringvalidator.RegexMatches(&rgxUserKeys, fmt.Sprintf("value must match pattern '%s'", rgxUserKeys.String())),
							},
						},
						"comment": schema.StringAttribute{
							Description: "Comment provided for user.",
							Computed:    true,
						},
					},
				},
			},
		},
	}

}
func (d *usersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(ProviderClientManager)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected ProviderClientManager, got: %T.", req.ProviderData),
		)
		return
	}

	d.clients = clients
}
func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state usersDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetNode := state.Node.ValueString()

	pveClient := d.clients[targetNode]
	if pveClient == nil {
		resp.Diagnostics.AddError(
			"No Proxmox Client Connection Available",
			fmt.Sprintf("Could not establish a connection manager pool entry to check node: %s", targetNode),
		)
		return
	}

	users, err := pveClient.Users(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Locate Proxmox Node",
			fmt.Sprintf("The node '%s' could not be resolved or fetched: %v", targetNode, err),
		)
		return
	}

	state.Users = make([]usersModel, 0, len(users))
	for _, user := range users {
		if user == nil {
			continue
		}

		groupSet, diag := types.SetValueFrom(ctx, types.StringType, user.Groups)
		if diag != nil {
			resp.Diagnostics.Append(diag...)
			return
		}

		state.Users = append(state.Users, usersModel{
			ID:             types.StringValue(user.UserID),
			RealmType:      types.StringValue(user.RealmType),
			Firstname:      types.StringValue(user.Firstname),
			Lastname:       types.StringValue(user.Lastname),
			Email:          types.StringValue(user.Email),
			Groups:         groupSet,
			Expire:         types.Int64Value(int64(user.Expire)),
			Enable:         types.BoolValue(bool(user.Enable)),
			TfaLockedUntil: types.StringValue(user.TFALockedUntil),
			TotpLocked:     types.BoolValue(bool(user.TOTPLocked)),
			Keys:           types.StringValue(user.Keys),
			Comment:        types.StringValue(user.Comment),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
