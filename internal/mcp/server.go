package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cleeryy/airsenal/internal/cheats"
)

const protocolVersion = "2024-11-05"

// Server is an MCP server that communicates over stdio using JSON-RPC 2.0.
type Server struct {
	tools *tools
}

// NewServer creates a new MCP Server backed by the given cheatsheet store.
func NewServer(store *cheats.Store) *Server {
	return &Server{tools: newTools(store)}
}

// RunStdio reads JSON-RPC messages from stdin and writes responses to stdout.
// All log output is directed to stderr so it does not pollute the protocol stream.
func (s *Server) RunStdio() error {
	return s.run(os.Stdin, os.Stdout)
}

func (s *Server) run(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("mcp: malformed message: %v", err)
			continue
		}

		resp := s.dispatch(req)
		if resp == nil {
			continue // notification — no response expected
		}

		encoded, err := json.Marshal(resp)
		if err != nil {
			log.Printf("mcp: marshal error: %v", err)
			continue
		}
		fmt.Fprintf(out, "%s\n", encoded)
	}
	return scanner.Err()
}

func (s *Server) dispatch(req request) *response {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "initialized":
		return nil // notification, no response
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	case "ping":
		return &response{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{}}
	default:
		return errorResponse(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

func (s *Server) handleInitialize(req request) *response {
	result := initializeResult{
		ProtocolVersion: protocolVersion,
		ServerInfo:      serverInfo{Name: "airsenal", Version: "1.0.0"},
		Capabilities:    serverCapabilities{Tools: &toolsCapability{}},
	}
	return &response{JSONRPC: "2.0", ID: req.ID, Result: result}
}

func (s *Server) handleToolsList(req request) *response {
	return &response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  toolsListResult{Tools: s.tools.definitions()},
	}
}

func (s *Server) handleToolsCall(req request) *response {
	var params toolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, -32602, "invalid params: "+err.Error())
	}
	result := s.tools.call(params.Name, params.Arguments)
	return &response{JSONRPC: "2.0", ID: req.ID, Result: result}
}

func errorResponse(id json.RawMessage, code int, msg string) *response {
	return &response{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}}
}
