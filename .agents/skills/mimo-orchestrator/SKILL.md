---
name: mimo-orchestrator
description: Orchestrate MiMo subagents by reading agent definitions from .agents/agents/*.md, mapping them to explore/general subagent types, and dispatching via the actor tool. Use when coordinating multiple MiMo agents for parallel tasks, complex workflows, or multi-step operations.
---

# MiMo Orchestrator

## Overview

The MiMo Orchestrator coordinates subagent dispatch by reading agent definitions from `.agents/agents/*.md`, mapping agents to appropriate subagent types (`explore` or `general`), and dispatching them via the `actor` tool.

This skill enables intelligent task decomposition and parallel execution across the MiMo agent ecosystem.

## Agent Registry

The orchestrator discovers agents by reading `.agents/agents/*.md` files. Each agent file has:

```yaml
---
name: <agent-name>
description: <when-to-use>
---
<instructions>
```

### Agent to Subagent Type Mapping

| Agent Pattern | Subagent Type | Rationale |
|---------------|---------------|-----------|
| `*hunter*`, `*finder*`, `*searcher*` | `explore` | Read-only discovery tasks |
| `*reviewer*`, `*auditor*`, `*checker*` | `general` | Analysis with verification |
| `*scaffolder*`, `*builder*`, `*creator*` | `general` | Code generation tasks |
| `*runner*`, `*tester*`, `*validator*` | `general` | Execution tasks |
| `*planner*`, `*manager*`, `*coordinator*` | `general` | Orchestration tasks |
| Unknown patterns | `general` | Default to general |

## Orchestration Workflow

### 1. Task Decomposition

Analyze the request and decompose into independent subtasks:

```
Input: "Fix failing tests across 3 modules"
├── Module A tests → Agent: bug-hunter → explore
├── Module B tests → Agent: bug-hunter → explore  
└── Module C tests → Agent: bug-hunter → explore
```

### 2. Agent Selection

For each subtask:
1. Read `.agents/agents/*.md` to find matching agent
2. Extract agent instructions and constraints
3. Map to subagent type using table above
4. Craft focused prompt with agent context

### 3. Parallel Dispatch

Use the `actor` tool to dispatch agents concurrently:

```
actor(operation: {
  action: "spawn",
  subagent_type: "explore" | "general",
  description: "<task-description>",
  prompt: "<agent-specific-instructions + task-context>"
})
```

### 4. Result Collection

Wait for agents to complete and collect results:

```
actor(operation: {
  action: "wait",
  actor_id: "<actor-id>"
})
```

## Dispatch Prompt Templates

See `dispatch-templates.md` for reusable prompt templates:

- **Bug Investigation**: For debugging and root cause analysis
- **Code Review**: For security and quality audits
- **Scaffolding**: For creating new components/modules
- **Testing**: For running and fixing test suites
- **Research**: For codebase exploration and documentation

## Error Handling

### Agent Failures

- If an agent times out, retry once with extended context
- If an agent returns partial results, synthesize and continue
- If an agent fails completely, log and proceed with remaining agents

### Conflict Detection

When multiple agents modify files:
1. Check for overlapping file modifications
2. Merge changes if compatible
3. Escalate conflicts to user if incompatible

## Usage Examples

### Example 1: Parallel Bug Investigation

```
Task: "Find and fix security vulnerabilities in auth module"

Dispatch:
1. bug-hunter (explore) → audit auth/handler.go
2. security-auditor (general) → review auth/middleware.go
3. bug-hunter (explore) → check auth tests

Collect:
- All three return findings
- Synthesize into prioritized fix list
```

### Example 2: Module Scaffolding

```
Task: "Create notification system with tests"

Dispatch:
1. go-scaffolder (general) → create notification handler
2. go-scaffolder (general) → create notification service
3. test-runner (general) → write and run tests

Sequential: scaffolding first, then tests
```

## Best Practices

1. **Scope Isolation**: Each agent gets one clear responsibility
2. **Context Minimization**: Only provide what's needed for the task
3. **Parallel When Possible**: Independent tasks run concurrently
4. **Verify Results**: Always review agent outputs before integrating
5. **Fail Gracefully**: Handle agent failures without blocking entire workflow

## Integration with MiMo

This skill integrates with MiMo's actor tool:

- `explore` agents: Fast, read-only, good for discovery
- `general` agents: Full capabilities, can modify files, run commands

Map agents appropriately based on their instructions and the task requirements.
