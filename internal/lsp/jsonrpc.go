// jsonrpc.go is the minimal JSON-RPC 2.0 layer: the message envelope, the
// Content-Length framing codec, and the connection both the stdio transport
// and in-memory harnesses run on. Hand-rolled on encoding/json per the
// issue's recommendation — small surface, custom methods first-class, zero
// new dependencies.

package lsp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

// JSON-RPC 2.0 / LSP error codes (github.com/microsoft/language-server-protocol).
const (
	codeParseError           = -32700
	codeInvalidRequest       = -32600
	codeMethodNotFound       = -32601
	codeInvalidParams        = -32602
	codeInternalError        = -32603
	codeServerNotInitialized = -32002
)

// jsonrpcVersion is the protocol version every message carries.
const jsonrpcVersion = "2.0"

// static transport errors.
var (
	errNilID            = errors.New("lsp.ID: UnmarshalJSON on nil pointer")
	errNoContentLength  = errors.New("json-rpc frame without Content-Length")
	errBadContentLength = errors.New("invalid Content-Length header")
	errBodyParse        = errors.New("malformed JSON-RPC body")
	errHeaderRead       = errors.New("read json-rpc headers")
)

// ID is a JSON-RPC request id: a JSON number, string, or null. The raw JSON
// is preserved so responses echo the request's id byte-for-byte.
type ID json.RawMessage

// UnmarshalJSON stores the id's raw JSON verbatim.
func (id *ID) UnmarshalJSON(data []byte) error {
	if id == nil {
		return errNilID
	}

	*id = append((*id)[0:0], data...)

	return nil
}

// MarshalJSON emits the stored raw JSON; an unset id encodes as null.
func (id *ID) MarshalJSON() ([]byte, error) {
	if id == nil || len(*id) == 0 {
		return []byte("null"), nil
	}

	return *id, nil
}

// ResponseError is the JSON-RPC error object carried on failed responses.
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Message is one JSON-RPC 2.0 envelope: a request (id+method), a response
// (id+result/error), or a notification (method, no id).
type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *ID             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

// Conn is a JSON-RPC connection over Content-Length framed streams — the LSP
// stdio transport, but built on plain io.Reader/io.Writer so any pipe (net.Pipe
// in tests, in-memory buffers in the GUI app) drives it unchanged.
type Conn struct {
	r *bufio.Reader
	w io.Writer

	wmu sync.Mutex // serializes writes; reads are owned by the Serve loop
}

// NewConn builds a connection reading framed messages from r and writing
// framed messages to w.
func NewConn(r io.Reader, w io.Writer) *Conn {
	return &Conn{r: bufio.NewReader(r), w: w}
}

// Read blocks for the next framed message. A framing failure (bad header,
// truncated body) is a hard transport error; a malformed JSON body returns
// an error wrapping errBodyParse so the caller can answer -32700 and keep
// the connection alive.
func (c *Conn) Read() (*Message, error) {
	length, err := c.readHeaders()
	if err != nil {
		return nil, err
	}

	body := make([]byte, length)
	if _, err := io.ReadFull(c.r, body); err != nil {
		return nil, fmt.Errorf("read json-rpc body: %w", err)
	}

	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("%w: %w", errBodyParse, err)
	}

	return &msg, nil
}

// Notify sends a server→client notification.
func (c *Conn) Notify(method string, params any) error {
	raw, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("marshal %s params: %w", method, err)
	}

	return c.Write(&Message{JSONRPC: jsonrpcVersion, Method: method, Params: raw})
}

// Write frames and writes one message.
func (c *Conn) Write(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal json-rpc message: %w", err)
	}

	c.wmu.Lock()
	defer c.wmu.Unlock()

	if _, err := fmt.Fprintf(c.w, "Content-Length: %d\r\n\r\n%s", len(data), data); err != nil {
		return fmt.Errorf("write json-rpc message: %w", err)
	}

	return nil
}

// readHeaders consumes the \r\n-terminated header block and returns the
// declared Content-Length. The blank line terminates the block.
func (c *Conn) readHeaders() (int64, error) {
	var length int64

	for {
		line, err := c.r.ReadString('\n')
		if err != nil {
			return 0, fmt.Errorf("%w: %w", errHeaderRead, err)
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break // end of headers
		}

		name, value, found := strings.Cut(line, ":")
		if !found || !strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			continue // Content-Type and future headers are ignored
		}

		n, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil || n < 0 {
			return 0, fmt.Errorf("%w: %q", errBadContentLength, value)
		}

		length = n
	}

	if length == 0 {
		return 0, errNoContentLength
	}

	return length, nil
}
