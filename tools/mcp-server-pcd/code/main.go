// mcp-server-pcd — PCD MCP server
// Serves PCD templates, hints, prompts, and lint tools via MCP protocol.
// Supports stdio and streamable-HTTP transports.
// SPDX-License-Identifier: GPL-2.0-only

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
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"mcp-server-pcd/internal/lint"
	"mcp-server-pcd/internal/milestone"
	"mcp-server-pcd/internal/store"
)

// serverVersion is set at build time via -ldflags="-X main.serverVersion=..."
var serverVersion = "0.2.0"

func main() {
	// Parse transport and options from os.Args
	transport, listenAddr, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(2)
	}

	// Initialise asset store
	assetStore, err := store.NewEmbeddedLayeredStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to initialise asset store: %v\n", err)
		os.Exit(1)
	}

	// Build MCP server
	s := buildServer(assetStore)

	switch transport {
	case "stdio":
		runStdio(s)
	case "http":
		runHTTP(s, listenAddr)
	}
}

// parseArgs parses bare-word transport and key=value options.
// Returns transport ("stdio"|"http"), listenAddr, and any error.
func parseArgs(args []string) (transport, listenAddr string, err error) {
	transport = "stdio" // default
	listenAddr = "127.0.0.1:8080"

	var transports []string
	var listenSet bool

	for _, arg := range args {
		if arg == "stdio" || arg == "http" {
			transports = append(transports, arg)
		} else if strings.HasPrefix(arg, "listen=") {
			listenAddr = strings.TrimPrefix(arg, "listen=")
			listenSet = true
		} else {
			return "", "", fmt.Errorf("unknown argument '%s'. Valid transports: stdio, http", arg)
		}
	}

	if len(transports) > 1 {
		return "", "", fmt.Errorf("multiple transports specified: %s. Specify exactly one.", strings.Join(transports, ", "))
	}

	if len(transports) == 1 {
		transport = transports[0]
	}

	if listenSet && transport != "http" {
		fmt.Fprintf(os.Stderr, "warning: listen= argument ignored (not using http transport)\n")
	}

	return transport, listenAddr, nil
}

// buildServer constructs the MCP server with all tools and resources registered.
func buildServer(assetStore store.AssetStore) *server.MCPServer {
	s := server.NewMCPServer(
		"mcp-server-pcd",
		serverVersion,
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
	)

	fs := milestone.OSFilesystem{}

	// ── Tools ──────────────────────────────────────────────────────────────────

	// list_templates
	s.AddTool(
		mcp.NewTool("list_templates",
			mcp.WithDescription("List all installed PCD deployment templates. Returns name, version, and language for each; content is omitted."),
		),
		makeListTemplatesHandler(assetStore),
	)

	// get_template
	s.AddTool(
		mcp.NewTool("get_template",
			mcp.WithDescription("Get a PCD deployment template by name. Returns full Markdown content."),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Template name, e.g. 'cli-tool', 'mcp-server'"),
			),
			mcp.WithString("version",
				mcp.Description("Template version, e.g. '0.3.21' or 'latest' (default: latest)"),
			),
		),
		makeGetTemplateHandler(assetStore),
	)

	// list_resources
	s.AddTool(
		mcp.NewTool("list_resources",
			mcp.WithDescription("List all PCD resources (templates, hints, prompts) as resource URIs."),
		),
		makeListResourcesHandler(assetStore),
	)

	// read_resource
	s.AddTool(
		mcp.NewTool("read_resource",
			mcp.WithDescription("Read a PCD resource by URI. URI format: pcd://<type>/<name>. Types: templates, hints, prompts."),
			mcp.WithString("uri",
				mcp.Required(),
				mcp.Description("Resource URI, e.g. 'pcd://templates/cli-tool', 'pcd://hints/cli-tool.go.milestones', 'pcd://prompts/interview'"),
			),
		),
		makeReadResourceHandler(assetStore),
	)

	// lint_content
	s.AddTool(
		mcp.NewTool("lint_content",
			mcp.WithDescription("Validate a PCD specification given as a string. Applies RULE-01 through RULE-17."),
			mcp.WithString("content",
				mcp.Required(),
				mcp.Description("Full Markdown text of the PCD spec"),
			),
			mcp.WithString("filename",
				mcp.Required(),
				mcp.Description("Filename for diagnostics; must have .md extension"),
			),
		),
		makeLintContentHandler(),
	)

	// lint_file
	s.AddTool(
		mcp.NewTool("lint_file",
			mcp.WithDescription("Validate a PCD specification file on disk. Applies RULE-01 through RULE-17."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Absolute path to the spec .md file"),
			),
		),
		makeLintFileHandler(fs),
	)

	// get_schema_version
	s.AddTool(
		mcp.NewTool("get_schema_version",
			mcp.WithDescription("Return the PCD Spec-Schema version this binary was built against."),
		),
		makeGetSchemaVersionHandler(),
	)

	// set_milestone_status
	s.AddTool(
		mcp.NewTool("set_milestone_status",
			mcp.WithDescription("Set the Status: field of a named MILESTONE section in a spec file on disk."),
			mcp.WithString("spec_path",
				mcp.Required(),
				mcp.Description("Absolute path to the spec .md file"),
			),
			mcp.WithString("milestone_name",
				mcp.Required(),
				mcp.Description("Exact MILESTONE label, e.g. '0.1.0'"),
			),
			mcp.WithString("new_status",
				mcp.Required(),
				mcp.Description("New status: pending | active | failed | released"),
			),
		),
		makeSetMilestoneStatusHandler(fs),
	)

	// ── Resource templates (dynamic URIs) ──────────────────────────────────────

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"pcd://templates/{name}",
			"PCD deployment template",
			mcp.WithTemplateDescription("Deployment template for PCD components"),
			mcp.WithTemplateMIMEType("text/markdown"),
		),
		makeTemplateResourceHandler(assetStore),
	)

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"pcd://hints/{key}",
			"PCD hints file",
			mcp.WithTemplateDescription("Library and milestone hints for PCD translation"),
			mcp.WithTemplateMIMEType("text/markdown"),
		),
		makeHintsResourceHandler(assetStore),
	)

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"pcd://prompts/{name}",
			"PCD prompt",
			mcp.WithTemplateDescription("PCD translation or interview prompt"),
			mcp.WithTemplateMIMEType("text/markdown"),
		),
		makePromptResourceHandler(assetStore),
	)

	return s
}

// ── Tool handlers ─────────────────────────────────────────────────────────────

func makeListTemplatesHandler(s store.AssetStore) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		records, err := s.ListTemplates()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("internal error: %v", err)), nil
		}
		type entry struct {
			Name     string `json:"name"`
			Version  string `json:"version"`
			Language string `json:"language"`
		}
		entries := make([]entry, len(records))
		for i, r := range records {
			entries[i] = entry{Name: r.Name, Version: r.Version, Language: r.Language}
		}
		data, _ := json.Marshal(entries)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func makeGetTemplateHandler(s store.AssetStore) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := req.GetString("name", "")
		version := req.GetString("version", "")
		if version == "" {
			version = "latest"
		}
		rec, err := s.GetTemplate(name, version)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("unknown template: %s", name)), nil
		}
		type result struct {
			Name     string `json:"name"`
			Version  string `json:"version"`
			Language string `json:"language"`
			Content  string `json:"content"`
		}
		data, _ := json.Marshal(result{
			Name: rec.Name, Version: rec.Version,
			Language: rec.Language, Content: rec.Content,
		})
		return mcp.NewToolResultText(string(data)), nil
	}
}

func makeListResourcesHandler(s store.AssetStore) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		type entry struct {
			URI  string `json:"uri"`
			Name string `json:"name"`
		}
		var entries []entry

		templates, err := s.ListTemplates()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("internal error: %v", err)), nil
		}
		for _, t := range templates {
			entries = append(entries, entry{
				URI:  "pcd://templates/" + t.Name,
				Name: t.Name,
			})
		}

		hintsKeys, err := s.ListHintsKeys()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("internal error: %v", err)), nil
		}
		for _, k := range hintsKeys {
			entries = append(entries, entry{
				URI:  "pcd://hints/" + k,
				Name: k,
			})
		}

		promptNames, err := s.ListPrompts()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("internal error: %v", err)), nil
		}
		for _, n := range promptNames {
			entries = append(entries, entry{
				URI:  "pcd://prompts/" + n,
				Name: n,
			})
		}

		data, _ := json.Marshal(entries)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func makeReadResourceHandler(s store.AssetStore) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uri := req.GetString("uri", "")

		// Step 1: parse URI
		if !strings.HasPrefix(uri, "pcd://") {
			return mcp.NewToolResultError(fmt.Sprintf("invalid resource URI: %s", uri)), nil
		}
		rest := strings.TrimPrefix(uri, "pcd://")
		slashIdx := strings.Index(rest, "/")
		if slashIdx < 0 {
			return mcp.NewToolResultError(fmt.Sprintf("invalid resource URI: %s", uri)), nil
		}
		resourceType := rest[:slashIdx]
		resourceName := rest[slashIdx+1:]
		if resourceName == "" {
			return mcp.NewToolResultError(fmt.Sprintf("invalid resource URI: %s", uri)), nil
		}

		// Step 2: dispatch by type
		var content string
		var notFound bool

		switch resourceType {
		case "templates":
			rec, err := s.GetTemplate(resourceName, "latest")
			if err != nil {
				notFound = true
			} else {
				content = rec.Content
			}
		case "hints":
			c, err := s.GetHints(resourceName)
			if err != nil {
				notFound = true
			} else {
				content = c
			}
		case "prompts":
			c, err := s.GetPrompt(resourceName)
			if err != nil {
				notFound = true
			} else {
				content = c
			}
		default:
			return mcp.NewToolResultError(fmt.Sprintf("invalid resource URI: %s (unknown type '%s')", uri, resourceType)), nil
		}

		// Step 3: not found
		if notFound {
			return mcp.NewToolResultError(fmt.Sprintf("resource not found: %s", uri)), nil
		}

		// Step 4: return ResourceRecord
		type result struct {
			URI     string `json:"uri"`
			Name    string `json:"name"`
			Content string `json:"content"`
		}
		data, _ := json.Marshal(result{URI: uri, Name: resourceName, Content: content})
		return mcp.NewToolResultText(string(data)), nil
	}
}

func makeLintContentHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		content := req.GetString("content", "")
		filename := req.GetString("filename", "")

		// Step 1: validate filename extension
		if !strings.HasSuffix(filename, ".md") {
			return mcp.NewToolResultError(fmt.Sprintf("filename must have .md extension: %s", filename)), nil
		}

		// Step 2: run lint engine
		result := lint.LintContent(content, filename)

		// Step 3: return LintResult
		type diagJSON struct {
			Severity string `json:"severity"`
			Line     int    `json:"line"`
			Section  string `json:"section"`
			Message  string `json:"message"`
			Rule     string `json:"rule"`
		}
		type lintResultJSON struct {
			Valid        bool       `json:"valid"`
			Errors       int        `json:"errors"`
			Warnings     int        `json:"warnings"`
			Diagnostics  []diagJSON `json:"diagnostics"`
		}

		out := lintResultJSON{
			Valid:    result.Valid,
			Errors:   result.Errors,
			Warnings: result.Warnings,
		}
		for _, d := range result.Diagnostics {
			out.Diagnostics = append(out.Diagnostics, diagJSON{
				Severity: d.Severity.String(),
				Line:     d.Line,
				Section:  d.Section,
				Message:  d.Message,
				Rule:     d.Rule,
			})
		}
		if out.Diagnostics == nil {
			out.Diagnostics = []diagJSON{}
		}

		data, _ := json.Marshal(out)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func makeLintFileHandler(fs milestone.Filesystem) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := req.GetString("path", "")

		// Step 1: read file
		content, err := fs.ReadFile(path)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("cannot open file: %s", path)), nil
		}

		// Step 2: extract filename
		filename := filepath.Base(path)

		// Step 3: run lint_content logic
		if !strings.HasSuffix(filename, ".md") {
			return mcp.NewToolResultError(fmt.Sprintf("filename must have .md extension: %s", filename)), nil
		}
		result := lint.LintContent(content, filename)

		type diagJSON struct {
			Severity string `json:"severity"`
			Line     int    `json:"line"`
			Section  string `json:"section"`
			Message  string `json:"message"`
			Rule     string `json:"rule"`
		}
		type lintResultJSON struct {
			Valid       bool       `json:"valid"`
			Errors      int        `json:"errors"`
			Warnings    int        `json:"warnings"`
			Diagnostics []diagJSON `json:"diagnostics"`
		}

		out := lintResultJSON{
			Valid:    result.Valid,
			Errors:   result.Errors,
			Warnings: result.Warnings,
		}
		for _, d := range result.Diagnostics {
			out.Diagnostics = append(out.Diagnostics, diagJSON{
				Severity: d.Severity.String(),
				Line:     d.Line,
				Section:  d.Section,
				Message:  d.Message,
				Rule:     d.Rule,
			})
		}
		if out.Diagnostics == nil {
			out.Diagnostics = []diagJSON{}
		}

		data, _ := json.Marshal(out)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func makeGetSchemaVersionHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(lint.SpecSchema), nil
	}
}

func makeSetMilestoneStatusHandler(fs milestone.Filesystem) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		specPath := req.GetString("spec_path", "")
		milestoneName := req.GetString("milestone_name", "")
		newStatus := req.GetString("new_status", "")

		if !milestone.IsValidStatus(newStatus) {
			return mcp.NewToolResultError(fmt.Sprintf("invalid status value: %s. Valid: pending, active, failed, released", newStatus)), nil
		}

		result, err := milestone.SetStatus(fs, specPath, milestoneName, newStatus)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		type resultJSON struct {
			SpecPath       string `json:"spec_path"`
			MilestoneName  string `json:"milestone_name"`
			PreviousStatus string `json:"previous_status"`
			NewStatus      string `json:"new_status"`
		}
		data, _ := json.Marshal(resultJSON{
			SpecPath:       result.SpecPath,
			MilestoneName:  result.MilestoneName,
			PreviousStatus: string(result.PreviousStatus),
			NewStatus:      string(result.NewStatus),
		})
		return mcp.NewToolResultText(string(data)), nil
	}
}

// ── Resource handlers ─────────────────────────────────────────────────────────

func makeTemplateResourceHandler(s store.AssetStore) server.ResourceTemplateHandlerFunc {
	return func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		name := extractURIParam(req.Params.URI, "pcd://templates/")
		rec, err := s.GetTemplate(name, "latest")
		if err != nil {
			return nil, fmt.Errorf("resource not found: %s", req.Params.URI)
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      req.Params.URI,
				MIMEType: "text/markdown",
				Text:     rec.Content,
			},
		}, nil
	}
}

func makeHintsResourceHandler(s store.AssetStore) server.ResourceTemplateHandlerFunc {
	return func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		key := extractURIParam(req.Params.URI, "pcd://hints/")
		content, err := s.GetHints(key)
		if err != nil {
			return nil, fmt.Errorf("resource not found: %s", req.Params.URI)
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      req.Params.URI,
				MIMEType: "text/markdown",
				Text:     content,
			},
		}, nil
	}
}

func makePromptResourceHandler(s store.AssetStore) server.ResourceTemplateHandlerFunc {
	return func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		name := extractURIParam(req.Params.URI, "pcd://prompts/")
		content, err := s.GetPrompt(name)
		if err != nil {
			return nil, fmt.Errorf("resource not found: %s", req.Params.URI)
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      req.Params.URI,
				MIMEType: "text/markdown",
				Text:     content,
			},
		}, nil
	}
}

func extractURIParam(uri, prefix string) string {
	return strings.TrimPrefix(uri, prefix)
}

// ── Transport runners ─────────────────────────────────────────────────────────

func runStdio(s *server.MCPServer) {
	fmt.Fprintf(os.Stderr, "mcp-server-pcd %s starting (stdio transport)\n", serverVersion)
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "stdio error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func runHTTP(s *server.MCPServer, addr string) {
	fmt.Fprintf(os.Stderr, "mcp-server-pcd %s starting (http transport, listen=%s)\n", serverVersion, addr)

	httpServer := server.NewStreamableHTTPServer(s)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := httpServer.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		fmt.Fprintf(os.Stderr, "http bind error: %v\n", err)
		os.Exit(1)
	case <-ctx.Done():
		fmt.Fprintf(os.Stderr, "mcp-server-pcd shutting down\n")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
		os.Exit(0)
	}
}
