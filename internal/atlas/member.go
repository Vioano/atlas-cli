package atlas

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func HTTPSProfile(p, inv Object) error {
	endpoint := strings.TrimRight(S(inv, "https_endpoint"), "/")
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" {
		return fmt.Errorf("a verified HTTPS origin is required")
	}
	p["url"] = endpoint + "/atlas-api"
	p["transport"] = "https"
	if gw := p["gateway"]; gw != nil {
		p["ssh_gateway"] = gw
	}
	delete(p, "gateway")
	g := Clone(O(inv, "gitea"))
	g["url"] = endpoint + "/atlas-git"
	p["gitea"] = g
	return nil
}

func Join(ctx context.Context, key string) (Object, error) {
	inv, err := Unpack(key)
	if err != nil {
		return nil, err
	}
	actor := S(inv, "actor")
	if err = Name(actor); err != nil {
		return nil, err
	}
	path := filepath.Join(Home(), "members", actor, "identity.json")
	if old, e := Load(path); e == nil {
		if S(old, "actor") != actor || (S(old, "url") != S(inv, "url") && S(old, "url") != strings.TrimRight(S(inv, "https_endpoint"), "/")+"/atlas-api") {
			return nil, fmt.Errorf("different identity already exists; it will not be replaced")
		}
		if B(inv, "prefer_https") {
			if err = HTTPSProfile(old, inv); err != nil {
				return nil, err
			}
		}
		c := NewClient(old, nil, false)
		defer c.Transport.Close()
		m, err := c.Call(ctx, "GET", "/v1/manifest", nil)
		if err != nil {
			return nil, err
		}
		if B(inv, "prefer_https") {
			if err = Save(path, old); err != nil {
				return nil, err
			}
		}
		if err = Save(filepath.Join(Home(), "config.json"), Object{"mode": "member", "profile": path}); err != nil {
			return nil, err
		}
		return Object{"profile": path, "actor": actor, "grants": m["grants"], "resumed": true}, nil
	} else if !os.IsNotExist(e) {
		return nil, e
	}
	pendingPath := path + ".pending"
	pending, err := Load(pendingPath)
	if os.IsNotExist(err) {
		host, _ := os.Hostname()
		host = strings.Split(host, ".")[0]
		host = strings.Map(func(r rune) rune {
			if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
				return r
			}
			return '-'
		}, host)
		if len(host) > 60 {
			host = host[:60]
		}
		if host == "" {
			host = "device"
		}
		pending = Object{"invitation": S(inv, "invitation"), "credential": ID() + ID(), "enrollment_id": ID(), "device": host}
		if err = Save(pendingPath, pending); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	if S(pending, "invitation") != S(inv, "invitation") {
		return nil, fmt.Errorf("another enrollment is pending; original state preserved")
	}
	p := Object{"url": S(inv, "url"), "gateway": inv["gateway"], "gitea": inv["gitea"]}
	if B(inv, "prefer_https") {
		if err = HTTPSProfile(p, inv); err != nil {
			return nil, err
		}
	}
	c := NewClient(p, nil, false)
	defer func() { c.Transport.Close() }()
	if _, err = c.Call(ctx, "GET", "/health", nil); err != nil {
		if S(inv, "https_endpoint") == "" || B(inv, "prefer_https") {
			return nil, fmt.Errorf("transport unavailable before enrollment: %w", err)
		}
		c.Transport.Close()
		if err = HTTPSProfile(p, inv); err != nil {
			return nil, err
		}
		c = NewClient(p, nil, false)
		if _, err = c.Call(ctx, "GET", "/health", nil); err != nil {
			return nil, fmt.Errorf("SSH and HTTPS unavailable: %w", err)
		}
	}
	enrolled, err := c.Call(ctx, "POST", "/v1/enroll", pending)
	if err != nil {
		return nil, fmt.Errorf("enrollment: %w", err)
	}
	p["actor"] = S(enrolled, "actor")
	p["device"] = S(enrolled, "device")
	p["token"] = S(pending, "credential")
	p["project"] = inv["project"]
	p["channel"] = inv["channel"]
	p["protocol"] = "personal-atlas/1"
	if err = Save(path, p); err != nil {
		return nil, err
	}
	if err = Save(filepath.Join(Home(), "config.json"), Object{"mode": "member", "profile": path}); err != nil {
		return nil, err
	}
	os.Remove(pendingPath)
	m, err := NewClient(p, c.Transport, false).Call(ctx, "GET", "/v1/manifest", nil)
	if err != nil {
		return nil, err
	}
	git := NewClient(O(p, "gitea"), c.Transport, true)
	if _, err = git.Call(ctx, "GET", "/api/v1/repos/"+S(O(p, "gitea"), "repository"), nil); err != nil {
		return nil, fmt.Errorf("identity saved; repository check: %w", err)
	}
	return Object{"profile": path, "actor": actor, "grants": m["grants"], "project": p["project"], "channel": p["channel"], "repository": S(O(p, "gitea"), "repository"), "enrolled_identity_expires": false}, nil
}

func OpenProfile(settings Object) (Object, *Client, error) {
	path := S(settings, "profile")
	if path == "" {
		path = S(settings, "atlas_profile")
	}
	p, err := Load(path)
	if err != nil {
		return nil, nil, err
	}
	if gp := S(settings, "gateway_profile"); gp != "" {
		g, e := Load(gp)
		if e != nil {
			return nil, nil, e
		}
		p["gateway"] = g
	}
	return p, NewClient(p, nil, false), nil
}
func ImportLegacy(path string) (Object, error) {
	p, err := Load(path)
	if err != nil {
		return nil, err
	}
	if S(p, "token") == "" || S(p, "actor") == "" {
		return nil, fmt.Errorf("existing enrolled identity required")
	}
	return Object{"mode": "member", "profile": path, "legacy_identity_preserved": true}, nil
}
func Send(ctx context.Context, c *Client, channel, session, text, target, reply string) (Object, error) {
	if session == "" {
		return nil, fmt.Errorf("bind the current agent or provide --session")
	}
	id := ID()
	body := Object{"channel": channel, "session": session, "client_id": id, "text": text, "kind": "message", "delivery": "both"}
	if target != "" {
		body["to_sessions"] = []string{target}
	}
	if reply != "" {
		return nil, fmt.Errorf("use the numeric --reply-to option")
	}
	outbox := filepath.Join(Home(), "outbox", id+".json")
	if err := Save(outbox, body); err != nil {
		return nil, err
	}
	r, err := c.Call(ctx, "POST", "/v1/messages", body)
	if err != nil {
		return nil, fmt.Errorf("send outcome uncertain; retry with atlas retry %s: %w", id, err)
	}
	return r, nil
}
func Benchmark(ctx context.Context, c *Client, count int) (Object, error) {
	if count < 1 || count > 100 {
		return nil, fmt.Errorf("sample count must be 1 to 100")
	}
	var ms []float64
	var failures []string
	for i := 0; i < count; i++ {
		start := time.Now()
		_, err := c.Call(ctx, "GET", "/health", nil)
		elapsed := float64(time.Since(start).Microseconds()) / 1000
		if err != nil {
			failures = append(failures, err.Error())
			c.Transport.Close()
		} else {
			ms = append(ms, elapsed)
		}
	}
	return Object{"samples_ms": ms, "failures": failures, "transport": c.Transport.Mode, "count": count, "measured_at": time.Now().UTC().Format(time.RFC3339)}, nil
}
func History(ctx context.Context, c *Client, channel string, after int) (Object, error) {
	return c.Call(ctx, "GET", "/v1/messages?channel="+url.QueryEscape(channel)+"&after="+fmt.Sprint(after), nil)
}
