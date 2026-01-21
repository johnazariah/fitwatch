# Update Public-Facing Documentation

Systematically review and update all documentation to ensure it accurately reflects the current state of the codebase.

**Goal**: Every public API, feature, and capability should be documented in:
1. GoDoc comments (source of truth)
2. README (high-level overview and usage)
3. Configuration examples
4. CLI help text

---

## Phase 1: Discover Public API

Build a complete picture of what needs to be documented.

1. **Find all exported types and functions**:
   ```bash
   cd fitwatch
   
   # Exported types (capitalized)
   grep -rE "^type [A-Z]" pkg/ internal/ --include="*.go" | grep -v "_test.go"
   
   # Exported functions
   grep -rE "^func [A-Z]" pkg/ internal/ --include="*.go" | grep -v "_test.go"
   ```

2. **Check package documentation**:
   ```bash
   # Each package should have a doc.go or package comment
   for pkg in pkg/* internal/*; do
       echo "=== $pkg ==="
       head -20 "$pkg"/*.go 2>/dev/null | grep -A5 "^// Package"
   done
   ```

3. **Create inventory** — list every public item that needs documentation:
   - Exported types (structs, interfaces)
   - Exported functions
   - Configuration options
   - CLI flags

---

## Phase 2: Audit GoDoc Comments

GoDoc comments are the source of truth. They must be **complete** AND **accurate**.

### 2.1 Check comments exist

```bash
cd fitwatch

# Find exported symbols without documentation
go doc ./... 2>&1 | grep -E "^(func|type|var|const)" | head -30
```

### 2.2 Verify comments match implementation

For each exported type/function, verify the comment is **accurate**:

1. **Read the implementation** alongside the comment
2. **Check parameter documentation matches actual parameters**
3. **Verify described behavior matches actual behavior**
4. **Check for stale documentation**:
   - Parameters that were renamed or removed
   - Behavior that was modified
   - New parameters not yet documented

### 2.3 Validate examples compile

For each package with examples in `*_test.go`:

```bash
cd fitwatch
go test -run Example ./...
```

### 2.4 GoDoc completeness checklist

For each exported symbol, verify:

| Check | Struct | Function | Interface |
|-------|--------|----------|-----------|
| One-line summary | ✓ | ✓ | ✓ |
| Extended description (if complex) | ✓ | ✓ | ✓ |
| Field documentation | ✓ | - | - |
| Parameter purpose clear | - | ✓ | - |
| Return value documented | - | ✓ | - |
| Error conditions documented | - | ✓ | - |
| Example function | ✓ | ✓ | ✓ |
| Methods documented | ✓ | - | ✓ |

### 2.5 Fix inaccurate comments

When fixing a GoDoc comment:

1. Read the full implementation to understand actual behavior
2. Update the description to match what the code does
3. Document any error conditions
4. Add/update Example functions in `*_test.go`
5. Run `go doc` to verify rendering

---

## Phase 3: Audit README

README is the first thing users see — keep it current and complete.

### 3.1 Read current README

```bash
cat fitwatch/README.md
```

### 3.2 Verify sections

- [ ] **Features list** — mentions all major capabilities
- [ ] **Installation** — correct build/download instructions
- [ ] **Quick start** — working example
- [ ] **Configuration** — all options documented
- [ ] **CLI usage** — all flags documented
- [ ] **Supported platforms** — accurate list

### 3.3 Test quick start example

```bash
cd fitwatch
go build -o bin/fitwatch ./cmd/fitwatch
./bin/fitwatch --help
```

### 3.4 Verify CLI help matches README

```bash
cd fitwatch
go run ./cmd/fitwatch --help
```

Compare output with what README documents.

---

## Phase 4: Audit Configuration

### 4.1 Check config.example.toml exists and is complete

```bash
cat fitwatch/config.example.toml
```

### 4.2 Compare with actual config struct

```bash
grep -A 50 "type Config struct" fitwatch/internal/config/*.go
```

Ensure every field in the struct is documented in the example.

### 4.3 Verify defaults are documented

Check that default values mentioned in docs match code defaults.

---

## Phase 5: Build and Validate

1. **Build to verify code compiles**:
   ```bash
   cd fitwatch
   go build ./...
   ```

2. **Run tests**:
   ```bash
   cd fitwatch
   go test ./...
   ```

3. **Check for formatting issues**:
   ```bash
   cd fitwatch
   gofmt -l .
   ```

---

## Phase 6: Final Verification

Run all quality checks:

```bash
cd fitwatch

# Vet
go vet ./...

# Tests pass
go test -short ./...

# Build succeeds
go build ./cmd/fitwatch
```

---

## Output: Summary Report

After completing all phases, provide a summary:

```markdown
## Documentation Update Summary

### GoDoc Comments
- Fixed: X types/functions
- Status: ✅ Complete / ⚠️ Issues remain

### README
- Updated sections: [list]
- Status: ✅ Current / ⚠️ Needs update

### Configuration
- config.example.toml: ✅ Complete / ⚠️ Missing fields
- Status: ✅ Current / ⚠️ Needs update

### CLI Help
- All flags documented: ✅ Yes / ⚠️ Missing
- Status: ✅ Current / ⚠️ Needs update

### Validation
- [ ] go build passes
- [ ] go test passes
- [ ] go vet passes
- [ ] Examples run correctly

### Remaining Issues
- [any issues that need follow-up]
```

---

## Completion Checklist

- [ ] All exported symbols have GoDoc comments
- [ ] README reflects current capabilities
- [ ] config.example.toml documents all options
- [ ] CLI --help matches README documentation
- [ ] All code examples execute successfully
- [ ] go vet passes
