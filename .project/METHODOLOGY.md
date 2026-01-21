# Spec-Driven Development Methodology

This document outlines how to structure, design, and implement this project using a spec-driven approach.

---

## Project Structure

```
.project/
├── METHODOLOGY.md          # This file - how we work
├── specs/
│   ├── fit-file-sync-and-analysis.md   # Main product spec
│   ├── features/                        # Detailed feature specs
│   │   ├── 001-local-folder-source.md
│   │   ├── 002-strava-sink.md
│   │   └── ...
│   └── decisions/                       # Architecture Decision Records
│       ├── 001-python-fastapi.md
│       ├── 002-llm-provider.md
│       └── ...
├── designs/
│   ├── architecture.md                  # System architecture details
│   ├── data-model.md                    # Database schema design
│   ├── api/                             # API specs (OpenAPI/schemas)
│   └── ui/                              # Wireframes, user flows
├── research/
│   ├── fit-file-format.md               # Technical research
│   ├── api-integrations/                # Notes on each API
│   │   ├── strava.md
│   │   ├── garmin.md
│   │   └── mywhoosh.md
│   └── llm-analysis-approaches.md
└── planning/
    ├── roadmap.md                       # High-level timeline
    ├── sprints/                         # Sprint planning docs
    └── retrospectives/                  # What we learned
```

---

## The Spec-Driven Workflow

### Phase 1: Discovery & Brainstorming

**Goal**: Understand the problem space deeply before writing code.

1. **Problem Definition** (✅ Done in main spec)
   - Who are the users?
   - What pain points are we solving?
   - What does success look like?

2. **Research Spikes**
   - Create research docs in `.project/research/`
   - Investigate APIs, libraries, and technical constraints
   - Document findings with code snippets and examples
   
   ```markdown
   # Research: MyWhoosh API
   
   ## Findings
   - Authentication: OAuth 2.0 / API key?
   - Endpoints discovered: ...
   - Rate limits: ...
   - Sample response: ...
   
   ## Unknowns
   - [ ] How to get historical activities?
   - [ ] Is there a webhook for new activities?
   
   ## Proof of Concept
   [Link to prototype code or notebook]
   ```

3. **Open Questions Resolution**
   - For each open question, create a decision doc
   - Use Architecture Decision Records (ADRs) format

---

### Phase 2: Design

**Goal**: Create detailed specifications before implementation.

#### 2.1 Feature Specs

For each feature, create a spec in `.project/specs/features/`:

```markdown
# Feature: Local Folder Source

## Summary
Watch a local directory for new .FIT files and import them automatically.

## User Stories
- As a user with a Garmin Edge, I want to drop files into a folder and have them sync

## Acceptance Criteria
- [ ] User can configure one or more watched folders
- [ ] New .FIT files are detected within 30 seconds
- [ ] Files are moved to archive after processing
- [ ] Duplicate files are detected and skipped
- [ ] Errors are logged and user is notified

## Technical Design

### Components
- FileWatcher service using `watchdog` library
- FIT file validator
- Duplicate detector (hash-based)

### Configuration
```yaml
sources:
  local_folders:
    - path: "C:/Users/me/Garmin"
      archive_path: "C:/Users/me/Garmin/processed"
      poll_interval: 30
```

### Error Handling
- Invalid FIT file → move to `errors/` folder
- Permission denied → retry with backoff, then alert

## Test Cases
- [ ] Detects new file in watched folder
- [ ] Ignores non-FIT files
- [ ] Handles corrupt FIT files gracefully
- [ ] Detects duplicate by content hash
- [ ] Recovers after folder is temporarily unavailable

## Dependencies
- FIT file parser must be implemented first
- Storage service for saving imported files

## Estimates
- Design: 2 hours
- Implementation: 4 hours  
- Testing: 2 hours
```

#### 2.2 Architecture Decision Records (ADRs)

For significant technical decisions:

```markdown
# ADR-001: Use Python with FastAPI

## Status
Accepted

## Context
We need to choose a backend language/framework for the sync service.

## Options Considered

### Option A: Python + FastAPI
- Pros: Rich FIT libraries, async support, fast development
- Cons: Slower than compiled languages

### Option B: Node.js + Express
- Pros: Good async, large ecosystem
- Cons: FIT parsing libraries less mature

### Option C: Go
- Pros: Fast, single binary deployment
- Cons: Fewer FIT libraries, slower development

## Decision
Python + FastAPI

## Rationale
- `fitdecode` library is well-maintained and handles edge cases
- FastAPI provides automatic OpenAPI docs
- Team familiarity with Python
- Performance is acceptable for our scale

## Consequences
- Need to handle Python deployment (Docker recommended)
- May need to optimize hot paths later
```

---

### Phase 3: Implementation

**Goal**: Build features incrementally, always referring back to specs.

#### 3.1 Implementation Workflow

```
For each feature:
1. Review the feature spec
2. Create a feature branch
3. Write failing tests based on acceptance criteria
4. Implement until tests pass
5. Update spec if design changed during implementation
6. Code review against spec
7. Merge and update spec status
```

#### 3.2 Vertical Slices

Build complete vertical slices, not horizontal layers:

```
❌ Wrong: Build all adapters, then all services, then all UI
✅ Right: Build one complete flow end-to-end, then the next

Example vertical slice:
"User drops FIT file in folder → file is parsed → uploaded to Strava → user sees success"

This proves the architecture works before building more features.
```

#### 3.3 Spec-Code Traceability

Keep specs and code linked:

```python
# src/sources/local_folder.py

class LocalFolderSource:
    """
    Watch a local directory for new .FIT files.
    
    Spec: .project/specs/features/001-local-folder-source.md
    """
```

---

### Phase 4: Testing

**Goal**: Tests should verify the spec, not just the implementation.

#### 4.1 Test Structure Mirrors Specs

```
tests/
├── unit/                    # Test individual components
├── integration/             # Test component interactions
├── e2e/                     # Test full user flows
└── acceptance/              # Tests derived directly from specs
    ├── test_local_folder_source.py
    ├── test_strava_sink.py
    └── ...
```

#### 4.2 Acceptance Tests from Specs

Convert acceptance criteria directly to tests:

```python
# tests/acceptance/test_local_folder_source.py
"""
Tests for: .project/specs/features/001-local-folder-source.md
"""

class TestLocalFolderSource:
    
    def test_detects_new_fit_file_in_watched_folder(self):
        """AC: New .FIT files are detected within 30 seconds"""
        # Arrange
        watcher = LocalFolderSource(watch_path="/tmp/test")
        
        # Act
        copy_fit_file_to("/tmp/test/activity.fit")
        detected = watcher.wait_for_file(timeout=30)
        
        # Assert
        assert detected is not None
        assert detected.name == "activity.fit"
    
    def test_ignores_non_fit_files(self):
        """AC: Files are validated before processing"""
        ...
    
    def test_skips_duplicate_files(self):
        """AC: Duplicate files are detected and skipped"""
        ...
```

---

## Practical Next Steps

### Recommended Order

1. **Research Phase** (1-2 days)
   - [ ] Research MyWhoosh API/download methods
   - [ ] Research Strava upload API
   - [ ] Test FIT file parsing libraries
   - [ ] Document findings in `.project/research/`

2. **Core Design** (1 day)
   - [ ] Finalize data model design
   - [ ] Write ADR for key decisions
   - [ ] Create first 3 feature specs (local source, Strava sink, FIT parser)

3. **MVP Skeleton** (1-2 days)
   - [ ] Set up project structure (Python, FastAPI, tests)
   - [ ] Implement first vertical slice
   - [ ] Prove the architecture works

4. **Feature Build-Out**
   - [ ] Implement features per spec priority
   - [ ] Write tests as you go
   - [ ] Update specs with learnings

---

## Tools & Templates

### Brainstorming Sessions

Use this template when exploring a topic:

```markdown
# Brainstorm: [Topic]

## Date: YYYY-MM-DD

## Participants
- [Names or just "solo"]

## Questions to Answer
1. ...
2. ...

## Ideas Generated
- Idea 1: ...
- Idea 2: ...

## Decisions Made
- We will... because...

## Action Items
- [ ] Research X
- [ ] Write spec for Y
- [ ] Prototype Z

## Open Questions Remaining
- ?
```

### Quick Prototyping

For technical spikes, create throwaway code:

```
.project/
└── prototypes/
    ├── fit-parsing-spike/
    ├── strava-auth-test/
    └── llm-summary-experiment/
```

These are NOT production code - they're learning tools.

---

## Anti-Patterns to Avoid

1. **Analysis Paralysis**: Don't spec everything upfront. Spec → Build → Learn → Spec more.

2. **Spec Drift**: If implementation differs from spec, UPDATE THE SPEC. Dead specs are worse than no specs.

3. **Over-Engineering**: Start simple. You can always add complexity later.

4. **Skipping Research**: 1 hour of research can save 10 hours of wrong implementation.

5. **Big Bang Integration**: Test integrations early with real APIs, not mocks.

---

## Definition of Done

A feature is "done" when:

- [ ] Implementation matches the spec
- [ ] Acceptance tests pass
- [ ] Spec is updated if design changed
- [ ] Code is reviewed
- [ ] Documentation is updated
- [ ] Feature works in a realistic environment
