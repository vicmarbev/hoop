package apiconnections

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/hoophq/hoop/gateway/models"
	"github.com/hoophq/hoop/gateway/pgrest"
	"github.com/hoophq/hoop/gateway/storagev2"
	"github.com/hoophq/hoop/gateway/storagev2/types"
	"github.com/stretchr/testify/assert"
)

type clientFunc func(req *http.Request) (*http.Response, error)

func (f clientFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

func createTestServer(plConn []*pgrest.PluginConnection) clientFunc {
	return clientFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/plugins":
			body := `{"id": "", "org_id": "", "name": "access_control"}`
			return &http.Response{
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
			}, nil
		case "/plugin_connections":
			pluginConnJson, _ := json.Marshal(plConn)
			return &http.Response{
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(pluginConnJson)),
			}, nil
		}
		return &http.Response{
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString(`{"msg": "test not implemented"}`)),
		}, nil

	})
}

func TestAccessControlAllowed(t *testing.T) {
	u, _ := url.Parse("http://localhost:3000")
	pgrest.WithBaseURL(u)
	for _, tt := range []struct {
		msg                string
		allow              bool
		wantConnectionName string
		groups             []string
		fakeClient         clientFunc
	}{
		{
			msg:        "it should allow access to admin users",
			allow:      true,
			groups:     []string{types.GroupAdmin},
			fakeClient: createTestServer([]*pgrest.PluginConnection{}),
		},
		{
			msg:                "it should allow access to users in the allowed groups and with the allowed connection",
			allow:              true,
			wantConnectionName: "bash",
			fakeClient: createTestServer([]*pgrest.PluginConnection{
				{ConnectionConfig: []string{"sre"}, Connection: pgrest.Connection{Name: "bash"}},
			}),
			groups: []string{"sre"},
		},
		{
			msg:                "it should allow access when the user has multiple groups and one of them is allowed",
			allow:              true,
			wantConnectionName: "bash",
			fakeClient: createTestServer([]*pgrest.PluginConnection{
				{ConnectionConfig: []string{"support"}, Connection: pgrest.Connection{Name: "bash"}},
			}),
			groups: []string{"sre", "support", "devops"},
		},
		{
			msg:                "it should deny access if the connection is not found",
			allow:              false,
			wantConnectionName: "bash-not-found",
			fakeClient: createTestServer([]*pgrest.PluginConnection{
				{ConnectionConfig: []string{"sre"}, Connection: pgrest.Connection{Name: "bash"}},
			}),
			groups: []string{"sre"},
		},
		{
			msg:                "it should deny access if the groups does not match",
			allow:              false,
			wantConnectionName: "bash",
			fakeClient: createTestServer([]*pgrest.PluginConnection{
				{ConnectionConfig: []string{"sre"}, Connection: pgrest.Connection{Name: "bash"}},
			}),
			groups: []string{""},
		},
	} {
		t.Run(tt.msg, func(t *testing.T) {
			pgrest.WithHttpClient(tt.fakeClient)
			ctx := storagev2.NewOrganizationContext("").WithUserInfo("", "", "", "", tt.groups)
			allowed, err := accessControlAllowed(ctx)
			if err != nil {
				t.Fatalf("did not expect error, got %v", err)
			}
			got := allowed(tt.wantConnectionName)
			if got != tt.allow {
				t.Errorf("expected %v, got %v", tt.allow, got)
			}
		})
	}
}

func TestConnectionFilterOptions(t *testing.T) {
	for _, tt := range []struct {
		msg     string
		opts    map[string]string
		want    models.ConnectionFilterOption
		wantErr string
	}{
		{
			msg:  "it must be able to accept all options",
			opts: map[string]string{"type": "database", "subtype": "postgres", "managed_by": "hoopagent", "tags": "prod,devops"},
			want: models.ConnectionFilterOption{Type: "database", SubType: "postgres", ManagedBy: "hoopagent", Tags: []string{"prod", "devops"}},
		},
		{
			msg:  "it must ignore unknown options",
			opts: map[string]string{"unknown_option": "val", "tags.foo.bar": "val"},
			want: models.ConnectionFilterOption{},
		},
		{
			msg:     "it must error with invalid option values",
			opts:    map[string]string{"subtype": "value with space"},
			wantErr: errInvalidOptionVal.Error(),
		},
		{
			msg:     "it must error with invalid option values, special characteres",
			opts:    map[string]string{"subtype": "value&^%$#@"},
			wantErr: errInvalidOptionVal.Error(),
		},
		{
			msg:     "it must error when tag values has invalid option values",
			opts:    map[string]string{"tags": "foo,tag with space"},
			wantErr: errInvalidOptionVal.Error(),
		},
		{
			msg:     "it must error when tag values are empty",
			opts:    map[string]string{"tags": "foo,,,,"},
			wantErr: errInvalidOptionVal.Error(),
		},
	} {
		t.Run(tt.msg, func(t *testing.T) {
			urlValues := url.Values{}
			for key, val := range tt.opts {
				urlValues[key] = []string{val}
			}
			got, err := validateListOptions(urlValues)
			if err != nil {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
