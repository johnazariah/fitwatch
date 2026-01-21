# Create Logical Commits from Changes

Analyze the current changes and group them into multiple logical commits.

## Instructions

1. **Analyze changes** using `get_changed_files`

2. **Run formatting first** to avoid formatting noise in commits:
   ```powershell
   cd fitwatch; gofmt -w .
   ```

3. **Identify logical groups** based on:
   - **Feature area**: Same feature or component (e.g., "watcher", "intervals client", "config")
   - **Change type**: Refactoring, bug fixes, documentation, new features
   - **Conventional commit scope**: What prefix/scope makes sense (feat, fix, refactor, docs, style, test, chore)

4. **Common groupings to look for**:
   - Documentation updates (.project/, README, docs/)
   - Code style/formatting changes
   - New feature implementation
   - Bug fixes
   - Test additions
   - Configuration changes (CI, devcontainer, Makefile)
   - Dependency updates (go.mod, go.sum)

5. **For each logical group**, execute:
   ```powershell
   # First unstage everything
   git reset HEAD

   # Stage files for first commit
   git add <file1> <file2> ...
   git commit -m "<type>(<scope>): <description>"

   # Stage files for second commit (skip hooks after first commit passed)
   git add <file3> <file4> ...
   git commit --no-verify -m "<type>(<scope>): <description>"

   # Repeat for remaining groups
   ```

6. **Commit message format** (conventional commits):
   - `feat(scope):` - New feature
   - `fix(scope):` - Bug fix
   - `refactor(scope):` - Code refactoring without behavior change
   - `docs(scope):` - Documentation only
   - `style(scope):` - Formatting, whitespace, etc.
   - `test(scope):` - Adding or updating tests
   - `chore(scope):` - Maintenance, dependencies, build

7. **Order commits logically**:
   - Infrastructure/refactoring first
   - Features second
   - Documentation last

## Example Output

For a mix of changes, you might create:
1. `chore: initial project structure with CI and devcontainer`
2. `feat(watcher): add file system watcher for FIT files`
3. `feat(intervals): add Intervals.icu API client`
4. `test(integration): add real API integration tests`
5. `docs: add specifications and research notes`

## Constraints

- Keep related changes together (don't split a feature across commits)
- Each commit should be atomic and buildable (`go build ./...` must pass)
- Use clear, descriptive commit messages
- If unsure, ask the user which grouping they prefer
- Always add co-authorship attribution for AI-assisted commits

## Co-Authorship

When commits are created with AI assistance, add co-author trailers to the commit message:

```
Co-authored-by: GitHub Copilot <copilot@github.com>
Co-authored-by: Claude <noreply@anthropic.com>
```

These should be added as the last lines of the commit message body, separated by a blank line from the main message.
