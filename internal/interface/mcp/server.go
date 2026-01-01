package mcp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/cookchen233/syzygy-mcp-go/internal/application"
	"github.com/cookchen233/syzygy-mcp-go/internal/infrastructure/persistence/fs"
)

type ServerConfig struct {
	Name    string
	Version string
	Logger  *log.Logger
}

type Server struct {
	cfg ServerConfig

	in  io.Reader
	out io.Writer

	app *application.App
}

func NewServer(cfg ServerConfig) *Server {
	if cfg.Logger == nil {
		cfg.Logger = log.New(os.Stderr, "syzygy-mcp: ", log.LstdFlags|log.LUTC)
	}

	dataDir := os.Getenv("SYZYGY_DATA_DIR")
	if dataDir == "" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			dataDir = filepath.Join(home, ".syzygy-data")
		} else {
			dataDir = "./syzygy-data"
		}
	}

	store := fs.NewFileStore(fs.FileStoreConfig{
		BaseDir: dataDir,
	})

	app := application.NewApp(store, cfg.Logger)

	return &Server{
		cfg: cfg,
		in:  os.Stdin,
		out: os.Stdout,
		app: app,
	}
}

func (s *Server) Run() error {
	s.cfg.Logger.Printf("starting %s %s", s.cfg.Name, s.cfg.Version)

	scanner := bufio.NewScanner(s.in)
	// Avoid token-too-long for big payloads
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 16*1024*1024)

	enc := json.NewEncoder(s.out)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			_ = enc.Encode(NewErrorResponse(nil, ErrParse, "invalid JSON", err.Error()))
			continue
		}

		resp := s.handle(req)
		if resp == nil {
			continue
		}

		if err := enc.Encode(resp); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (s *Server) handle(req JSONRPCRequest) *JSONRPCResponse {
	ctx := requestContext{StartAt: time.Now()}

	s.cfg.Logger.Printf("rpc method=%s id=%v", req.Method, req.ID)

	switch req.Method {
	case "initialize":
		resp := s.handleInitialize(req)
		return &resp
	case "tools/list":
		resp := s.handleToolsList(req)
		return &resp
	case "tools/call":
		resp := s.handleToolsCall(req, ctx)
		return &resp
	default:
		resp := NewErrorResponse(req.ID, ErrMethodNotFound, "method not found", req.Method)
		return &resp
	}
}

func (s *Server) handleInitialize(req JSONRPCRequest) JSONRPCResponse {
	result := map[string]any{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]any{
			"name":    s.cfg.Name,
			"version": s.cfg.Version,
		},
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
	}
	return NewResultResponse(req.ID, result)
}

func (s *Server) handleToolsList(req JSONRPCRequest) JSONRPCResponse {
	tools := s.app.ToolRegistry().ListTools()
	return NewResultResponse(req.ID, map[string]any{"tools": tools})
}

func (s *Server) handleToolsCall(req JSONRPCRequest, _ requestContext) JSONRPCResponse {
	var params ToolsCallParams
	if err := decodeParams(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, ErrInvalidParams, "invalid params", err.Error())
	}

	res, err := s.app.ToolRegistry().CallTool(params.Name, params.Arguments)
	if err != nil {
		var apiErr *application.AppError
		if errors.As(err, &apiErr) {
			return NewResultResponse(req.ID, map[string]any{
				"content": []map[string]any{{
					"type": "text",
					"text": fmt.Sprintf("ERROR: %s (%s)", apiErr.Message, apiErr.Code),
				}},
				"isError": true,
			})
		}

		return NewResultResponse(req.ID, map[string]any{
			"content": []map[string]any{{
				"type": "text",
				"text": fmt.Sprintf("ERROR: %v", err),
			}},
			"isError": true,
		})
	}

	return NewResultResponse(req.ID, map[string]any{
		"content": []map[string]any{{
			"type": "text",
			"text": mustJSON(res),
		}},
	})
}

func mustJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func decodeParams(raw any, out any) error {
	if raw == nil {
		return errors.New("missing params")
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

type requestContext struct {
	StartAt time.Time
}
