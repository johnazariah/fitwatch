# Prepare and Execute a Go Binary Release

Analyze changes since the last release, prepare documentation, validate quality, and execute the release ceremony.

---

## Phase 1: Analyze Release Scope

1. **Get the current version** from git tags:
   ```bash
   git describe --tags --abbrev=0 2>/dev/null || echo "No tags yet"
   ```

2. **List changes since last release**:
   ```bash
   LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
   if [ -n "$LAST_TAG" ]; then
       git log "$LAST_TAG..HEAD" --oneline
   else
       git log --oneline -20
   fi
   ```

3. **Categorize changes** and determine version bump:

   | Change Type | Version Bump | Examples |
   |-------------|--------------|----------|
   | Breaking CLI/config changes | **MAJOR** (X.0.0) | Removed flags, changed config format |
   | New features | **MINOR** (0.X.0) | New upload targets, new watch modes |
   | Bug fixes, docs, tests | **PATCH** (0.0.X) | Fixed edge cases, improved logging |

4. **Report recommendation** to user:
   ```
   ## Release Analysis

   Current version: vX.Y.Z (or none)
   Recommended bump: MINOR → vX.(Y+1).0

   ### Changes included:
   - feat: ...
   - fix: ...
   - docs: ...

   ### Breaking changes: None / [list them]
   ```

5. **Ask for confirmation** before proceeding.

---

## Phase 2: Prepare Documentation

1. **Update CHANGELOG.md** (create if needed) with new version section:
   ```markdown
   ## [vX.Y.Z] - YYYY-MM-DD

   ### Added
   - New feature 1
   - New feature 2

   ### Changed
   - Changed behavior 1

   ### Fixed
   - Bug fix 1
   ```

2. **Validate README.md** is current:

   a. **Check feature coverage** — Ensure all major features are documented:
      - Installation instructions
      - Configuration format
      - CLI usage examples
      - Supported platforms

   b. **Verify usage examples work**:
      ```bash
      cd fitwatch
      go build -o bin/fitwatch ./cmd/fitwatch
      ./bin/fitwatch --help
      ```

   c. **Check installation instructions** match current build:
      - Verify build commands work
      - Verify download links will be correct

---

## Phase 3: Quality Validation

1. **Run full test suite**:
   ```bash
   cd fitwatch
   go test -v -race ./...
   ```
   - All tests must pass

2. **Run static analysis**:
   ```bash
   cd fitwatch
   go vet ./...
   ```
   - 0 errors required

3. **Run linting** (if golangci-lint available):
   ```bash
   cd fitwatch
   golangci-lint run --timeout=5m
   ```

4. **Check formatting**:
   ```bash
   cd fitwatch
   gofmt -l .
   ```
   - No output means all formatted

5. **Build all platforms** to verify cross-compilation:
   ```bash
   cd fitwatch
   GOOS=linux GOARCH=amd64 go build -o /dev/null ./cmd/fitwatch
   GOOS=windows GOARCH=amd64 go build -o /dev/null ./cmd/fitwatch
   GOOS=darwin GOARCH=arm64 go build -o /dev/null ./cmd/fitwatch
   ```

6. **Report quality status**:
   ```
   ## Quality Validation

   | Check | Status |
   |-------|--------|
   | Tests | ✅ XXX passed |
   | Vet | ✅ Clean |
   | Lint | ✅ Clean |
   | Format | ✅ Clean |
   | Build (linux) | ✅ Success |
   | Build (windows) | ✅ Success |
   | Build (darwin) | ✅ Success |
   ```

---

## Phase 4: Execute Release

1. **Update version in code** (if version is embedded):
   ```go
   // In cmd/fitwatch/main.go or internal/version/version.go
   var version = "X.Y.Z"
   ```

2. **Commit release preparation**:
   ```bash
   git add CHANGELOG.md README.md
   git commit -m "chore: prepare release vX.Y.Z"
   git push origin main
   ```

3. **Create annotated tag**:
   ```bash
   git tag -a vX.Y.Z -m "Release vX.Y.Z

   Highlights:
   - Feature 1
   - Feature 2
   - Bug fix 1"
   ```

4. **Push tag to trigger release workflow**:
   ```bash
   git push origin vX.Y.Z
   ```

---

## Phase 5: Monitor and Verify

1. **Watch the release workflow**:
   ```bash
   gh run list --limit 5
   gh run watch  # Interactive watch
   ```

2. **If pipeline fails**:
   ```bash
   # Get failure details
   gh run view <run-id> --log-failed

   # Fix the issue, then delete and recreate tag
   git tag -d vX.Y.Z
   git push origin --delete vX.Y.Z

   # After fix, re-tag and push
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push origin vX.Y.Z
   ```

3. **Verify GitHub Release**:
   - Check release page has all binaries
   - Verify checksums.txt is present
   - Download and test a binary

4. **Test downloaded binary**:
   ```bash
   # Download from release
   ./fitwatch-linux-amd64 --help
   ./fitwatch-linux-amd64 --version
   ```

5. **Report final status**:
   ```
   ## Release Complete ✅

   - **Version**: vX.Y.Z
   - **Release**: https://github.com/<user>/mywhoosh-download/releases/tag/vX.Y.Z
   - **Binaries**: linux-amd64, linux-arm64, windows-amd64, darwin-amd64, darwin-arm64

   ### Post-release tasks:
   - [ ] Update any documentation links
   - [ ] Close related issues
   ```

---

## Pre-Release Versions

For pre-releases, use semver suffixes:
- `-alpha.1` for alpha releases
- `-beta.1` for beta releases
- `-rc.1` for release candidates

Example:
```bash
git tag -a v1.1.0-beta.1 -m "Beta: New MyWhoosh integration"
git push origin v1.1.0-beta.1
```

---

## Rollback Procedure

If a release has critical issues:

1. **Create hotfix branch**:
   ```bash
   git checkout -b hotfix/vX.Y.(Z+1)
   # Fix the issue
   git commit -m "fix: critical issue in vX.Y.Z"
   git checkout main
   git merge hotfix/vX.Y.(Z+1)
   git push origin main
   ```

2. **Release patch version** following the same ceremony.

3. **Update release notes** on the broken release to warn users.

---

## Checklist Summary

**Before tagging:**
- [ ] All tests pass
- [ ] Static analysis passes (go vet)
- [ ] Linting passes
- [ ] Code is formatted
- [ ] Cross-platform builds succeed
- [ ] CHANGELOG.md updated
- [ ] README.md is current
- [ ] Changes committed and pushed

**After tagging:**
- [ ] CI/CD pipeline succeeds
- [ ] Release appears on GitHub
- [ ] All platform binaries present
- [ ] Checksums file present
- [ ] Downloaded binary works
