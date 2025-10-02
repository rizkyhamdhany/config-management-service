package validator

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaValidator_Validate(t *testing.T) {
	sv := NewSchemaValidator()

	type args struct {
		schema string
		data   json.RawMessage
	}
	type want struct {
		ok          bool
		errContains []string // any of these substrings may appear
	}

	cases := []struct {
		name string
		in   args
		out  want
	}{
		{
			name: "when unknown schema type should return error",
			in:   args{schema: "does_not_exist", data: json.RawMessage(`{}`)},
			out:  want{ok: false, errContains: []string{"unknown config type: does_not_exist"}},
		},
		{
			name: "when malformed json should return error",
			in:   args{schema: "feature_toggle", data: json.RawMessage(`{`)},
			out:  want{ok: false, errContains: []string{"unexpected EOF", "unexpected end of JSON input"}},
		},
		{
			name: "when feature_toggle valid should return nil",
			in:   args{schema: "feature_toggle", data: json.RawMessage(`{"enabled": true, "rollout_percentage": 10, "tags":["a","b"], "description":"ok"}`)},
			out:  want{ok: true},
		},
		{
			name: "when feature_toggle missing required 'enabled' should return error",
			in:   args{schema: "feature_toggle", data: json.RawMessage(`{"rollout_percentage": 10}`)},
			out:  want{ok: false, errContains: []string{"(root): enabled is required", "enabled: enabled is required"}},
		},
		{
			name: "when feature_toggle additional property rejected should return error",
			in:   args{schema: "feature_toggle", data: json.RawMessage(`{"enabled": true, "unknown": 1}`)},
			out:  want{ok: false, errContains: []string{"Additional property unknown is not allowed", "unknown additional property"}},
		},

		{
			name: "when experiment_config invalid variants (<2) should return error",
			in: args{
				schema: "experiment_config",
				data: json.RawMessage(`{
					"experiment_key":"exp-1",
					"active": true,
					"variants":[{"name":"A","weight":0.5}]
				}`),
			},
			out: want{ok: false, errContains: []string{"variants: Array must have at least 2 items"}},
		},
		{
			name: "when experiment_config valid should return nil",
			in: args{
				schema: "experiment_config",
				data: json.RawMessage(`{
					"experiment_key":"exp-1",
					"active": true,
					"variants":[{"name":"A","weight":0.5}, {"name":"B","weight":0.5}],
					"audience":{"countries":["ID","SG"],"os":["ios","android"]},
					"description":"A/B test"
				}`),
			},
			out: want{ok: true},
		},

		{
			name: "when service_client invalid uri should return error",
			in:   args{schema: "service_client", data: json.RawMessage(`{"name":"svc","base_url":"not-a-uri","timeout_ms":200}`)},
			out:  want{ok: false, errContains: []string{"base_url: Does not match format 'uri'"}},
		},

		{
			name: "when rate_limit_policy invalid identifier_type should return error",
			in:   args{schema: "rate_limit_policy", data: json.RawMessage(`{"identifier_type":"device","window_seconds":60,"max_requests":100}`)},
			out:  want{ok: false, errContains: []string{"identifier_type must be one of the following", "identifier_type: identifier_type must be one of the following"}},
		},

		{
			name: "when notification_policy valid",
			in:   args{schema: "notification_policy", data: json.RawMessage(`{"channel":"email","enabled":true,"daily_limit":10,"template_id":"tpl","placeholders":["name"]}`)},
			out:  want{ok: true},
		},

		{
			name: "when schedule_rule invalid window - missing end should return error",
			in: args{
				schema: "schedule_rule",
				data: json.RawMessage(`{
					"active": true,
					"timezone": "Asia/Jakarta",
					"cron": "0 0 * * *",
					"windows":[{"start":"2025-10-01T00:00:00Z"}]
				}`),
			},
			out: want{ok: false, errContains: []string{"windows.0: end is required", "windows.0.end: end is required"}},
		},

		{
			name: "when threshold_policy valid with null min/max",
			in:   args{schema: "threshold_policy", data: json.RawMessage(`{"metric":"p95","unit":"ms","min":null,"max":null,"inclusive":true,"enabled":true}`)},
			out:  want{ok: true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := sv.Validate(tc.in.schema, tc.in.data)
			if tc.out.ok {
				assert.NoError(t, err)
				return
			}
			if assert.Error(t, err) {
				msg := err.Error()
				found := false
				for _, sub := range tc.out.errContains {
					if sub != "" && strings.Contains(msg, sub) {
						found = true
						break
					}
				}
				if !found {
					assert.Failf(t, "error substring match", "error %q does not contain any of %#v", msg, tc.out.errContains)
				}
			}
		})
	}
}
