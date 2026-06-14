# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Roadmap
Nothing right now

## Unreleased
### Added
- golangci-lint configuration (`.golangci.yml`) focused on security (gosec, bodyclose, noctx) and complexity (gocyclo, gocognit, cyclop, funlen, nestif, maintidx)
- `just validate` target that runs the Go unit tests and golangci-lint, plus `just setup-hooks` to enable the committed git hooks
- Pre-commit hook (`.githooks/pre-commit`) that runs `just validate` before each commit
- Claude Code post-edit hook that reminds the agent to run `just validate` after editing `.go` files

## v0.5.4 - 2026.06.13
### Security
- Remediated all open Dependabot alerts across both lockfiles (0 vulnerabilities remaining)
- Removed unused `office-addin-debugging` dependency, which eliminated the `@microsoft/teamsfx-core` chain responsible for the bulk of the alerts (axios, handlebars, tar, @xmldom/xmldom, fast-xml-parser) along with the orphaned `start`/`stop`/`signin`/`signout` scripts
- Updated Office Add-in tooling to current majors: office-addin-cli ^2.0.9, office-addin-dev-certs ^2.0.9, office-addin-lint ^3.0.9, office-addin-manifest ^2.1.5, office-addin-prettier-config ^2.0.4, eslint-plugin-office-addins ^4.0.9, custom-functions-metadata-plugin ^2.1.9
- Added an `overrides` block forcing `uuid` ^11.1.1 for the remaining dev-only copies pinned by office-addin-manifest and webpack-dev-server's sockjs (GHSA-w5hq-g745-h8pq)

## v0.5.3 - 2026.03.07
### Security
- Fixed high and moderate severity vulnerabilities (CVE-2025-54798, CVE-2025-15284, CVE-2026-23745, CVE-2026-26278, CVE-2025-13465, GHSA-5c6j-r48x-rmvq)
- Updated copy-webpack-plugin to ^14.0.0
- Updated office-addin-debugging to ^6.0.4

## v0.5.2 - 2025.12.07
### Change 
- dependencies updated

## v0.5.1 - 2025.07.21
### Change
- upgraded a couple dependecies with cve's 

## v0.5.0 - 2025.07.07

### Change
- cells that have four repeating characters will not get checked, eg: XXXX or ****

### Add

- data validation from the AllCohorts rule
- move the schedule to the second row of the schedule page
- real excel app in microsoft store
- list number of broken rules somewhere

## v0.4.1 - 2025.05.10

### Change
- upgraded a few packages to address vulnerabilities
- added some dev docs

## v0.4.0 - 2025.03.08

### Added
- OneClassAtATime rule to prevent cohorts from being scheduled for multiple classes at the same time

### Changed
- Replaced manual rule check buttons with auto-check toggle switch
- Added automatic rule validation when data changes in Rules or Schedule sheets
- Added performance timing for run and clear functions, some basic optimization

### Fixed
- Bug where red highlighting would remain on cells after their content was deleted

## v0.3.1 - 2025.03.02

### Fixed
- FINDCLASSCOHORT function

## v0.3.0 - 2025.03.02

### Add
- FINDCLASSCOHORT function and docs
- CohortBlacklist rule and docs

## v0.2.1 - 2025.02.26

### Changed
- made the findcohort function show "-" when there is no class found

### Fixed
- included the schedule range in the FINDCOHORTCLASS function to allow for proper updating

## v0.2.0 - 2025.02.22

### Add
- the FindCohortClass function, and docs

### Changed
- made the panel contents smaller, added version


## v0.1.0 - 2025.02.16

### Added

- the AllCohorts rule, and docs
- the ClassRequiresTravel rule, and docs
- initial website
