package main

import (
	a "github.com/Vioano/atlas-cli/internal/atlas"
	"strings"
	"testing"
)

func TestInvitationUsesRunningReleaseAndKeepsTwoLines(t *testing.T) {
	previous := version
	version = "9.8.7"
	t.Cleanup(func() { version = previous })
	for _, endpoint := range []string{"", "https://verified.example"} {
		key, err := a.Pack(a.Object{"actor": "fixture", "url": "https://verified.example/atlas-api", "https_endpoint": endpoint})
		if err != nil {
			t.Fatal(err)
		}
		for _, platform := range []string{"", "windows"} {
			text := invitationText(key, platform)
			if len(strings.Split(strings.TrimSpace(text), "\n")) != 2 || !strings.Contains(text, "join "+key+" --agent current") {
				t.Fatalf("%q invitation does not contain the original key and two-line flow", platform)
			}
			if platform != "windows" || endpoint == "" {
				if !strings.Contains(text, "/v9.8.7/") && !strings.Contains(text, "/tag/v9.8.7 ") {
					t.Fatalf("%q invitation does not install the running published release", platform)
				}
			} else if !strings.Contains(text, endpoint+"/atlas/v9.8.7/install.ps1") || !strings.Contains(text, "ATLAS_DOWNLOAD_BASE='"+endpoint+"/atlas/v9.8.7'") || !strings.Contains(text, "--https "+endpoint) {
				t.Fatal("Windows invitation lost the existing verified HTTPS entry")
			}
		}
	}
	version = "dev"
	key, _ := a.Pack(a.Object{"actor": "fixture", "url": "https://verified.example/atlas-api"})
	if strings.Contains(invitationText(key, "windows"), "/vdev/") || strings.Contains(invitationText(key, ""), "/tag/vdev") {
		t.Fatal("development build emitted a nonexistent public release tag")
	}
}
