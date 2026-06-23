package mcpserver

import (
	"context"
	"embed"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// guides holds the operating-manual markdown surfaced as guide:// resources and
// woven into the prompts. views holds the canonical live-view HTML skeletons
// surfaced as view:// resources and returned by the live_view prompt. Both are
// embedded so the binary carries its own instructions offline.
//
//go:embed guides/*.md views/*.html
var docs embed.FS

// instructions is the always-on operating manual the server advertises to every
// client on connect. It teaches the .db model and the render-not-compute rule,
// then points at the resources and prompts for detail.
const instructions = `Quilt manages school-style schedules. A schedule is ONE self-contained SQLite ` +
	`.db file — there is no shared server state. Every tool takes a "db" path argument, so one ` +
	`session can work on many schedules by passing different paths.

Core loop: build structure (add_class / add_timeslot / add_cohort), configure rules ` +
	`(enable_rule / set_travel / add_blackout), fill cells (assign / unassign), then read ` +
	`(board / validate / report).

The golden rule: the engine decides validity; tools and views only render. After ANY ` +
	`assign/unassign or structural edit, re-call "board" (grid + violations + timeslot metadata ` +
	`in one call) or "validate" to get fresh violations — never infer validity yourself.

To build a live HTML view, prefer the "board" tool and use the live_view prompt. For more ` +
	`detail read the guide:// resources (guide://db-model, guide://migrate-xlsx, ` +
	`guide://live-views, guide://tools) and the view:// skeletons (view://master, view://class).`

// guidePage names an embedded markdown guide and how to advertise it.
type guidePage struct {
	uri, file, title, desc string
}

var guidePages = []guidePage{
	{"guide://db-model", "guides/db-model.md", "The .db model", "How a Quilt schedule is stored and its lifecycle."},
	{"guide://migrate-xlsx", "guides/migrate-xlsx.md", "Migrate an .xlsx workbook", "Importing a legacy Excel workbook into a .db."},
	{"guide://live-views", "guides/live-views.md", "Live views cookbook", "Building self-refreshing HTML views on the board tool, plus the Canva handoff."},
	{"guide://tools", "guides/tools.md", "Quilt MCP tools", "Reference for every tool the server exposes."},
}

// viewPage names an embedded live-view HTML skeleton.
type viewPage struct {
	uri, file, title, desc string
}

var viewPages = []viewPage{
	{"view://master", "views/master.html", "Master schedule view", "Classes × timeslots grid with day tabs and all four rule colors."},
	{"view://class", "views/class.html", "Per-class view", "One class as a Day × Period calendar — the Canva per-class layout."},
}

// registerDocs installs the guide:// and view:// resources and the prompts.
func registerDocs(s *mcp.Server) {
	for _, g := range guidePages {
		s.AddResource(&mcp.Resource{
			URI: g.uri, Name: g.title, Description: g.desc, MIMEType: "text/markdown",
		}, embeddedResource(g.uri, g.file, "text/markdown"))
	}
	for _, v := range viewPages {
		s.AddResource(&mcp.Resource{
			URI: v.uri, Name: v.title, Description: v.desc, MIMEType: "text/html",
		}, embeddedResource(v.uri, v.file, "text/html"))
	}
	registerPrompts(s)
}

// embeddedResource serves one embedded file as a resource.
func embeddedResource(uri, file, mime string) mcp.ResourceHandler {
	return func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		b, err := docs.ReadFile(file)
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{URI: uri, MIMEType: mime, Text: string(b)}},
		}, nil
	}
}

// mustRead returns an embedded file's contents, panicking on error — used only
// for files this package itself embeds, so a failure is a build/programmer bug.
func mustRead(file string) string {
	b, err := docs.ReadFile(file)
	if err != nil {
		panic(fmt.Sprintf("mcpserver: embedded file %q: %v", file, err))
	}
	return string(b)
}
