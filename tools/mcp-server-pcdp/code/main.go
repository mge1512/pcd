package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pcdp/mcp-server-pcdp/internal/lint"
	"github.com/pcdp/mcp-server-pcdp/internal/store"
)

const (
	schemaVersion = "0.3.17"
	serverVersion = "0.1.0"
	serverName    = "mcp-server-pcdp"
)

var (
	templateStore store.TemplateStore
	promptStore   store.PromptStore
	filesystem    store.Filesystem
)

func main() {
	// Parse transport selector and options
	transport := "stdio" // default
	listenAddr := "127.0.0.1:8080"

	if len(os.Args) > 1 {
		// Check for transport selector (bare word)
		if os.Args[1] == "stdio" || os.Args[1] == "http" {
			transport = os.Args[1]
		} else if !strings.HasPrefix(os.Args[1], "-") {
			fmt.Fprintf(os.Stderr, "error: unknown transport '%s'. Valid: stdio, http\n", os.Args[1])
			os.Exit(2)
		}

		// Parse key=value arguments
		for i := 1; i < len(os.Args); i++ {
			arg := os.Args[i]
			if strings.HasPrefix(arg, "listen=") {
				listenAddr = strings.TrimPrefix(arg, "listen=")
			} else if arg == "stdio" || arg == "http" {
				transport = arg
			}
		}
	}

	// Initialize stores
	templateStore = store.NewLayeredTemplateStore()
	promptStore = store.NewEmbeddedPromptStore()
	filesystem = store.NewOSFilesystem()

	// Create MCP server
	s := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
	)

	// Register tools (BEHAVIOR blocks)
	registerTools(s)

	// Register resources
	registerResources(s)

	// Run selected transport
	switch transport {
	case "stdio":
		runStdioTransport(s)
	case "http":
		runHTTPTransport(s, listenAddr)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown transport '%s'. Valid: stdio, http\n", transport)
		os.Exit(2)
	}
}

func registerTools(s *server.MCPServer) {
	// list_templates
	s.AddTool(
		mcp.NewTool("list_templates",
			mcp.WithDescription("List all installed PCDP templates"),
		),
		listTemplatesHandler,
	)

	// get_template
	s.AddTool(
		mcp.NewTool("get_template",
			mcp.WithDescription("Get a PCDP template by name and version"),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Template name (e.g., cli-tool, mcp-server)"),
			),
			mcp.WithString("version",
				mcp.Description("Template version (default: latest)"),
			),
		),
		getTemplateHandler,
	)

	// list_resources
	s.AddTool(
		mcp.NewTool("list_resources",
			mcp.WithDescription("List all available PCDP resources (templates, prompts, hints)"),
		),
		listResourcesHandler,
	)

	// read_resource
	s.AddTool(
		mcp.NewTool("read_resource",
			mcp.WithDescription("Read a PCDP resource by URI"),
			mcp.WithString("uri",
				mcp.Required(),
				mcp.Description("Resource URI (pcdp://type/name)"),
			),
		),
		readResourceHandler,
	)

	// lint_content
	s.AddTool(
		mcp.NewTool("lint_content",
			mcp.WithDescription("Validate a PCDP specification from string content"),
			mcp.WithString("content",
				mcp.Required(),
				mcp.Description("Full Markdown text of the PCDP specification"),
			),
			mcp.WithString("filename",
				mcp.Description("Filename for diagnostic references (default: spec.md)"),
			),
		),
		lintContentHandler,
	)

	// lint_file
	s.AddTool(
		mcp.NewTool("lint_file",
			mcp.WithDescription("Validate a PCDP specification from a file"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Absolute or relative path to a .md file"),
			),
		),
		lintFileHandler,
	)

	// get_schema_version
	s.AddTool(
		mcp.NewTool("get_schema_version",
			mcp.WithDescription("Get the Spec-Schema version this server was built against"),
		),
		getSchemaVersionHandler,
	)
}

func registerResources(s *server.MCPServer) {
	// Resource templates for dynamic URIs
	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"pcdp://templates/{name}",
			"PCDP template",
			mcp.WithTemplateDescription("PCDP deployment template"),
			mcp.WithTemplateMIMEType("text/markdown"),
		),
		templateResourceHandler,
	)

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"pcdp://prompts/{name}",
			"PCDP prompt",
			mcp.WithTemplateDescription("PCDP interview or translator prompt"),
			mcp.WithTemplateMIMEType("text/markdown"),
		),
		promptResourceHandler,
	)

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"pcdp://hints/{key}",
			"PCDP hints",
			mcp.WithTemplateDescription("PCDP library hints file"),
			mcp.WithTemplateMIMEType("text/markdown"),
		),
		hintsResourceHandler,
	)
}

// ============ BEHAVIOR: list_templates ============
func listTemplatesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	templates, err := templateStore.ListTemplates()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("store error: %v", err)), nil
	}

	// Omit content field from list results
	type templateRecord struct {
		Name     string `json:"name"`
		Version  string `json:"version"`
		Language string `json:"language"`
	}

	records := make([]templateRecord, len(templates))
	for i, t := range templates {
		records[i] = templateRecord{
			Name:     t.Name,
			Version:  t.Version,
			Language: t.Language,
		}
	}

	return mcp.NewToolResultText(marshalJSON(records)), nil
}


// ============ BEHAVIOR: get_template ============
func getTemplateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := req.GetString("name", "")
	version := req.GetString("version", "latest")

	if name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}

	template, err := templateStore.GetTemplate(name, version)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			if strings.Contains(err.Error(), "version") {
				return mcp.NewToolResultError(fmt.Sprintf("version %s not found for template %s", version, name)), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("unknown template: %s", name)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("store error: %v", err)), nil
	}

	return mcp.NewToolResultText(marshalJSON(template)), nil
}

// ============ BEHAVIOR: list_resources ============
func listResourcesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type resourceRecord struct {
		URI  string `json:"uri"`
		Name string `json:"name"`
	}

	var records []resourceRecord

	// List templates
	templates, err := templateStore.ListTemplates()
	if err == nil {
		for _, t := range templates {
			records = append(records, resourceRecord{
				URI:  fmt.Sprintf("pcdp://templates/%s", t.Name),
				Name: t.Name,
			})
		}
	}

	// List prompts (never fails)
	prompts := promptStore.ListPrompts()
	for _, p := range prompts {
		records = append(records, resourceRecord{
			URI:  fmt.Sprintf("pcdp://prompts/%s", p),
			Name: p,
		})
	}

	// List hints (from template store)
	hints := templateStore.ListHints()
	for _, h := range hints {
		records = append(records, resourceRecord{
			URI:  fmt.Sprintf("pcdp://hints/%s", h),
			Name: h,
		})
	}

	return mcp.NewToolResultText(marshalJSON(records)), nil
}

// ============ BEHAVIOR: read_resource ============
func readResourceHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	uri := req.GetString("uri", "")
	if uri == "" {
		return mcp.NewToolResultError("uri is required"), nil
	}

	// Parse URI: pcdp://<type>/<name>
	parts := strings.Split(uri, "://")
	if len(parts) != 2 || parts[0] != "pcdp" {
		return mcp.NewToolResultError(fmt.Sprintf("invalid resource URI: %s", uri)), nil
	}

	typeParts := strings.SplitN(parts[1], "/", 2)
	if len(typeParts) != 2 {
		return mcp.NewToolResultError(fmt.Sprintf("invalid resource URI: %s", uri)), nil
	}

	resType := typeParts[0]
	name := typeParts[1]

	type resourceRecord struct {
		URI     string `json:"uri"`
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	switch resType {
	case "templates":
		template, err := templateStore.GetTemplate(name, "latest")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("resource not found: %s", uri)), nil
		}
		return mcp.NewToolResultText(marshalJSON(resourceRecord{
			URI:     uri,
			Name:    name,
			Content: template.Content,
		})), nil

	case "prompts":
		content, err := promptStore.GetPrompt(name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("resource not found: %s", uri)), nil
		}
		return mcp.NewToolResultText(marshalJSON(resourceRecord{
			URI:     uri,
			Name:    name,
			Content: content,
		})), nil

	case "hints":
		content, err := templateStore.GetHints(name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("resource not found: %s", uri)), nil
		}
		return mcp.NewToolResultText(marshalJSON(resourceRecord{
			URI:     uri,
			Name:    name,
			Content: content,
		})), nil

	default:
		return mcp.NewToolResultError(fmt.Sprintf("unknown resource type: %s", resType)), nil
	}
}

// ============ BEHAVIOR: lint_content ============
func lintContentHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	content := req.GetString("content", "")
	filename := req.GetString("filename", "spec.md")

	if content == "" {
		return mcp.NewToolResultError("content is required"), nil
	}

	if !strings.HasSuffix(filename, ".md") {
		return mcp.NewToolResultError("filename must have .md extension"), nil
	}

	result := lint.LintContent(content, filename)
	return mcp.NewToolResultText(marshalJSON(result)), nil
}

// ============ BEHAVIOR: lint_file ============
func lintFileHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := req.GetString("path", "")
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	if !strings.HasSuffix(path, ".md") {
		return mcp.NewToolResultError(fmt.Sprintf("file must have .md extension: %s", path)), nil
	}

	content, err := filesystem.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return mcp.NewToolResultError(fmt.Sprintf("cannot open file: %s", path)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("read error: %s: %v", path, err)), nil
	}

	filename := filepath.Base(path)
	result := lint.LintContent(content, filename)
	return mcp.NewToolResultText(marshalJSON(result)), nil
}

// ============ BEHAVIOR: get_schema_version ============
func getSchemaVersionHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(fmt.Sprintf(`"%s"`, schemaVersion)), nil
}

// ============ Resource Handlers ============

func templateResourceHandler(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := req.Params.URI
	parts := strings.SplitN(strings.TrimPrefix(uri, "pcdp://templates/"), "/", 2)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid template URI")
	}
	name := parts[0]

	template, err := templateStore.GetTemplate(name, "latest")
	if err != nil {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "text/markdown",
			Text:     template.Content,
		},
	}, nil
}

func promptResourceHandler(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := req.Params.URI
	name := strings.TrimPrefix(uri, "pcdp://prompts/")

	content, err := promptStore.GetPrompt(name)
	if err != nil {
		return nil, fmt.Errorf("prompt not found: %s", name)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "text/markdown",
			Text:     content,
		},
	}, nil
}

func hintsResourceHandler(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := req.Params.URI
	key := strings.TrimPrefix(uri, "pcdp://hints/")

	content, err := templateStore.GetHints(key)
	if err != nil {
		return nil, fmt.Errorf("hints not found: %s", key)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "text/markdown",
			Text:     content,
		},
	}, nil
}

// ============ Transport Implementations ============

// BEHAVIOR: stdio-transport
func runStdioTransport(s *server.MCPServer) {
	if err := server.ServeStdio(s); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "stdio error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

// BEHAVIOR: http-transport
func runHTTPTransport(s *server.MCPServer, listenAddr string) {
	httpServer := server.NewStreamableHTTPServer(s)

	// Set up graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Start HTTP server in background
	var wg sync.WaitGroup
	var startErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpServer.Start(listenAddr); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			startErr = err
		}
	}()

	// Wait for signal or startup error
	select {
	case <-ctx.Done():
		// Signal received, proceed to shutdown
	case <-time.After(100 * time.Millisecond):
		// Give server time to start
		if startErr != nil {
			fmt.Fprintf(os.Stderr, "http bind error: %v\n", startErr)
			os.Exit(1)
		}
	}

	// Graceful shutdown
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(), 10*time.Second)
	defer cancel()
	httpServer.Shutdown(shutdownCtx)

	wg.Wait()
	os.Exit(0)
}

// ============ Utilities ============

func marshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}
