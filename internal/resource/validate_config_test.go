package resource_test

import (
	"context"
	"strings"
	"terraform-provider-eci/internal/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// validateConfig runs Terraform's config validation for typeName through the real provider server.
// Attributes missing from attrs are sent as null.
func validateConfig(
	t *testing.T, typeName string, build func(tftypes.Object) tftypes.Value,
) []*tfprotov6.Diagnostic {
	t.Helper()
	ctx := context.Background()

	server, err := providerserver.NewProtocol6WithError(provider.New("test", false)())()
	if err != nil {
		t.Fatalf("provider server: %v", err)
	}

	schemaResp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("GetProviderSchema: %v", err)
	}
	s, ok := schemaResp.ResourceSchemas[typeName]
	if !ok {
		t.Fatalf("resource %s is not registered in the provider", typeName)
	}

	typ := s.ValueType().(tftypes.Object)
	config, err := tfprotov6.NewDynamicValue(typ, build(typ))
	if err != nil {
		t.Fatalf("NewDynamicValue: %v", err)
	}

	resp, err := server.ValidateResourceConfig(ctx, &tfprotov6.ValidateResourceConfigRequest{
		TypeName: typeName,
		Config:   &config,
	})
	if err != nil {
		t.Fatalf("ValidateResourceConfig: %v", err)
	}
	return resp.Diagnostics
}

// object builds an object value of typ, filling attributes not in attrs with null.
func object(typ tftypes.Object, attrs map[string]tftypes.Value) tftypes.Value {
	full := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		if v, ok := attrs[name]; ok {
			full[name] = v
		} else {
			full[name] = tftypes.NewValue(attrType, nil)
		}
	}
	return tftypes.NewValue(typ, full)
}

func str(v string) tftypes.Value { return tftypes.NewValue(tftypes.String, v) }
func num(v int64) tftypes.Value  { return tftypes.NewValue(tftypes.Number, v) }

func strMap(m map[string]string) tftypes.Value {
	values := map[string]tftypes.Value{}
	for k, v := range m {
		values[k] = str(v)
	}
	return tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, values)
}

func errorSummaries(diags []*tfprotov6.Diagnostic) []string {
	var result []string
	for _, d := range diags {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			result = append(result, d.Summary+": "+d.Detail)
		}
	}
	return result
}

func assertValid(t *testing.T, diags []*tfprotov6.Diagnostic) {
	t.Helper()
	if errs := errorSummaries(diags); len(errs) > 0 {
		t.Errorf("expected valid config, got errors:\n%s", strings.Join(errs, "\n"))
	}
}

func assertInvalid(t *testing.T, diags []*tfprotov6.Diagnostic, wantSubstring string) {
	t.Helper()
	errs := errorSummaries(diags)
	for _, e := range errs {
		if strings.Contains(e, wantSubstring) {
			return
		}
	}
	t.Errorf("expected an error containing %q, got:\n%s", wantSubstring, strings.Join(errs, "\n"))
}
