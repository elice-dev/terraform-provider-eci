package resource

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"terraform-provider-eci/internal/api"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const publicIpId = "11111111-2222-3333-4444-555555555555"

// publicIpServer fakes the API and returns the body of every PATCH it receives.
func publicIpServer(t *testing.T) (*api.APIClient, *[]map[string]interface{}) {
	t.Helper()
	var patches []map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/user/organization":
			_, _ = io.WriteString(w, `{"id":"`+publicIpId+`","created":"2026-11-18T00:00:00Z",
				"name":"org","ident":"org","allowed_ips":[]}`)
		case r.Method == http.MethodPatch:
			body := map[string]interface{}{}
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &body)
			patches = append(patches, body)
			_, _ = io.WriteString(w, `{"id":"`+publicIpId+`"}`)
		default:
			_, _ = io.WriteString(w, `{"id":"`+publicIpId+`","tags":{},
				"created":"2026-11-18T00:00:00Z","zone_id":"`+publicIpId+`",
				"organization_id":"`+publicIpId+`","dr":false,"pool_id":"`+publicIpId+`",
				"status":"active","ip":"1.2.3.4","pricing_id":"`+publicIpId+`",
				"pricing_type":"ondemand"}`)
		}
	}))
	t.Cleanup(srv.Close)

	client, err := api.NewAPIClient("token", srv.URL, "/api", publicIpId, false)
	if err != nil {
		t.Fatalf("NewAPIClient: %v", err)
	}
	return client, &patches
}

func publicIpValue(typ tftypes.Object, nic *string, tags map[string]string) tftypes.Value {
	values := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	tagValues := map[string]tftypes.Value{}
	for k, v := range tags {
		tagValues[k] = tftypes.NewValue(tftypes.String, v)
	}
	values["id"] = tftypes.NewValue(tftypes.String, publicIpId)
	values["dr"] = tftypes.NewValue(tftypes.Bool, false)
	values["pricing_id"] = tftypes.NewValue(tftypes.String, publicIpId)
	values["tags"] = tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, tagValues)
	if nic != nil {
		values["attached_network_interface_id"] = tftypes.NewValue(tftypes.String, *nic)
	}
	return tftypes.NewValue(typ, values)
}

func TestPublicIpUpdate_SendsAttachmentOnlyWhenChanged(t *testing.T) {
	ctx := context.Background()
	nic := "nic-1"
	newNic := "nic-2"

	cases := []struct {
		name            string
		stateNic        *string
		planNic         *string
		wantAttachments []interface{}
	}{
		// Sending null here would detach the network interface on a tags-only change.
		{"tags only, attached", &nic, &nic, nil},
		{"tags only, unattached", nil, nil, nil},
		{"attach", nil, &newNic, []interface{}{newNic}},
		{"move to another interface", &nic, &newNic, []interface{}{nil, newNic}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, patches := publicIpServer(t)
			r := &ResourcePublicIp{client: client}
			s := validSchema(t, r)
			typ := s.Type().TerraformType(ctx).(tftypes.Object)

			req := fwresource.UpdateRequest{
				State: tfsdk.State{
					Schema: s,
					Raw:    publicIpValue(typ, tc.stateNic, map[string]string{"a": "b"}),
				},
				Plan: tfsdk.Plan{
					Schema: s,
					Raw:    publicIpValue(typ, tc.planNic, map[string]string{}),
				},
			}
			resp := &fwresource.UpdateResponse{State: tfsdk.State{Schema: s, Raw: req.State.Raw}}
			r.Update(ctx, req, resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("Update: %v", resp.Diagnostics)
			}

			var attachments []interface{}
			for _, body := range *patches {
				if v, ok := body["attached_network_interface_id"]; ok {
					attachments = append(attachments, v)
				}
			}
			if len(attachments) != len(tc.wantAttachments) {
				t.Fatalf("attachments sent = %v, want %v", attachments, tc.wantAttachments)
			}
			for i := range attachments {
				if attachments[i] != tc.wantAttachments[i] {
					t.Errorf("attachments sent = %v, want %v", attachments, tc.wantAttachments)
				}
			}
		})
	}
}
