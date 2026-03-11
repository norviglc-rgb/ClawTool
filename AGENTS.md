# AGENTS.md

## 1. Purpose and authority

This file is the authoritative development specification for the `clawtool` repository.
Codex must follow this document when planning, implementing, testing, and refactoring code.
If the codebase and this file disagree, update the codebase to match this file unless there is a clear bug in the specification. If the specification is incomplete, choose the smallest safe implementation that preserves the architecture, automation, i18n, and test requirements below.

`clawtool` is a CLI-first, cross-platform control plane for OpenClaw. It is not a GUI product and not a one-off installer. Its job is to standardize installation, configuration, verification, diagnostics, repair, and remote operations.

## 2. Product goals

Build a maintainable Go CLI that:

1. Detects local environment and installation state.
2. Installs and configures OpenClaw in a repeatable way.
3. Plans changes before applying them.
4. Verifies the final state after apply.
5. Backs up and rolls back configuration safely.
6. Collects logs and performs deterministic diagnostics.
7. Supports internationalized user-facing output from day one.
8. Ships with full automation for linting, testing, release, and documentation.
9. Supports automated tests at unit, integration, golden, and end-to-end levels.

## 3. Non-goals

Do not build these in the initial implementation unless explicitly required by a later milestone:

- GUI or web dashboard
- Cloud synchronization service
- Fully autonomous AI self-healing
- Bulk fleet orchestration beyond SSH-based single-host remote operations
- Proprietary vendor lock-in around message catalogs or templates

## 4. Primary deliverable

The first production-quality deliverable is a local-first CLI with the following stable command chain:

- `clawtool detect`
- `clawtool doctor`
- `clawtool init`
- `clawtool profile`
- `clawtool plan`
- `clawtool apply`
- `clawtool verify`
- `clawtool inspect`
- `clawtool status`
- `clawtool logs`
- `clawtool rollback`
- `clawtool repair`

Remote and diagnostics extensions come after the local lifecycle is stable.

## 5. Technology and architectural decisions

### 5.1 Language and runtime

- Use Go for all core implementation.
- Keep the project buildable as a single native binary per target platform.
- Prefer the Go standard library where reasonable.

### 5.2 Recommended libraries

Use these unless there is a strong technical reason not to:

- `cobra` for CLI structure and help generation
- `pflag` through Cobra for flags
- `go-i18n/v2` for localization
- `yaml.v3` for YAML parsing
- `x/crypto/ssh` for SSH support
- `log/slog` for structured logs

Avoid excessive framework usage. Keep the architecture explicit and testable.

### 5.3 Output model

Separate execution results from user-facing rendering.

Rules:

- Core logic must return typed result structs and typed error codes.
- Rendering must be handled by a presentation layer.
- Human-readable output must be localized.
- JSON output must stay stable, machine-readable, and locale-neutral.
- Never hard-code human-visible strings in business logic.

### 5.4 Idempotency and safety

- `apply` must be idempotent.
- Configuration writes must create backups before mutation.
- Destructive operations must require explicit confirmation flags in non-interactive mode or be rollback-safe.
- Failure must preserve enough state to debug and recover.

## 6. Supported platforms and scope by phase

### 6.1 Local execution

Target local support for:

- macOS
- Linux
- Windows

### 6.2 Remote execution

Phase 3 remote support:

- SSH to Linux and macOS hosts
- Windows remote support is optional and may be deferred

### 6.3 Architecture rule

Platform-specific behavior must live behind adapters. Do not scatter OS checks throughout business logic.

## 7. Repository structure

Use this structure unless there is a compelling reason to adjust it:

```text
.
  cmd/
    clawtool/
      main.go
  internal/
    app/
    cli/
    core/
    config/
    platform/
      common/
      darwin/
      linux/
      windows/
    install/
    verify/
    backup/
    state/
    diff/
    logs/
    diagnostics/
    remote/
    i18n/
    render/
    schema/
    testutil/
  locales/
    en.json
    zh-CN.json
    ja.json
  templates/
    default.yaml
    local.yaml
    remote-ssh.yaml
  schemas/
    manifest.schema.json
    profile.schema.json
  docs/
    architecture.md
    cli/
  scripts/
  testdata/
  .github/
    workflows/
  go.mod
  go.sum
  AGENTS.md
  README.md
```

## 8. Command scope and acceptance criteria

### 8.1 `detect`

Purpose:

- Collect platform facts.
- Detect whether OpenClaw is installed.
- Detect executable path, config path, writable directories, shell, and relevant dependencies.

Acceptance criteria:

- Works on macOS, Linux, and Windows.
- Supports human output and `--json`.
- Returns a typed result model with stable JSON keys.

### 8.2 `doctor`

Purpose:

- Run health checks and classify issues as `PASS`, `WARN`, or `FAIL`.

Acceptance criteria:

- Checks include dependency presence, config readability, write permissions, installation completeness, and path sanity.
- Each finding has a stable code, severity, message key, and remediation hint key.
- Supports JSON output with deterministic ordering.

### 8.3 `init`

Purpose:

- Initialize local project state and default configuration workspace.

Acceptance criteria:

- Creates a predictable working directory such as `.clawtool/` with subdirectories for profiles, backups, cache, state, and logs.
- Writes default config and sample profile files.
- Is safe to run multiple times.

### 8.4 `profile`

Purpose:

- Manage named execution profiles.

Required subcommands:

- `list`
- `show`
- `create`
- `validate`
- `use`

Acceptance criteria:

- Profiles are YAML-based.
- Validation uses JSON Schema or equivalent explicit schema validation.
- Profiles support local and remote target definitions.

### 8.5 `plan`

Purpose:

- Resolve current state, profile, templates, and overrides into a diffable execution plan.

Acceptance criteria:

- Outputs which files will change, which install steps will run, whether backup is required, and what verification checks will run.
- Produces deterministic output.
- Supports `--json` and human-readable diff summaries.
- Must not mutate the environment.

### 8.6 `apply`

Purpose:

- Execute a previously computed or inline-generated plan.

Acceptance criteria:

- Performs pre-checks.
- Creates backups.
- Installs or updates OpenClaw where required.
- Writes resolved configuration.
- Records state.
- Runs verification.
- Returns actionable failures.
- Re-running the same apply should converge without unnecessary changes.

### 8.7 `verify`

Purpose:

- Validate that the installed and configured state matches expectations.

Acceptance criteria:

- Checks executable availability, configuration validity, required paths, and profile-specific constraints.
- Returns non-zero exit status on failure.
- Produces stable JSON for CI usage.

### 8.8 `inspect` and `status`

Purpose:

- Show current effective state and the last known lifecycle information.

Acceptance criteria:

- Surface current profile, install path, config path, last apply time, last apply result, and backup availability.
- `status` is compact; `inspect` is detailed.

### 8.9 `logs`

Purpose:

- Surface recent local lifecycle logs and optionally bundle them.

Acceptance criteria:

- Supports `--tail`, `--since`, and `--bundle`.
- Bundles logs, current state summary, and recent failure metadata into a zip or tar archive.

### 8.10 `rollback`

Purpose:

- Restore the most recent or selected backup.

Acceptance criteria:

- Refuses to run without an available backup.
- Updates state after rollback.
- Verifies the rolled-back state.

### 8.11 `repair`

Purpose:

- Perform deterministic repair actions for known problems.

Acceptance criteria:

- Uses explicit rules first.
- Never hides destructive actions.
- Can suggest but not auto-execute risky remediations without an explicit flag.

### 8.12 Phase 3+ commands

Add after the local lifecycle is stable:

- `remote exec`
- `remote plan`
- `remote apply`
- `remote verify`
- `diagnose`
- `mcp` (optional future integration surface)

## 9. Data model requirements

Implement explicit typed models for:

- `Profile`
- `Manifest`
- `Plan`
- `PlanStep`
- `BackupRecord`
- `StateRecord`
- `DoctorFinding`
- `VerifyFinding`
- `RepairAction`
- `LogBundleMetadata`

Rules:

- Avoid untyped maps in core logic.
- Prefer explicit version fields on persisted data.
- Store state in JSON for easy automation and inspection.
- Persist schemas under `schemas/`.

## 10. Configuration and file format rules

- Use YAML for profiles and templates.
- Use JSON for state snapshots and machine-readable output.
- Use schema validation for profile and manifest files.
- Include sample files under `templates/` and `testdata/`.
- Use Go `embed` for default templates, default schemas, and locales so the binary is self-contained.

## 11. i18n requirements

Internationalization is mandatory from the first commit. Do not defer it.

### 11.1 Supported locales

Implement these locales from the start:

- `en`
- `zh-CN`
- `ja`

English is the fallback locale.

### 11.2 Locale resolution order

Resolve locale in this order:

1. `--lang`
2. `CLAWTOOL_LANG`
3. config file setting
4. system locale
5. fallback `en`

### 11.3 Message design

Rules:

- Every user-visible string must come from a message catalog.
- Message keys must be stable and namespaced, for example `doctor.summary.fail_count`.
- Use placeholders for dynamic values.
- Error codes must be stable and separate from localized text.
- JSON field names must remain English and stable across locales.
- Keep log event names stable and English; only user-facing rendered summaries are localized.

### 11.4 i18n quality gates

CI must fail if:

- A message key exists in `en` but is missing in `zh-CN` or `ja`.
- A placeholder set is inconsistent across locales.
- A command introduces a hard-coded user-facing string outside the i18n layer.

### 11.5 Test expectations for i18n

Add tests for:

- locale resolution order
- fallback behavior
- placeholder rendering
- completeness of catalogs
- sample CLI golden outputs for `en`, `zh-CN`, and `ja`

## 12. Automation requirements

Automation is part of the product. Do not leave it for later.

### 12.1 Local developer automation

Provide the following commands via `Makefile` and PowerShell equivalents under `scripts/`:

- `make fmt`
- `make lint`
- `make test`
- `make test-unit`
- `make test-integration`
- `make test-e2e`
- `make test-i18n`
- `make build`
- `make build-all`
- `make generate`
- `make check`
- `make release-dry-run`

Windows users must have matching entry points, for example `scripts/dev.ps1 check`, so local automation is not POSIX-only.

### 12.2 CI automation

Create GitHub Actions workflows for:

1. `ci.yml`
   - run on pull requests and pushes
   - format verification
   - lint
   - unit tests
   - integration tests
   - i18n checks
   - coverage upload or artifact generation

2. `cross-platform.yml`
   - smoke test build and CLI help on Linux, macOS, and Windows

3. `release.yml`
   - tag-driven release build
   - checksum generation
   - artifact packaging

4. `nightly-e2e.yml`
   - optional nightly heavier end-to-end tests

### 12.3 Dependency and maintenance automation

Add:

- Dependabot or Renovate configuration
- `.editorconfig`
- `.gitattributes`
- generated CLI docs under `docs/cli/`
- automatic doc generation if supported by the CLI framework

## 13. Testing strategy

Automated testing is mandatory for each feature.

### 13.1 Test layers

Implement all of these:

1. Unit tests
   - pure business logic
   - diff generation
   - schema validation
   - i18n utilities
   - renderers

2. Integration tests
   - filesystem operations in temp directories
   - backup and rollback behavior
   - plan/apply/verify lifecycle with fake installers and fake platform adapters

3. Golden tests
   - human CLI output
   - localized output in all supported locales
   - JSON snapshots for stable machine output

4. End-to-end tests
   - invoke the built binary
   - validate exit codes
   - validate local lifecycle commands

5. Remote integration tests
   - for Phase 3, use a containerized SSH server in CI where possible
   - keep these isolated and skippable for local contributors

### 13.2 Coverage targets

Minimum expectations:

- overall line coverage >= 80%
- `internal/core`, `internal/config`, `internal/i18n`, and `internal/render` >= 85%
- every command package must have at least a smoke test

If a package is hard to cover, add focused tests or refactor the design.

### 13.3 Mandatory checks before a change is considered done

Codex must run, and the repository must pass:

- formatting
- linting
- unit tests
- integration tests
- i18n validation
- command smoke tests

If end-to-end tests are intentionally skipped, state exactly why.

## 14. Logging and diagnostics

### 14.1 Logging design

- Use structured logs with stable event names.
- Separate machine logs from human-readable CLI summaries.
- Logs must include timestamp, command, step, result, and error code when relevant.
- Redact secrets and tokens.

### 14.2 Deterministic diagnostics first

Before any AI-based analysis exists, implement a deterministic diagnostics engine that:

- matches known errors by code or pattern
- suggests specific next commands
- classifies severity
- can serialize findings for future AI augmentation

### 14.3 AI integration boundary

If a future `diagnose` command is implemented, keep the provider behind an interface. Do not hard-wire external network calls into core logic. Tests must run offline by default.

## 15. State, backup, and rollback rules

- Persist state after every successful apply and rollback.
- Keep backup metadata index in state.
- Give every backup a timestamp and profile name.
- Rollback must be reversible where possible by creating a pre-rollback backup.
- Surface backup inventory in `inspect`.

## 16. UX and CLI behavior rules

- Every command must support `--json` where practical.
- Exit codes must be deterministic.
- Help text must be concise and localized where supported.
- Use stable command names and avoid renaming after release.
- Interactive prompts must have `--yes` or equivalent non-interactive bypass for automation.
- Do not print stack traces to normal users unless `--debug` is enabled.

## 17. Security and safety rules

- Never log secrets.
- Validate all external inputs.
- Normalize and sanitize filesystem paths.
- Avoid shelling out when native Go functionality is enough.
- When shelling out is required, use explicit arguments and no unsafe string interpolation.
- Remote operations must support host key verification strategy configuration.

## 18. Development phases and execution order

Implement in this order.

### Phase 0: foundation

Deliverables:

- repository skeleton
- root command and subcommand scaffolding
- i18n layer and locale catalogs
- rendering abstraction
- typed error codes
- state and schema packages
- Makefile and PowerShell automation
- GitHub Actions CI
- initial docs generation

Exit criteria:

- `clawtool --help` works
- locale switching works
- CI passes on a minimal skeleton

### Phase 1: local MVP lifecycle

Deliverables:

- `detect`
- `doctor`
- `init`
- `profile` basic operations
- `plan`
- `apply`
- `verify`
- minimal templates and sample profiles

Exit criteria:

- local lifecycle works end to end in temp directories
- idempotent apply verified by tests
- JSON and localized human outputs both tested

### Phase 2: operational completeness

Deliverables:

- `inspect`
- `status`
- `logs`
- `rollback`
- `repair`
- richer state and backup management

Exit criteria:

- failed apply leaves usable backups
- rollback works in automated tests
- log bundle generation tested

### Phase 3: remote support

Deliverables:

- SSH executor
- remote profile support
- `remote exec`
- `remote plan`
- `remote apply`
- `remote verify`

Exit criteria:

- remote lifecycle tested against CI SSH target
- local and remote plans share the same core plan model

### Phase 4: diagnostics extension

Deliverables:

- deterministic `diagnose`
- provider interface for optional AI analysis
- optional `mcp` scaffolding if needed later

Exit criteria:

- deterministic diagnosis works offline
- no network dependency in default test pipeline

## 19. Definition of done

A task is not done unless all of the following are true:

1. The implementation matches this specification.
2. All new user-facing strings use i18n catalogs.
3. Tests are added or updated.
4. Automation commands continue to pass.
5. CLI docs or help output are updated if behavior changed.
6. JSON outputs remain stable or changes are explicitly documented.
7. Safety and rollback behavior are preserved.

## 20. Coding rules for Codex

Follow these working rules:

- Make small, reviewable changes.
- Prefer explicit interfaces over magic global state.
- Keep platform logic behind adapters.
- Keep business logic independent from Cobra command construction.
- Do not postpone tests.
- Do not introduce hard-coded English strings in command handlers.
- Do not add TODO comments unless they include a concrete issue reference and rationale.
- Do not add dead scaffolding that is not connected to current phases.
- When uncertain, choose the simplest architecture that preserves extensibility for remote execution, i18n, and diagnostics.

## 21. Expected initial file set

Codex should create and wire at least these files early:

```text
cmd/clawtool/main.go
internal/cli/root.go
internal/cli/detect.go
internal/cli/doctor.go
internal/cli/init.go
internal/cli/profile.go
internal/cli/plan.go
internal/cli/apply.go
internal/cli/verify.go
internal/i18n/loader.go
internal/i18n/resolver.go
internal/render/human.go
internal/render/json.go
internal/core/types.go
internal/core/errors.go
internal/state/store.go
internal/schema/validate.go
locales/en.json
locales/zh-CN.json
locales/ja.json
schemas/profile.schema.json
schemas/manifest.schema.json
templates/default.yaml
Makefile
scripts/dev.ps1
.github/workflows/ci.yml
README.md
```

## 22. Required documentation outputs

Generate and maintain:

- `README.md` with quickstart
- `docs/architecture.md`
- generated CLI command docs under `docs/cli/`
- sample profiles under `testdata/` or `templates/`

## 23. Final instruction to Codex

Implement the repository according to this file in phase order. Do not skip the i18n foundation. Do not skip automation. Do not skip tests. Build the smallest complete system that satisfies the requirements, then iterate phase by phase until the local lifecycle is production-ready.
