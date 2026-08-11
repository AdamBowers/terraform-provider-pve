package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type usersDataSourceModel struct {
	Node  types.String `tfsdk:"node"`
	Users []usersModel `tfsdk:"users"`
}

type usersModel struct {
	ID             types.String `tfsdk:"id"`
	RealmType      types.String `tfsdk:"realm_type"`
	Firstname      types.String `tfsdk:"firstname"`
	Lastname       types.String `tfsdk:"lastname"`
	Email          types.String `tfsdk:"email"`
	Groups         types.Set    `tfsdk:"groups"`
	Expire         types.Int64  `tfsdk:"expire"`
	Enable         types.Bool   `tfsdk:"enable"`
	TfaLockedUntil types.String `tfsdk:"tfa_locked_until"`
	TotpLocked     types.Bool   `tfsdk:"totp_locked"`
	Keys           types.String `tfsdk:"keys"`
	Comment        types.String `tfsdk:"comment"`
}
