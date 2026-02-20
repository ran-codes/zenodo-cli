package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/ran-codes/zenodo-cli/internal/api"
	"github.com/ran-codes/zenodo-cli/internal/config"
)

//go:embed instructions.md
var baseInstructions string

func main() {
	// Load config and resolve token using the same chain as the CLI.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	profile := config.ResolveProfile("")
	if profile == "" {
		profile = cfg.DefaultProfile()
	}

	kr := config.NewKeyring()
	token := config.ResolveTokenFull("", kr, cfg, profile)

	sandbox := os.Getenv("ZENODO_SANDBOX") == "true" || os.Getenv("ZENODO_SANDBOX") == "1"
	baseURL := cfg.ResolveBaseURL(profile, sandbox)

	client := api.NewClient(baseURL, token)

	// Build server instructions with user context.
	instructions := buildInstructions(cfg)

	// Create MCP server.
	s := server.NewMCPServer("zenodo", "0.1.0",
		server.WithInstructions(instructions),
	)

	// Register tools.
	s.AddTool(recordsListTool(), recordsListHandler(client))
	s.AddTool(recordsSearchTool(), recordsSearchHandler(client))
	s.AddTool(recordsGetTool(), recordsGetHandler(client))
	s.AddTool(recordsVersionsTool(), recordsVersionsHandler(client))
	s.AddTool(communitiesListTool(), communitiesListHandler(client))
	s.AddTool(licensesSearchTool(), licensesSearchHandler(client))

	// Run over stdio.
	stdio := server.NewStdioServer(s)
	if err := stdio.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
		log.Fatalf("mcp server error: %v", err)
	}
}

// buildInstructions combines the embedded markdown instructions with dynamic user context.
func buildInstructions(cfg *config.Config) string {
	instructions := baseInstructions

	// Append user's ORCID if configured.
	orcid := fmt.Sprintf("%v", cfg.Get("orcid"))
	if orcid != "" && orcid != "<nil>" {
		instructions += fmt.Sprintf("\n## User context\n\n- The authenticated user's ORCID is: %s\n- When asked about \"my records\" or similar, search both creators.orcid and contributors.orcid with this ORCID.\n", orcid)
	}

	return instructions
}

// jsonResult marshals v to JSON and returns it as an MCP text result.
func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling result: %w", err)
	}
	return mcp.NewToolResultText(string(b)), nil
}

// --- records_list ---

func recordsListTool() mcp.Tool {
	return mcp.NewTool("records_list",
		mcp.WithDescription("List the authenticated user's own Zenodo records and drafts"),
		mcp.WithNumber("page", mcp.Description("Page number (default 1)")),
		mcp.WithNumber("size", mcp.Description("Results per page (default 10)")),
		mcp.WithString("status", mcp.Description("Filter by status: draft or published")),
		mcp.WithString("sort", mcp.Description("Sort order, e.g. mostrecent, bestmatch")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
	)
}

func recordsListHandler(client *api.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := api.RecordListParams{
			Page:   req.GetInt("page", 1),
			Size:   req.GetInt("size", 10),
			Status: req.GetString("status", ""),
			Sort:   req.GetString("sort", ""),
		}
		records, err := client.ListUserRecords(params)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(records)
	}
}

// --- records_search ---

func recordsSearchTool() mcp.Tool {
	return mcp.NewTool("records_search",
		mcp.WithDescription("Search published Zenodo records using Elasticsearch query syntax"),
		mcp.WithString("q", mcp.Required(), mcp.Description("Search query (Elasticsearch syntax)")),
		mcp.WithNumber("page", mcp.Description("Page number (default 1)")),
		mcp.WithNumber("size", mcp.Description("Results per page (default 10)")),
		mcp.WithString("sort", mcp.Description("Sort order, e.g. mostrecent, bestmatch")),
		mcp.WithString("community", mcp.Description("Filter by community slug")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
	)
}

func recordsSearchHandler(client *api.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q, err := req.RequireString("q")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		params := api.RecordListParams{
			Page:      req.GetInt("page", 1),
			Size:      req.GetInt("size", 10),
			Sort:      req.GetString("sort", ""),
			Community: req.GetString("community", ""),
		}
		result, err := client.SearchRecords(q, params)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(result)
	}
}

// --- records_get ---

func recordsGetTool() mcp.Tool {
	return mcp.NewTool("records_get",
		mcp.WithDescription("Get a single published Zenodo record by its numeric ID"),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("Record ID")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
	)
}

func recordsGetHandler(client *api.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		record, err := client.GetRecord(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(record)
	}
}

// --- records_versions ---

func recordsVersionsTool() mcp.Tool {
	return mcp.NewTool("records_versions",
		mcp.WithDescription("List all versions of a Zenodo record"),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("Record ID (any version)")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
	)
}

func recordsVersionsHandler(client *api.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		result, err := client.ListVersions(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(result)
	}
}

// --- communities_list ---

func communitiesListTool() mcp.Tool {
	return mcp.NewTool("communities_list",
		mcp.WithDescription("List the authenticated user's Zenodo communities"),
		mcp.WithString("q", mcp.Description("Search query to filter communities")),
		mcp.WithNumber("page", mcp.Description("Page number (default 1)")),
		mcp.WithNumber("size", mcp.Description("Results per page (default 10)")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
	)
}

func communitiesListHandler(client *api.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := req.GetString("q", "")
		page := req.GetInt("page", 1)
		size := req.GetInt("size", 10)
		result, err := client.ListUserCommunities(q, page, size)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(result)
	}
}

// --- licenses_search ---

func licensesSearchTool() mcp.Tool {
	return mcp.NewTool("licenses_search",
		mcp.WithDescription("Search available Zenodo licenses"),
		mcp.WithString("q", mcp.Description("Search query to filter licenses")),
		mcp.WithNumber("page", mcp.Description("Page number (default 1)")),
		mcp.WithNumber("size", mcp.Description("Results per page (default 10)")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
	)
}

func licensesSearchHandler(client *api.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := req.GetString("q", "")
		page := req.GetInt("page", 1)
		size := req.GetInt("size", 10)
		result, err := client.SearchLicenses(q, page, size)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(result)
	}
}
