//go:build windows

package atlas

import (
	"context"
	"github.com/Microsoft/go-winio"
	"net"
)

func dialDesktop(ctx context.Context, _ string) (net.Conn, error) {
	return winio.DialPipeContext(ctx, `\\.\pipe\codex-ipc`)
}
