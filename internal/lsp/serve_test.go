// serve_test.go runs the full stdio transport: raw Content-Length framed
// JSON-RPC bytes over an in-memory pipe (net.Pipe), driving Serve end to end
// — framing, dispatch, published diagnostics on the wire, and exit.

package lsp_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/Djarvur/c4drill/internal/lsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// frameClient is a raw-protocol LSP client over a net.Conn.
type frameClient struct {
	t *testing.T
	c net.Conn
	r *bufio.Reader
}

func dial(t *testing.T, c net.Conn) *frameClient {
	t.Helper()

	return &frameClient{t: t, c: c, r: bufio.NewReader(c)}
}

// send writes one framed message built from raw JSON (no structs — raw wire).
func (f *frameClient) send(raw string) {
	f.t.Helper()

	_, err := fmt.Fprintf(f.c, "Content-Length: %d\r\n\r\n%s", len(raw), raw)
	require.NoError(f.t, err)
}

// receive reads one framed message and decodes it into a generic map.
func (f *frameClient) receive() map[string]any {
	f.t.Helper()

	require.NoError(f.t, f.c.SetReadDeadline(time.Now().Add(10*time.Second)))

	line, err := f.r.ReadString('\n')
	require.NoError(f.t, err)
	require.Contains(f.t, line, "Content-Length:", "first header line, got %q", line)

	var length int

	_, err = fmt.Sscanf(line, "Content-Length: %d", &length)
	require.NoError(f.t, err)

	for {
		line, err = f.r.ReadString('\n')
		require.NoError(f.t, err)

		if line == "\r\n" || line == "\n" {
			break
		}
	}

	body := make([]byte, length)
	_, err = readFull(f.r, body)
	require.NoError(f.t, err)

	var msg map[string]any
	require.NoError(f.t, json.Unmarshal(body, &msg))

	return msg
}

// readFull is io.ReadFull without the import dance in test helpers.
func readFull(r *bufio.Reader, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := r.Read(buf[total:])
		total += n

		if err != nil {
			return total, err //nolint:wrapcheck // test-local io.ReadFull substitute
		}
	}

	return total, nil
}

func TestServeOverPipe(t *testing.T) {
	t.Parallel()

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	serveErr := make(chan error, 1)
	go func() { serveErr <- lsp.Serve(context.Background(), serverConn, serverConn) }()

	client := dial(t, clientConn)

	client.send(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	init := client.receive()
	require.Nil(t, init["error"])

	result, ok := init["result"].(map[string]any)
	require.True(t, ok, "initialize result is an object")

	caps, ok := result["capabilities"].(map[string]any)
	require.True(t, ok, "capabilities is an object")

	sync, ok := caps["textDocumentSync"].(map[string]any)
	require.True(t, ok, "textDocumentSync is an object")

	assert.InDelta(t, lsp.SyncFull, sync["change"], 0)

	client.send(`{"jsonrpc":"2.0","method":"initialized","params":{}}`)

	modelJSON, err := json.Marshal(`[web]
type = "system"
name = "Web"
description = "w"

[[web.link]]
peer = "missing"
description = "x"
`)
	require.NoError(t, err)

	docJSON := fmt.Sprintf(`{"uri":"file:///ws/model.toml","languageId":"toml","version":1,"text":%s}`, modelJSON)
	client.send(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":` + docJSON + `}}`)

	pub := client.receive()
	assert.Equal(t, "textDocument/publishDiagnostics", pub["method"])

	params, ok := pub["params"].(map[string]any)
	require.True(t, ok, "publish params is an object")

	assert.Equal(t, "file:///ws/model.toml", params["uri"])

	diags, ok := params["diagnostics"].([]any)
	require.True(t, ok, "diagnostics is an array")
	require.Len(t, diags, 1)

	first, ok := diags[0].(map[string]any)
	require.True(t, ok, "diagnostic is an object")

	assert.Contains(t, first["message"], `cannot resolve peer "missing"`)

	client.send(`{"jsonrpc":"2.0","id":2,"method":"shutdown"}`)
	shutdown := client.receive()
	assert.InDelta(t, 2, shutdown["id"], 0)
	require.Nil(t, shutdown["result"])

	client.send(`{"jsonrpc":"2.0","method":"exit"}`)

	select {
	case err := <-serveErr:
		require.NoError(t, err, "Serve returns nil after exit")
	case <-time.After(10 * time.Second):
		t.Fatal("Serve did not return after exit")
	}
}

func TestServeAnswersMalformedBodyAndContinues(t *testing.T) {
	t.Parallel()

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	serveErr := make(chan error, 1)
	go func() { serveErr <- lsp.Serve(context.Background(), serverConn, serverConn) }()

	client := dial(t, clientConn)

	client.send(`{not json`)
	bad := client.receive()
	require.NotNil(t, bad["error"])

	assert.InDelta(t, -32700, bad["error"].(map[string]any)["code"], 0) //nolint:forcetypeassert // shape asserted above

	// The connection stays up: a well-formed request still works.
	client.send(`{"jsonrpc":"2.0","id":9,"method":"initialize","params":{}}`)
	ok := client.receive()
	assert.Nil(t, ok["error"])

	client.send(`{"jsonrpc":"2.0","method":"exit"}`)

	select {
	case err := <-serveErr:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Serve did not return after exit")
	}
}
