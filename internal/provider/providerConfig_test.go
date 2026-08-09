package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestResolveNodePort(t *testing.T) {
	t.Parallel()

	t.Run("uses node port when set", func(t *testing.T) {
		t.Parallel()
		got := resolveNodePort(
			types.Int32Value(1234),
			types.Int32Null(),
		)
		if got != 1234 {
			t.Fatalf("expected 1234, got %d", got)
		}
	})

	t.Run("uses global port when node port is null", func(t *testing.T) {
		t.Parallel()
		got := resolveNodePort(
			types.Int32Null(),
			types.Int32Value(4321),
		)
		if got != 4321 {
			t.Fatalf("expected 4321, got %d", got)
		}
	})

	t.Run("falls back to default when both are null", func(t *testing.T) {
		t.Parallel()
		got := resolveNodePort(
			types.Int32Null(),
			types.Int32Null(),
		)
		if got != 8006 {
			t.Fatalf("expected 8006, got %d", got)
		}
	})

	t.Run("treats unknown as not set (falls back)", func(t *testing.T) {
		t.Parallel()
		got := resolveNodePort(
			types.Int32Unknown(),
			types.Int32Null(),
		)
		if got != 8006 {
			t.Fatalf("expected 8006, got %d", got)
		}
	})
}

func TestResolveNodeIgnoreSSL(t *testing.T) {
	t.Parallel()

	t.Run("uses node ignore_ssl when set", func(t *testing.T) {
		t.Parallel()
		got := resolveNodeIgnoreSSL(
			types.BoolValue(true),
			types.BoolNull(),
		)
		if got != true {
			t.Fatalf("expected true, got %v", got)
		}
	})

	t.Run("uses global ignore_ssl when node is null", func(t *testing.T) {
		t.Parallel()
		got := resolveNodeIgnoreSSL(
			types.BoolNull(),
			types.BoolValue(false),
		)
		if got != false {
			t.Fatalf("expected false, got %v", got)
		}
	})

	t.Run("falls back to false when both are null", func(t *testing.T) {
		t.Parallel()
		got := resolveNodeIgnoreSSL(
			types.BoolNull(),
			types.BoolNull(),
		)
		if got != false {
			t.Fatalf("expected false, got %v", got)
		}
	})

	t.Run("unknown behaves as not set", func(t *testing.T) {
		t.Parallel()
		got := resolveNodeIgnoreSSL(
			types.BoolUnknown(),
			types.BoolNull(),
		)
		if got != false {
			t.Fatalf("expected false, got %v", got)
		}
	})
}

func TestResolveNodeCredential(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("uses node credentials when set", func(t *testing.T) {
		t.Parallel()
		nodeCred := mustCredentialObject(ctx, types.ObjectNull(credentialModelType), credentialModel{
			Username:    types.StringValue("u1"),
			Password:    types.StringValue("p1"),
			Otp:         types.StringValue("o1"),
			TokenName:   types.StringValue(""),
			TokenSecret: types.StringValue(""),
		})
		globalCred := types.ObjectNull(credentialModelType)

		got, diags := resolveNodeCredential(ctx, nodeCred, globalCred)
		if diags.HasError() {
			t.Fatalf("expected no diag errors, got: %v", diags)
		}
		if got.Username.ValueString() != "u1" {
			t.Fatalf("expected username u1, got %q", got.Username.ValueString())
		}
	})

	t.Run("falls back to global when node is null", func(t *testing.T) {
		t.Parallel()
		nodeCred := types.ObjectNull(credentialModelType)
		globalCred := mustCredentialObject(ctx, nodeCred, credentialModel{
			Username:    types.StringValue("gu"),
			Password:    types.StringValue("gp"),
			Otp:         types.StringValue(""),
			TokenName:   types.StringValue("tn"),
			TokenSecret: types.StringValue("ts"),
		})

		got, diags := resolveNodeCredential(ctx, nodeCred, globalCred)
		if diags.HasError() {
			t.Fatalf("expected no diag errors, got: %v", diags)
		}
		if got.TokenName.ValueString() != "tn" {
			t.Fatalf("expected token_name tn, got %q", got.TokenName.ValueString())
		}
	})

	t.Run("errors when both node and global are null", func(t *testing.T) {
		t.Parallel()
		nodeCred := types.ObjectNull(credentialModelType)
		globalCred := types.ObjectNull(credentialModelType)

		_, diags := resolveNodeCredential(ctx, nodeCred, globalCred)
		if !diags.HasError() {
			t.Fatalf("expected diag error, got none")
		}
	})
}

func TestResolveNodeConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	node := nodeModel{
		Name:       types.StringValue("n1"),
		Target:     types.StringValue("10.0.0.1"),
		Port:       types.Int32Null(),
		IgnoreSSL:  types.BoolValue(true),
		Credential: types.ObjectNull(credentialModelType),
	}

	global := pveProviderModel{
		Nodes:     types.ListNull(types.ObjectType{ /* unused in this helper */ }),
		Port:      types.Int32Value(9999),
		IgnoreSSL: types.BoolValue(false),
		Credential: mustCredentialObject(ctx, types.ObjectNull(credentialModelType), credentialModel{
			Username:    types.StringValue("user"),
			Password:    types.StringValue("pass"),
			Otp:         types.StringValue("otp"),
			TokenName:   types.StringValue(""),
			TokenSecret: types.StringValue(""),
		}),
	}

	got, diags := resolveNodeConfig(ctx, node, global)
	if diags.HasError() {
		t.Fatalf("expected no diag errors, got: %v", diags)
	}

	if got.Name != "n1" {
		t.Fatalf("expected Name n1, got %q", got.Name)
	}
	if got.Target != "10.0.0.1" {
		t.Fatalf("expected Target 10.0.0.1, got %q", got.Target)
	}
	if got.Port != 9999 {
		t.Fatalf("expected Port 9999, got %d", got.Port)
	}
	if got.IgnoreSSL != true {
		t.Fatalf("expected IgnoreSSL true (node overrides), got %v", got.IgnoreSSL)
	}
	if got.Username != "user" || got.Password != "pass" || got.Otp != "otp" {
		t.Fatalf("unexpected credentials mapping: %#v", got)
	}
}

/*
Helpers to build types.ObjectValue for credentialModel.
The framework needs an ObjectType. We’ll define it once here.
*/
var credentialModelType = map[string]attr.Type{
	"username":     types.StringType,
	"password":     types.StringType,
	"otp":          types.StringType,
	"token_name":   types.StringType,
	"token_secret": types.StringType,
}

func mustCredentialObject(_ context.Context, _ types.Object, c credentialModel) types.Object {
	av := map[string]attr.Value{
		"username":     c.Username,
		"password":     c.Password,
		"otp":          c.Otp,
		"token_name":   c.TokenName,
		"token_secret": c.TokenSecret,
	}
	obj, err := types.ObjectValue(credentialModelType, av)
	if err != nil {
		panic(err)
	}
	return obj
}

// ensure we import basetypes and Context packages in case of future adjustments
var _ = basetypes.ObjectAsOptions{}
