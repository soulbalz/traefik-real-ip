package traefik_real_ip

import (
	"context"
	"net"
	"net/http"
	"strings"
)

const (
	xRealIP        = "X-Real-Ip"
	xForwardedFor  = "X-Forwarded-For"
	cfConnectingIP = "Cf-Connecting-Ip"
)

// Config the plugin configuration.
type Config struct {
	ExcludedNets []string `json:"excludednets,omitempty" toml:"excludednets,omitempty" yaml:"excludednets,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		ExcludedNets: []string{},
	}
}

// RealIPOverWriter is a plugin that blocks incoming requests depending on their source IP.
type RealIPOverWriter struct {
	next         http.Handler
	name         string
	ExcludedNets []*net.IPNet
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	ipOverWriter := &RealIPOverWriter{
		next: next,
		name: name,
	}

	for _, v := range config.ExcludedNets {
		_, excludedNet, err := net.ParseCIDR(v)
		if err != nil {
			return nil, err
		}

		ipOverWriter.ExcludedNets = append(ipOverWriter.ExcludedNets, excludedNet)
	}

	return ipOverWriter, nil
}

func (r *RealIPOverWriter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	forwardedIPs := strings.Split(req.Header.Get(xForwardedFor), ",")

	// TODO - Implement a max for the iterations
	var realIP string
	for i := len(forwardedIPs) - 1; i >= 0; i-- {
		// TODO - Check if TrimSpace is necessary
		trimmedIP := strings.TrimSpace(forwardedIPs[i])
		if !r.excludedIP(trimmedIP) {
			realIP = trimmedIP
			break
		}
	}

	if realIP == "" {
		realIP = req.Header.Get(cfConnectingIP)
		req.Header.Set(xForwardedFor, realIP)
	}

	req.Header.Set(xRealIP, realIP)

	r.next.ServeHTTP(rw, req)
}

func (r *RealIPOverWriter) excludedIP(s string) bool {
	ip := net.ParseIP(s)
	if ip == nil {
		// log the error and fallback to the default value (check if true is ok)
		return true
	}

	for _, network := range r.ExcludedNets {
		if network.Contains(ip) {
			return true
		}
	}

	return false
}
