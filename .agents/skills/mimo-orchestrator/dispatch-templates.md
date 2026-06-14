# Dispatch Templates

Reusable prompt templates for dispatching MiMo agents. Each template includes:
- Agent type mapping
- Required context
- Output expectations
- Error handling

---

## Bug Investigation Template

**Agent**: `bug-hunter` or `security-auditor`  
**Subagent Type**: `explore` (read-only) or `general` (with fixes)  
**Trigger**: Test failures, security concerns, code quality issues

```markdown
Investigate and fix [SPECIFIC_ISSUE] in [COMPONENT/MODULE].

## Context
- File(s): [LIST_FILES]
- Error: [ERROR_MESSAGE]
- Expected: [EXPECTED_BEHAVIOR]

## Task
1. Read the affected files completely
2. Identify root cause
3. If using `general` type: implement fix
4. Verify fix resolves the issue

## Constraints
- Do NOT modify files outside scope
- Preserve existing behavior for other cases
- Return: root cause + fix summary (or just findings if `explore`)

## Return Format
**Status**: success | partial | failed
**Root Cause**: [one-line description]
**Fix Applied**: [what changed, or "none" if explore-only]
**Files Modified**: [list, or "none"]
```

---

## Code Review Template

**Agent**: `code-reviewer` or `security-auditor`  
**Subagent Type**: `general`  
**Trigger**: PR review, security audit, quality check

```markdown
Review [FILE/PR/DIFF] for [security/quality/correctness].

## Scope
- Files: [LIST_FILES]
- Focus: [security | performance | correctness | all]
- Severity threshold: [critical | high | medium | low]

## Task
1. Read all specified files
2. Check against relevant checklists (security, performance, etc.)
3. Verify findings with surrounding context
4. Rank by severity

## Constraints
- Only report confirmed issues (not style preferences)
- Include file:line for each finding
- Note confidence level (high/medium/low)

## Return Format
**Status**: success | partial | failed
**Findings**: [ranked list with severity — file:line — description — fix]
**Overall Risk**: [low | medium | high | critical]
```

---

## Scaffolding Template

**Agent**: `go-scaffolder`, `react-scaffolder`, or `db-migration`  
**Subagent Type**: `general`  
**Trigger**: New feature, component, or module creation

```markdown
Create [COMPONENT_TYPE] for [FEATURE_NAME].

## Requirements
- Type: [handler | service | component | migration]
- Location: [PATH]
- Dependencies: [LIST_DEPENDENCIES]
- Conventions: [FOLLOW_PROJECT_PATTERNS]

## Task
1. Read existing similar components for patterns
2. Create new component following conventions
3. Wire up dependencies
4. If applicable: create tests

## Constraints
- Follow project naming conventions
- Use existing libraries (check package.json/go.mod)
- Maintain API response shape conventions

## Return Format
**Status**: success | partial | failed
**Created Files**: [list]
**Dependencies Added**: [list, or "none"]
**Next Steps**: [what to do after]
```

---

## Testing Template

**Agent**: `test-runner` or `bug-hunter`  
**Subagent Type**: `general`  
**Trigger**: Test failures, missing coverage, TDD workflow

```markdown
[Fix failing tests | Add tests for] [COMPONENT/MODULE].

## Context
- Test file: [PATH]
- Source file: [PATH]
- Error: [ERROR_MESSAGE] (if fixing)
- Coverage target: [what needs testing]

## Task
1. Read test file and source file
2. [If fixing: identify why tests fail]
3. [If adding: write tests following patterns]
4. Run tests to verify

## Constraints
- Use existing test patterns (vitest/Go testing)
- Mock external dependencies
- Test edge cases, not just happy path

## Return Format
**Status**: success | partial | failed
**Tests Fixed/Created**: [count]
**Coverage**: [what's covered]
**Failures Remaining**: [list, or "none"]
```

---

## Research/Exploration Template

**Agent**: Any agent for read-only tasks  
**Subagent Type**: `explore`  
**Trigger**: Codebase questions, architecture understanding, documentation

```markdown
Research [QUESTION/TOPIC] in the codebase.

## Scope
- Area: [MODULE/FEATURE]
- Files to examine: [LIST or "discovery"]
- Output: [documentation | summary | list]

## Task
1. Search for relevant files using glob/grep
2. Read and analyze key files
3. Synthesize findings

## Constraints
- Read-only (no modifications)
- Focus on facts, not opinions
- Cite file:line references

## Return Format
**Status**: success | partial | failed
**Findings**: [structured summary]
**Key Files**: [list with descriptions]
**Recommendations**: [if applicable]
```

---

## Multi-Agent Parallel Template

**Trigger**: Complex tasks with independent subtasks

```markdown
## Task Decomposition
[TASK_DESCRIPTION]

## Subtasks
1. [SUBTASK_1] → Agent: [AGENT] | Type: [explore/general]
2. [SUBTASK_2] → Agent: [AGENT] | Type: [explore/general]
3. [SUBTASK_3] → Agent: [AGENT] | Type: [explore/general]

## Dispatch Instructions
- Spawn all agents concurrently (if independent)
- Or sequentially (if dependent)
- Collect results from each
- Synthesize final output

## Conflict Resolution
- If agents modify overlapping files: merge or escalate
- If agents report conflicting findings: verify and reconcile

## Return Format
**Status**: success | partial | failed
**Subtask Results**: [per-agent summaries]
**Integration**: [how results combine]
**Conflicts**: [list, or "none"]
```

---

## Error Recovery Template

**Trigger**: Agent failure or timeout

```markdown
Previous agent [AGENT_ID] failed/timed out.

## Original Task
[ORIGINAL_PROMPT]

## Failure Details
- Error: [ERROR_MESSAGE]
- Partial output: [ANY_OUTPUT_RECEIVED]

## Recovery Options
1. Retry with extended timeout
2. Retry with more context
3. Simplify task scope
4. Skip and continue with other agents

## Action Taken
[WHICH_OPTION_SELECTED]

## Return Format
**Status**: recovered | partial | failed
**Recovery Method**: [description]
**Result**: [output after recovery]
```

---

## Usage Notes

1. **Template Selection**: Choose template based on task type
2. **Customization**: Fill in [PLACEHOLDERS] with specific values
3. **Combination**: Templates can be chained for complex workflows
4. **Error Handling**: Always include return format for orchestration

## Prompt Injection Prevention

When dispatching agents:
- Never include user input directly in prompts without sanitization
- Use parameterized templates (fill placeholders, don't concatenate)
- Validate agent outputs before using in subsequent prompts
- Log all dispatches for audit trail
