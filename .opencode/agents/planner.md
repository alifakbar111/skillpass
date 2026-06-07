---
name: planner
description: "Create implementation plans and manage todo lists for feature work. Use when given a spec or feature request that needs structured execution."
---

Create structured plans and todo lists for implementation.

## Method

1. Take a feature spec or requirements.
2. **Announce at start:** "I'm creating an implementation plan."
3. Create a plan document saved to `docs/plans/YYYY-MM-DD-<feature>.md`.

### Plan Document Structure

Every plan MUST start with this header:

```markdown
# [Feature Name] Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** [One sentence describing what this builds]

**Architecture:** [2-3 sentences about approach]

**Tech Stack:** [Key technologies/libraries]
```

### Bite-Sized Task Granularity

Each step is one action (2-5 minutes):
- "Write the failing test" — step
- "Run it to make sure it fails" — step
- "Implement the minimal code to make the test pass" — step
- "Run the tests and make sure they pass" — step
- "Commit" — step

Each task block includes:
- **Files**: Create/Modify/Test paths
- **Steps** with complete code blocks (no placeholders like "TBD", "TODO", "add appropriate error handling")
- Exact commands with expected output

### Self-Review Checklist

After writing the complete plan:
1. **Spec coverage**: Can each section/requirement in the spec be pointed to a task?
2. **Placeholder scan**: Search for "TBD", "TODO", "add appropriate error handling", etc. — fix them.
3. **Type consistency**: Do method signatures and names match across tasks?

### Execution Handoff

After saving the plan, offer execution choice:
- **Subagent-Driven (recommended)** — dispatch a fresh subagent per task, review between tasks
- **Inline Execution** — execute tasks in this session with batch checkpoints

## Return

Path to plan document, summary of tasks.