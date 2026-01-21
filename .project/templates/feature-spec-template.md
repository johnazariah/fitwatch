# Feature Specification Template

> **How to use**: Copy this template for each new feature. Fill in sections as you design.
> Some sections can be minimal initially and expanded during implementation.

---

# Feature: [Feature Name]

## Priority
<!-- P0 = Must have for MVP, P1 = Important, P2 = Nice to have -->
P?

## Summary
<!-- One paragraph: What is this feature and why do we need it? -->

## User Stories

<!--
Format: As a [role], I want [capability] so that [benefit]
List 2-5 user stories that this feature addresses.
-->

- As a [user type], I want [action] so that [benefit]
- 

## Acceptance Criteria

<!--
GUIDING QUESTIONS:
- What must be true for this feature to be "done"?
- How will we test/demo this feature?
- What are the edge cases?

Format as checkboxes - these become your test cases!
-->

### Core Functionality
- [ ] [User can do X]
- [ ] [System handles Y correctly]
- [ ] 

### Edge Cases & Error Handling
- [ ] [What happens when Z fails?]
- [ ] [How do we handle invalid input?]
- [ ] 

### Performance (if applicable)
- [ ] [Response time < X seconds]
- [ ] [Memory usage < Y MB]

## Technical Design

<!--
GUIDING QUESTIONS:
- What components/modules are involved?
- What's the data flow?
- What external services/APIs do we interact with?
- What new data structures do we need?
-->

### Components
<!-- List the main pieces that make up this feature -->

### Data Model
<!-- Any new tables, fields, or structures? -->

```python
# Sketch out key data structures
```

### API (if applicable)
<!-- New endpoints or changes to existing ones -->

```
POST /api/example
GET  /api/example/{id}
```

### Configuration
<!-- Any new config options? -->

```yaml
feature_name:
  option: value
```

## Dependencies

<!--
- What other features must exist first?
- What external services do we need?
- Any new libraries to add?
-->

### Depends On
- [ ] [Feature/component that must exist first]

### Depended On By
- [ ] [Features that will use this]

## Test Plan

### Unit Tests
- [ ] [Test case 1]
- [ ] [Test case 2]

### Integration Tests
- [ ] [Test interaction with X]

### Manual Testing Steps
1. [Step to manually verify feature works]
2. 

## Estimates

<!--
Be honest - include time for:
- Design/research
- Implementation  
- Testing
- Documentation
- Code review
-->

| Task | Estimate |
|------|----------|
| Research | ? hours |
| Implementation | ? hours |
| Testing | ? hours |
| Documentation | ? hours |
| **Total** | **? hours** |

## Open Questions

<!--
Things you don't know yet. Come back and answer these.
-->

- [ ] ?

## Notes

<!-- Any other context, links to research, related discussions -->

## Changelog

| Date | Author | Change |
|------|--------|--------|
| YYYY-MM-DD | [Name] | Initial draft |
