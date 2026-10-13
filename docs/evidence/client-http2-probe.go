//go:build darwin

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"
)

func main() {
	const host = "sub.kessokuu.top"
	const origin = "45.76.168.154:443"
	iface, err := net.InterfaceByName("en0")
	if err != nil {
		panic(err)
	}
	names := []string{"cloudflare_current_tun", "origin_current_tun", "origin_physical_wifi_socket"}
	clients := map[string]*http.Client{}
	proof := map[string]bool{}
	for _, name := range names {
		dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 20 * time.Second}
		if name == "origin_physical_wifi_socket" {
			dialer.Control = func(_, _ string, raw syscall.RawConn) error {
				var inner error
				err := raw.Control(func(fd uintptr) {
					inner = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, 25, iface.Index)
					if inner == nil {
						var index int
						index, inner = syscall.GetsockoptInt(int(fd), syscall.IPPROTO_IP, 25)
						proof[name] = index == iface.Index
					}
				})
				if err != nil {
					return err
				}
				return inner
			}
		}
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.Proxy = nil
		transport.ForceAttemptHTTP2 = true
		transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
			if name != "cloudflare_current_tun" {
				address = origin
			}
			return dialer.DialContext(ctx, network, address)
		}
		clients[name] = &http.Client{Transport: transport, Timeout: 15 * time.Second}
	}
	var samples []map[string]any
	for i := 0; i < 11; i++ {
		for offset := range names {
			name := names[(i+offset)%len(names)]
			start := time.Now()
			response, err := clients[name].Get("https://" + host + "/atlas-api/health")
			if err != nil {
				panic(fmt.Errorf("%s: %w", name, err))
			}
			body, err := io.ReadAll(response.Body)
			response.Body.Close()
			var health map[string]any
			if err != nil || response.StatusCode != 200 || json.Unmarshal(body, &health) != nil || health["status"] != "ok" {
				panic("invalid health response")
			}
			kind := "persistent_https"
			if i == 0 {
				kind = "cold_https"
			}
			samples = append(samples, map[string]any{"route": name, "kind": kind, "seconds": time.Since(start).Seconds(), "protocol": response.Proto})
		}
	}
	for _, client := range clients {
		client.CloseIdleConnections()
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"scope": "Native Go on captain Mac, standard certificate verification, HTTP/2 allowed, concurrent connections and rotated order. No system route or VPN change.", "samples": samples, "physical_interface_socket_verified": proof["origin_physical_wifi_socket"]})
}
