# SkillPass Team Setup

Use this to restore the agent team in a new session. Paste the following commands or ask Cline to run them.

## Spawn Teammates

```
team_spawn_teammate(
  agentId: "agent-manager",
  rolePrompt: "You are the Agent Manager — the orchestrator..."
)

Then for each specialist, read its workflow file from .clinerules/workflows/
```

## Available Agents

| Slash Command | Teammate ID | Workflow File |
|---|---|---|
| `/agent-manager` | agent-manager | `.clinerules/workflows/agent-manager.md` |
| `/bug-hunter` | bug-hunter | `.clinerules/workflows/bug-hunter.md` |
| `/code-reviewer` | code-reviewer | `.clinerules/workflows/code-reviewer.md` |
| `/db-migration` | db-migration | `.clinerules/workflows/db-migration.md` |
| `/go-scaffolder` | go-scaffolder | `.clinerules/workflows/go-scaffolder.md` |
| `/planner` | planner | `.clinerules/workflows/planner.md` |
| `/product-owner` | product-owner | `.clinerules/workflows/product-owner.md` |
| `/product-researcher` | product-researcher | `.clinerules/workflows/product-researcher.md` |
| `/react-scaffolder` | react-scaffolder | `.clinerules/workflows/react-scaffolder.md` |
| `/security-auditor` | security-auditor | `.clinerules/workflows/security-auditor.md` |
| `/technical-writer` | technical-writer | `.clinerules/workflows/technical-writer.md` |
| `/test-runner` | test-runner | `.clinerules/workflows/test-runner.md` |
| `/ui-ux-designer` | ui-ux-designer | `.clinerules/workflows/ui-ux-designer.md` |

## Restore Command

In a new session, just say: "Restore the SkillPass agent team" — and I'll respawn all 13 teammates.
