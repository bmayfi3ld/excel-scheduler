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