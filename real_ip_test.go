package traefik_real_ip_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	plugin "github.com/soulbalz/traefik-real-ip"
)

func TestNew(t *testing.T) {
	cfg := plugin.CreateConfig()
	cfg.ExcludedNets = []string{"127.0.0.1/24"}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := plugin.New(ctx, next, cfg, "traefik-real-ip")
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		header        string
		desc          string
		xForwardedFor string
		expected      string
	}{
		{
			header:        "X-Forwarded-For",
			desc:          "don't forward",
			xForwardedFor: "127.0.0.2",
			expected:      "",
		},
		{
			header:        "X-Forwarded-For",
			desc:          "forward",
			xForwardedFor: "10.0.0.1",
			expected:      "10.0.0.1",
		},
		{
			header:        "Cf-Connecting-Ip",
			desc:          "forward",
			xForwardedFor: "10.0.0.1",
			expected:      "10.0.0.1",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set(test.header, test.xForwardedFor)

			handler.ServeHTTP(recorder, req)

			assertHeader(t, req, "X-Real-Ip", test.expected)
		})
	}
}

func assertHeader(t *testing.T, req *http.Request, key, expected string) {
	t.Helper()

	if req.Header.Get(key) != expected {
		t.Errorf("invalid header value: %s", req.Header.Get(key))
	}
}
