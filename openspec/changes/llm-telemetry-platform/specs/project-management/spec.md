## ADDED Requirements

### Requirement: Admin-mode web UI access
The system SHALL provide authenticated admin access to the web UI using a shared admin token, without requiring user login.

#### Scenario: SvelteKit server-side access
- **WHEN** a user navigates to any page in the web UI
- **THEN** SvelteKit's server-side load functions call the Go API using the admin token (from `TELEPROMPTR_ADMIN_TOKEN` env var) and the `X-Admin-Token` header, with `X-Project-Id` set to the currently selected project

#### Scenario: Admin token auto-generation
- **WHEN** the Go API starts without a `TELEPROMPTR_ADMIN_TOKEN` environment variable
- **THEN** the system generates a random admin token, prints it to stdout, and uses it for the session

#### Scenario: No browser-to-API auth secrets
- **WHEN** a user interacts with the web UI
- **THEN** the browser communicates only with SvelteKit — all Go API calls happen server-to-server, and no admin token or API key is exposed to the browser

#### Scenario: Admin token grants cross-project access
- **WHEN** a server-side request includes a valid `X-Admin-Token` and `X-Project-Id`
- **THEN** the system returns data for the specified project regardless of API key scoping

### Requirement: Create and manage projects
The system SHALL allow the admin to create projects that serve as the top-level data isolation boundary.

#### Scenario: Create a new project via UI
- **WHEN** the admin submits a project creation form with a name and optional description
- **THEN** the system creates the project and it becomes available for API key generation and data ingestion

#### Scenario: Create a new project via admin API
- **WHEN** a request with a valid admin token sends `POST /api/v1/projects`
- **THEN** the system creates the project and returns it with its ID

#### Scenario: List projects
- **WHEN** the admin views the project list
- **THEN** the system returns all projects with name, description, creation date, and trace count

#### Scenario: Update project metadata
- **WHEN** the admin updates a project's name or description
- **THEN** the system persists the changes

### Requirement: API key management with SHA-256
The system SHALL provide API key generation and management scoped to projects, using SHA-256 for key storage.

#### Scenario: Generate an API key
- **WHEN** the admin requests a new API key for a project
- **THEN** the system generates a key (`tp_proj_` + 32 random hex bytes), displays it once in full, and stores `sha256(key)` plus the key prefix for future lookup

#### Scenario: List API keys for a project
- **WHEN** the admin views API keys for a project
- **THEN** the system returns a list showing key prefix (first 8 chars after `tp_proj_`), creation date, last used date, and active/revoked status

#### Scenario: Revoke an API key
- **WHEN** the admin revokes an API key
- **THEN** the key is marked as revoked and all subsequent requests using it are rejected

### Requirement: Project data isolation
The system SHALL enforce strict data isolation between projects for all API-key-authenticated requests.

#### Scenario: Trace data isolation
- **WHEN** traces are ingested under project A's API key
- **THEN** traces are only visible within project A

#### Scenario: Cross-project access prevention via API key
- **WHEN** an API client attempts to access resources belonging to a different project
- **THEN** the system returns HTTP 403 Forbidden

### Requirement: Project settings — cost-per-token rates
The system SHALL allow configuring cost-per-token rates per model per project.

#### Scenario: Configure cost-per-token rates
- **WHEN** the admin sets cost rates for a model (e.g., `{"model": "gpt-4", "input_cost_per_token": 0.00003, "output_cost_per_token": 0.00006}`)
- **THEN** the system uses these rates for cost computation on future ingested traces

#### Scenario: Default cost rates
- **WHEN** no custom rates are configured for a model
- **THEN** the system stores null cost for that model's traces (cost is unknown, not zero)

### Requirement: Project settings — model display names
The system SHALL allow configuring display name aliases for model identifiers.

#### Scenario: Configure model display names
- **WHEN** the admin sets a display name for a model identifier (e.g., `gpt-4-0125-preview` -> `GPT-4 Turbo`)
- **THEN** the system uses the display name in UI and API responses for that project

### Requirement: LLM provider configuration (Phase 2)
The system SHALL allow configuring LLM provider credentials per project for use by the evaluation framework and prompt playground.

#### Scenario: Add an LLM provider
- **WHEN** the admin configures an LLM provider with `{"provider": "openai", "api_key": "sk-...", "models": ["gpt-4", "gpt-3.5-turbo"]}`
- **THEN** the system encrypts the API key with AES-256-GCM (using `TELEPROMPTR_ENCRYPTION_KEY`), stores the configuration, and makes the provider available for evaluation and playground use

#### Scenario: List configured providers
- **WHEN** the admin views LLM provider settings
- **THEN** the system displays providers with display name, provider type, available models, and creation date — without exposing the API key

#### Scenario: Remove an LLM provider
- **WHEN** the admin removes a configured provider
- **THEN** the system deletes the provider configuration and any evaluations using it are updated to indicate a missing provider

#### Scenario: Custom/self-hosted provider
- **WHEN** the admin configures a provider with `{"provider": "custom", "base_url": "https://my-llm.internal/v1", "api_key": "..."}`
- **THEN** the system stores the custom base URL and uses it for API calls to that provider

### Requirement: Project settings — rate limits
The system SHALL allow configuring ingestion rate limits per project.

#### Scenario: Configure spans-per-second limit
- **WHEN** the admin sets a rate limit for a project (e.g., 10,000 spans/second)
- **THEN** the ingestion pipeline enforces this limit, returning throttling responses when exceeded

#### Scenario: No rate limit configured
- **WHEN** no rate limit is set for a project
- **THEN** the system applies no per-project rate limiting (global buffer limits still apply)
