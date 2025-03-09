# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Roadmap

### Add

- data validation from the AllCohorts rule
- move the schedule to the second row of the schedule page

## v0.4.0 - Unreleased

### Added
- OneClassAtATime rule to prevent cohorts from being scheduled for multiple classes at the same time

### Changed
- Replaced manual rule check buttons with auto-check toggle switch
- Added automatic rule validation when data changes in Rules or Schedule sheets

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
