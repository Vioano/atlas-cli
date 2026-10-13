package atlas

// Desktop's local IPC adapter is isolated from the Atlas protocol. It uses the
// installed application's owner discovery and turn admission; no conversation
// is recreated, model setting changed, or public runtime port opened.
import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Desktop struct {
	conn   net.Conn
	client string
	owner  string
}

func desktopPath() string {
	if h := os.Getenv("CODEX_HOME"); h != "" {
		return filepath.Join(h, "ipc", "ipc.sock")
	}
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".codex", "ipc", "ipc.sock")
}
func OpenDesktop(ctx context.Context, thread string) (*Desktop, error) {
	conn, err := dialDesktop(ctx, desktopPath())
	if err != nil {
		return nil, err
	}
	d := &Desktop{conn: conn}
	r, err := d.call("initialize", Object{"clientType": "atlas-cli"}, 0)
	if err != nil {
		conn.Close()
		return nil, err
	}
	d.client = S(r, "clientId")
	r, err = d.call("thread-owner-discovery", Object{"hostId": "local", "conversationId": thread}, 1)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return d, nil
}
func (d *Desktop) Close() { d.conn.Close() }
func (d *Desktop) call(method string, params Object, version int) (Object, error) {
	id := ID()
	q := Object{"type": "request", "requestId": id, "method": method, "params": params, "version": version, "timeoutMs": 8000}
	if d.client != "" {
		q["sourceClientId"] = d.client
	}
	if d.owner != "" && method != "initialize" {
		q["targetClientId"] = d.owner
	}
	b, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	if len(b) > 1<<20 {
		return nil, fmt.Errorf("IPC request too large")
	}
	d.conn.SetDeadline(time.Now().Add(10 * time.Second))
	header := make([]byte, 4)
	binary.LittleEndian.PutUint32(header, uint32(len(b)))
	if _, err = d.conn.Write(append(header, b...)); err != nil {
		return nil, err
	}
	for {
		if _, err = io.ReadFull(d.conn, header); err != nil {
			return nil, err
		}
		n := binary.LittleEndian.Uint32(header)
		if n == 0 || n > 4<<20 {
			return nil, fmt.Errorf("invalid Desktop IPC frame")
		}
		b = make([]byte, n)
		if _, err = io.ReadFull(d.conn, b); err != nil {
			return nil, err
		}
		var r Object
		if err = json.Unmarshal(b, &r); err != nil {
			return nil, err
		}
		if S(r, "requestId") != id {
			continue
		}
		if S(r, "resultType") != "success" {
			return nil, fmt.Errorf("Desktop %s: %s", method, S(r, "error"))
		}
		if method == "thread-owner-discovery" {
			d.owner = S(r, "handledByClientId")
		}
		return O(r, "result"), nil
	}
}
func (d *Desktop) Settings(thread string) (Object, error) {
	return d.call("thread-follower-read-model-settings", Object{"conversationId": thread}, 1)
}
func (d *Desktop) Queue(thread, text string, messageID int) (Object, error) {
	h := sha256.Sum256([]byte("atlas/" + thread + "/" + fmt.Sprint(messageID)))
	hexID := hex.EncodeToString(h[:16])
	clientID := hexID[:8] + "-" + hexID[8:12] + "-" + hexID[12:16] + "-" + hexID[16:20] + "-" + hexID[20:]
	request := Object{"threadId": thread, "clientUserMessageId": clientID, "input": []Object{{"type": "text", "text": text, "text_elements": []any{}}}}
	return d.call("thread-follower-start-turn", Object{"conversationId": thread, "turnStart": Object{"request": request, "context": Object{"inheritThreadSettings": true}}}, 2)
}
