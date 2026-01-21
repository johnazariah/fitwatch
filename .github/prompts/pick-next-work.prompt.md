# Pick Next Work

You are helping the user decide what to work on next. Analyze the project state and make a prioritized recommendation.

**IMPORTANT**: This prompt is for DISCUSSION and PLANNING only. Do NOT start implementing anything until the user explicitly says to proceed (e.g., "let's do #1", "start on X", "go ahead").

## Step 1: Gather Context

Read the following to understand current state:

1. **Specifications** - Read `.project/specifications/`:
   - Understand the overall architecture and design goals
   - Note which specifications are implemented vs. planned
   - Pay attention to dependencies between specifications

2. **Execution Plan** - Read `.project/plan/execution-plan.md`:
   - Understand the phased approach and priorities
   - Identify what phase we're currently in
   - Note any milestones or deadlines

3. **Recent Git History** - Check momentum and recent work:
   ```bash
   git log --oneline -20
   ```

4. **Check for tagged TODOs** in source code:
   ```bash
   grep -rn "TODO\|FIXME\|XXX\|HACK" fitwatch/ --include="*.go" | head -20
   ```

5. **Project Health** - Run checks:
   ```bash
   cd fitwatch
   go test -short ./... 2>&1 | tail -10
   go vet ./... 2>&1 | tail -10
   golangci-lint run 2>&1 | tail -10
   ```

## Step 2: Categorize Work Items

Group findings into:

| Category | Priority | Examples |
|----------|----------|----------|
| **Blocking** | 🔴 High | Failing tests, type errors, broken imports |
| **Momentum** | 🟡 Medium | Continue partially-done specs, quick wins |
| **Strategic** | 🟢 Normal | New features aligned with spec sequence |
| **Debt** | ⚪ Low | Refactoring, TODOs, documentation |

## Step 3: Evaluate Specifications

For each specification in `.project/specifications/`, assess:

1. **Implementation Status** - Fully implemented, partial, or not started?
2. **Dependencies** - Does it require other specs first?
3. **Effort** - Small (hours), Medium (day), Large (days)?
4. **Impact** - How critical is it to the core functionality?

## Step 4: Make Recommendation

Suggest the **top 3** work items with reasoning:

```markdown
## Recommended Next Work

### 1. [Feature/Item Name]
- **Why**: Brief justification
- **Spec**: Link to relevant specification
- **Effort**: Small/Medium/Large
- **First Step**: Concrete action to start

### 2. [Alternative Option]
- **Why**: ...
- **Effort**: ...

### 3. [Another Option]
- **Why**: ...
- **Effort**: ...

## Also Consider
- Quick wins that can be done in parallel
- Debt items worth addressing opportunistically
```

## Step 5: Check Prerequisites

Before starting the recommended work:

1. Is the specification clear and complete?
2. Are there dependencies to address first?
3. Is the scope well-defined?
4. Are there existing tests that define expected behavior?

## Output Format

Present findings as:

1. **Current State Summary** - One paragraph on project health and momentum
2. **Work Items Found** - Categorized list with priorities
3. **Recommendation** - Top 3 prioritized items with reasoning
4. **Ready to Start?** - Confirmation or prerequisites needed

## STOP HERE

After presenting your recommendation, **STOP and wait for user input**.

Do NOT:
- Create files
- Write code
- Start implementation
- Run commands (except for gathering context in Steps 1-3)

DO:
- Present your analysis clearly
- Answer questions about the options
- Discuss trade-offs if asked
- Refine recommendations based on feedback

Only proceed with implementation when the user explicitly instructs you to start (e.g., "let's do #1", "start working on X", "go ahead with the recommendation").
