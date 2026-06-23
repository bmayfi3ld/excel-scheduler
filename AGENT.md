# Excel Scheduler Development Guidelines

## Build & Development Commands
- Install dependencies: `npm install`
- Start addin dev server: `just addin-dev` (alternatively: `cd addin && npm run dev-server`)
- Build addin for production: `cd addin && npm run build`
- Lint code: `cd addin && npm run lint`
- Fix linting issues: `cd addin && npm run lint:fix`
- Format code with prettier: `cd addin && npm run prettier`
- Start docs dev server: `just docs-dev` (alternatively: `cd docs && hugo server`)


## Organization
- there is an addin folder with a npm project using office addin Commands
- there is a docs folder with a hugo project in it
- to make linting work there is a package.json in the root and one in the addin folder, both need to be updated
- there are two package.json files, one for the ide at the top level, one that is used by the actual project at the lower level


## Code Style Guidelines
- Follow Office Add-in patterns and practices
- Use Office UI Fabric components for UI elements
- Format code using prettier with office-addin-prettier-config
- Use camelCase for variable and function names
- Include JSDoc comments for functions, especially custom Excel functions
- Follow semantic versioning for releases
- Update changelog.md with all changes using Keep a Changelog format
- Add proper error handling with descriptive error messages
- Use proper TypeScript types from @types/office-js where applicable


## Other
- read the readme to understand processes for this project

<!-- code-review-graph MCP tools -->
## MCP Tools: code-review-graph

**IMPORTANT: This project has a knowledge graph. ALWAYS use the
code-review-graph MCP tools BEFORE using Grep/Glob/Read to explore
the codebase.** The graph is faster, cheaper (fewer tokens), and gives
you structural context (callers, dependents, test coverage) that file
scanning cannot.

### When to use graph tools FIRST

- **Exploring code**: `semantic_search_nodes` or `query_graph` instead of Grep
- **Understanding impact**: `get_impact_radius` instead of manually tracing imports
- **Code review**: `detect_changes` + `get_review_context` instead of reading entire files
- **Finding relationships**: `query_graph` with callers_of/callees_of/imports_of/tests_for
- **Architecture questions**: `get_architecture_overview` + `list_communities`

Fall back to Grep/Glob/Read **only** when the graph doesn't cover what you need.

### Key Tools

| Tool | Use when |
| ------ | ---------- |
| `detect_changes` | Reviewing code changes — gives risk-scored analysis |
| `get_review_context` | Need source snippets for review — token-efficient |
| `get_impact_radius` | Understanding blast radius of a change |
| `get_affected_flows` | Finding which execution paths are impacted |
| `query_graph` | Tracing callers, callees, imports, tests, dependencies |
| `semantic_search_nodes` | Finding functions/classes by name or keyword |
| `get_architecture_overview` | Understanding high-level codebase structure |
| `refactor_tool` | Planning renames, finding dead code |

### Workflow

1. The graph auto-updates on file changes (via hooks).
2. Use `detect_changes` for code review.
3. Use `get_affected_flows` to understand impact.
4. Use `query_graph` pattern="tests_for" to check coverage.
