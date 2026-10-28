package permission

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/telepath/telepath/internal/agent"
)

// serverName and toolName compose the permission-prompt tool reference that
// Claude Code expects: mcp__<serverName>__<toolName>.
const (
	serverName = "telepath"
	toolName   = "approve"
)

// MCPServer is a minimal in-process MCP server (Streamable HTTP transport) that
// exposes a single "approve" tool. Claude Code is launched with
// --permission-prompt-tool mcp__telepath__approve and an --mcp-config pointing
// at this server; when the agent wants a gated tool, Claude calls approve, which
// we forward to the supplied ApprovalFunc (which blocks on the Telegram user).
//
// A fresh server is started per interactive run and bound to a random loopback
// port, so no cross-process IPC or authentication is required.
type MCPServer struct {
	hook    agent.ApprovalFunc
	ln      net.Listener
	httpSrv *http.Server
}

// StartMCPServer binds a loopback listener and starts serving. Close it when the
// run finishes.
func StartMCPServer(hook agent.ApprovalFunc) (*MCPServer, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("bind mcp listener: %w", err)
	}
	s := &MCPServer{hook: hook, ln: ln}
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handle)
	s.httpSrv = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() { _ = s.httpSrv.Serve(ln) }()
	return s, nil
}

// URL is the endpoint Claude should POST to.
func (s *MCPServer) URL() string {
	return fmt.Sprintf("http://%s/mcp", s.ln.Addr().String())
}

// ToolRef is the value for --permission-prompt-tool.
func (s *MCPServer) ToolRef() string {
	return fmt.Sprintf("mcp__%s__%s", serverName, toolName)
}

// MCPConfigJSON is the content for --mcp-config registering this HTTP server.
func (s *MCPServer) MCPConfigJSON() []byte {
	cfg := map[string]any{
		"mcpServers": map[string]any{
			serverName: map[string]any{
				"type": "http",
				"url":  s.URL(),
			},
		},
	}
	b, _ := json.Marshal(cfg)
	return b
}

// Close shuts down the server.
func (s *MCPServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.httpSrv.Shutdown(ctx)
}

// --- JSON-RPC 2.0 plumbing -------------------------------------------------

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s *MCPServer) handle(w http.ResponseWriter, r *http.Request) {
	// The optional server->client SSE stream (GET) is not offered.
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req rpcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRPCError(w, nil, -32700, "parse error")
		return
	}

	// Notifications (no id) get an empty 202 and no body.
	isNotification := len(req.ID) == 0
	if isNotification {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	switch req.Method {
	case "initialize":
		writeRPCResult(w, req.ID, s.initializeResult(req.Params))
	case "tools/list":
		writeRPCResult(w, req.ID, s.toolsListResult())
	case "tools/call":
		result, rpcErr := s.toolsCall(r.Context(), req.Params)
		if rpcErr != nil {
			writeRPCError(w, req.ID, rpcErr.Code, rpcErr.Message)
			return
		}
		writeRPCResult(w, req.ID, result)
	case "ping":
		writeRPCResult(w, req.ID, map[string]any{})
	default:
		writeRPCError(w, req.ID, -32601, "method not found: "+req.Method)
	}
}

func (s *MCPServer) initializeResult(params json.RawMessage) map[string]any {
	version := "2025-06-18"
	var p struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	if len(params) > 0 && json.Unmarshal(params, &p) == nil && p.ProtocolVersion != "" {
		version = p.ProtocolVersion
	}
	return map[string]any{
		"protocolVersion": version,
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo":      map[string]any{"name": serverName, "version": "1.0.0"},
	}
}

func (s *MCPServer) toolsListResult() map[string]any {
	return map[string]any{
		"tools": []map[string]any{
			{
				"name":        toolName,
				"description": "Approve or deny a tool invocation. Telepath forwards the request to the user on Telegram and returns their decision.",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"tool_name": map[string]any{"type": "string"},
						"input":     map[string]any{"type": "object"},
					},
					"required":             []string{"tool_name", "input"},
					"additionalProperties": true,
				},
			},
		},
	}
}

// callArgs is the argument shape Claude Code sends to a permission-prompt tool.
type callArgs struct {
	ToolName string         `json:"tool_name"`
	Input    map[string]any `json:"input"`
}

func (s *MCPServer) toolsCall(ctx context.Context, params json.RawMessage) (map[string]any, *rpcError) {
	var p struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcError{Code: -32602, Message: "invalid params"}
	}
	if p.Name != toolName {
		return nil, &rpcError{Code: -32602, Message: "unknown tool: " + p.Name}
	}

	var args callArgs
	_ = json.Unmarshal(p.Arguments, &args)
	if args.Input == nil {
		args.Input = map[string]any{}
	}

	req := agent.ApprovalRequest{
		Tool:  args.ToolName,
		Input: args.Input,
		Risk:  AssessRisk(args.ToolName, args.Input),
	}
	decision := s.hook(ctx, req)

	// The permission-prompt tool must return its decision as a JSON string in
	// the (single) text content block.
	var payload map[string]any
	if decision.Allow {
		payload = map[string]any{"behavior": "allow"}
		if len(decision.UpdatedInput) > 0 {
			payload["updatedInput"] = json.RawMessage(decision.UpdatedInput)
		} else {
			payload["updatedInput"] = args.Input
		}
	} else {
		msg := decision.Note
		if msg == "" {
			msg = "denied by user"
		}
		payload = map[string]any{"behavior": "deny", "message": msg}
	}
	text, _ := json.Marshal(payload)

	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": string(text)},
		},
	}, nil
}

func writeRPCResult(w http.ResponseWriter, id json.RawMessage, result any) {
	writeJSON(w, rpcResponse{JSONRPC: "2.0", ID: id, Result: result})
}

func writeRPCError(w http.ResponseWriter, id json.RawMessage, code int, msg string) {
	writeJSON(w, rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
