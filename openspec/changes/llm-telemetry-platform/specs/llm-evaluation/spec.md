## ADDED Requirements

### Requirement: Create evaluation definitions
The system SHALL allow users to define evaluations with a name, description, and one or more scoring criteria.

#### Scenario: Create an evaluation with rule-based criteria via API
- **WHEN** a client sends `POST /api/v1/evaluations` with criteria using rule-based scorers (regex match, contains, JSON schema validation)
- **THEN** the system stores the evaluation definition and returns it with its ID

#### Scenario: Create an evaluation with LLM-as-judge criteria
- **WHEN** a client sends `POST /api/v1/evaluations` with criteria specifying an LLM-as-judge scorer (provider, model, rubric, scoring scale)
- **THEN** the system stores the evaluation definition including the judge configuration

#### Scenario: Create an evaluation with mixed criteria
- **WHEN** a client creates an evaluation combining rule-based and LLM-as-judge criteria
- **THEN** the system stores all criteria and applies each using its respective scorer type

#### Scenario: List evaluations via API
- **WHEN** a client sends `GET /api/v1/evaluations`
- **THEN** the system returns a paginated list of evaluation definitions for the project

### Requirement: Evaluation datasets
The system SHALL support curated datasets of test cases for systematic and repeatable evaluation.

#### Scenario: Create a dataset via API
- **WHEN** a client sends `POST /api/v1/datasets` with `{"name": "...", "description": "..."}` and a list of test cases, each containing `input` text and optionally `expected_output` text
- **THEN** the system stores the dataset and its test cases under the authenticated project

#### Scenario: Add test cases to an existing dataset
- **WHEN** a client sends `POST /api/v1/datasets/:id/cases` with additional test cases
- **THEN** the system appends the cases to the dataset

#### Scenario: List datasets
- **WHEN** a client sends `GET /api/v1/datasets`
- **THEN** the system returns a paginated list of datasets with name, description, and case count

#### Scenario: Import dataset from traces
- **WHEN** a user selects traces in the trace explorer and clicks "Add to Dataset"
- **THEN** the system creates test cases from the selected traces' LLM input/output content and adds them to the specified dataset

### Requirement: Run evaluations against live traces
The system SHALL execute evaluations against traces matching filter criteria.

#### Scenario: Run evaluation on filtered traces via API
- **WHEN** a client sends `POST /api/v1/evaluations/:id/runs` with `{"source": "traces", "filters": {"start": "...", "end": "...", "prompt_id": "...", "prompt_version": N}}`
- **THEN** the system queues the evaluation run, returns a run ID, and processes matching traces asynchronously

#### Scenario: Scoring uses output_content from spans
- **WHEN** the evaluation runner processes a trace
- **THEN** the system applies each criterion's scorer to the `output_content` text extracted from the trace's LLM span

#### Scenario: Skip traces without content
- **WHEN** a matched trace has no `output_content` (null)
- **THEN** the system skips it and records it as "skipped — no output content" in the run results

### Requirement: Run evaluations against datasets
The system SHALL execute evaluations against curated datasets by calling an LLM with each test case's input.

#### Scenario: Run evaluation on dataset via API
- **WHEN** a client sends `POST /api/v1/evaluations/:id/runs` with `{"source": "dataset", "dataset_id": "..."}`
- **THEN** the system runs each test case through the configured LLM (using the project's LLM provider), applies scorers to the output, and compares against expected output (if provided)

#### Scenario: Dataset run requires LLM provider
- **WHEN** a dataset evaluation run is requested but no LLM provider is configured for the project
- **THEN** the system returns an error indicating LLM provider configuration is required

### Requirement: Scorer execution
The system SHALL support pluggable scorers implementing a common interface.

#### Scenario: Rule-based scorer — regex
- **WHEN** a regex criterion is applied to an LLM output
- **THEN** the system returns pass if the output matches the regex, fail otherwise

#### Scenario: Rule-based scorer — contains
- **WHEN** a contains/not-contains criterion is applied
- **THEN** the system returns pass/fail based on substring presence

#### Scenario: Rule-based scorer — JSON schema
- **WHEN** a JSON schema criterion is applied
- **THEN** the system attempts to parse the output as JSON and validates against the schema

#### Scenario: LLM-as-judge scorer
- **WHEN** an LLM-as-judge criterion is applied
- **THEN** the system sends the output (and expected output if available) with the rubric to the configured judge LLM, parses the numeric score from the response

### Requirement: Evaluation run results
The system SHALL store and display evaluation run results with aggregate statistics.

#### Scenario: Get run status and results via API
- **WHEN** a client sends `GET /api/v1/evaluations/:id/runs/:runId`
- **THEN** the system returns run status (pending, running, complete, failed), progress percentage, per-trace/per-case scores, and aggregate statistics (mean, median, pass rate per criterion)

#### Scenario: View results in UI
- **WHEN** a user views a completed evaluation run
- **THEN** the system displays per-criterion score distributions, aggregate statistics, and a per-item results table with scores and output text

### Requirement: Prompt version comparison via evaluation
The system SHALL support comparing prompt versions by running the same evaluation against traces from different versions.

#### Scenario: Compare two prompt versions
- **WHEN** a user runs the same evaluation against traces from prompt version N and version N+1
- **THEN** the system displays side-by-side aggregate scores showing how the prompt change affected output quality

#### Scenario: A/B comparison via dataset
- **WHEN** a user runs the same evaluation against a dataset using two different prompt versions
- **THEN** the system displays per-case score comparisons showing which version performed better on each test case

### Requirement: Human review workflow (Phase 3)
The system SHALL support human review of LLM outputs with scoring and annotation.

#### Scenario: Submit human review scores
- **WHEN** a reviewer views an LLM output in the UI and submits scores per criterion plus an optional annotation
- **THEN** the system stores the human scores linked to the trace/case and evaluation

#### Scenario: View pending reviews
- **WHEN** a user navigates to the human review queue
- **THEN** the system displays items pending review, ordered by submission time

### Requirement: Evaluation trend tracking
The system SHALL track evaluation scores over time for trend analysis.

#### Scenario: View score trends
- **WHEN** a user views the trend chart for an evaluation definition
- **THEN** the system displays aggregate scores across all runs, showing improvement or regression over time
