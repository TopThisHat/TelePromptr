## ADDED Requirements

### Requirement: Create and store prompt templates
The system SHALL allow users to create prompt templates with a name, description, and template body supporting variable interpolation.

#### Scenario: Create a new prompt via API
- **WHEN** a client sends `POST /api/v1/prompts` with `{"name": "...", "description": "...", "body": "Hello {{name}}, ..."}`
- **THEN** the system creates the prompt with version 1 in "draft" status, stores it under the authenticated project, and returns the created resource with its ID

#### Scenario: Reject duplicate prompt names within a project
- **WHEN** a client attempts to create a prompt with a name that already exists in the project
- **THEN** the system returns HTTP 409 Conflict

#### Scenario: List prompts via API
- **WHEN** a client sends `GET /api/v1/prompts`
- **THEN** the system returns a paginated list of prompts with their latest version info and tags

### Requirement: Prompt versioning
The system SHALL maintain an immutable version history for each prompt template.

#### Scenario: Create a new version via API
- **WHEN** a client sends `POST /api/v1/prompts/:id/versions` with an updated template body
- **THEN** the system creates a new version (incrementing version number) in "draft" status without modifying previous versions

#### Scenario: View version history
- **WHEN** a client sends `GET /api/v1/prompts/:id` or a user views a prompt's detail page
- **THEN** the system returns all versions with version number, creation timestamp, status, and a diff from the previous version

#### Scenario: View a specific version
- **WHEN** a client sends `GET /api/v1/prompts/:id/versions/:version`
- **THEN** the system returns the full template body and metadata for that version

### Requirement: Prompt lifecycle management
The system SHALL enforce a deployment lifecycle for prompt versions: draft -> active -> archived.

#### Scenario: Activate a draft prompt version
- **WHEN** a user or client marks a draft prompt version as "active"
- **THEN** the system sets this version to "active" and any previously active version transitions to "archived"

#### Scenario: Archive a prompt version
- **WHEN** a user or client archives an active prompt version
- **THEN** the version status becomes "archived" and the prompt has no active version until another is activated

#### Scenario: Prevent editing non-draft versions
- **WHEN** a user attempts to modify an active or archived prompt version
- **THEN** the system rejects the modification with an error suggesting to create a new version

### Requirement: Prompt tagging
The system SHALL support user-defined tags on prompt templates.

#### Scenario: Add tags to a prompt
- **WHEN** a user adds tags to a prompt template
- **THEN** the tags are stored and the prompt appears when filtering by those tags

#### Scenario: Filter prompts by tag
- **WHEN** a client sends `GET /api/v1/prompts?tags=production,chat`
- **THEN** the system returns only prompts that have all specified tags

### Requirement: Prompt rendering with trace linkage
The system SHALL render prompt templates by substituting variables and return tracking identifiers for trace correlation.

#### Scenario: Render a prompt with all variables provided
- **WHEN** a client sends `POST /api/v1/prompts/:id/render` with `{"variables": {"name": "Alice"}}`
- **THEN** the system renders the active version's template, generates a `prompt_execution_id` (UUID), and returns `{"rendered": "Hello Alice, ...", "prompt_execution_id": "uuid", "prompt_id": "...", "version": N}`

#### Scenario: Render with missing variables
- **WHEN** a client requests rendering without providing values for all `{{variable}}` placeholders
- **THEN** the system returns HTTP 400 with an error listing the missing variable names

#### Scenario: Render a specific version
- **WHEN** a client sends `POST /api/v1/prompts/:id/versions/:version/render` with variables
- **THEN** the system renders the specified version (regardless of its lifecycle status) and returns the rendered text with tracking identifiers

### Requirement: Passive prompt management
The system SHALL support creating and versioning prompts without requiring the render API in the application's hot path.

#### Scenario: Store prompts for record-keeping only
- **WHEN** a user creates and versions prompts in TelePromptr but does not call the render API from their application
- **THEN** the system stores the prompts normally — the render API and trace linkage are optional, not required

#### Scenario: Manual trace linkage without render API
- **WHEN** an application sets `telepromptr.prompt.id` and `telepromptr.prompt.version` span attributes without calling the render API
- **THEN** the system links traces to prompts based on those attributes

### Requirement: Prompt-linked trace queries
The system SHALL support querying traces by prompt ID and version.

#### Scenario: Find traces for a prompt version via API
- **WHEN** a client sends `GET /api/v1/traces?prompt_id=xyz&prompt_version=3`
- **THEN** the system returns all traces containing spans linked to that prompt version

#### Scenario: View linked traces from prompt detail page
- **WHEN** a user views a prompt version's detail page in the UI
- **THEN** the system displays a link to the trace explorer pre-filtered by that prompt version, with count of matching traces

### Requirement: Prompt playground (Phase 2)
The system SHALL provide a UI playground for testing prompts against LLM providers without deploying application code.

#### Scenario: Test a prompt in the playground
- **WHEN** a user selects a prompt version in the UI, fills in variable values, selects a configured LLM provider and model, and clicks "Run"
- **THEN** the system renders the prompt, calls the selected LLM, and displays the response inline with token count and cost information

#### Scenario: Playground requires LLM provider configuration
- **WHEN** a user attempts to use the playground but no LLM providers are configured for the project
- **THEN** the system displays a message directing them to project settings to configure a provider
