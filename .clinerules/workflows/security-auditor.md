# Security Auditor (Cline Workflow)

Invoke via: `/security-auditor <target>`

Audit code for security vulnerabilities. Read-only — does not modify code.

## Scope

- JWT token handling and middleware ordering
- SQL injection paths in raw queries and go-jet expressions
- Role guards: `RequireRole`, `RequireVerifiedCompany`
- Password hashing: bcrypt vs argon2id usage
- CORS configuration
- Error message verbosity (info leakage)
- Environment variable exposure

## Method

### Confidence Levels
| Level | Criteria | Action |
|-------|----------|--------|
| **HIGH** | Vulnerable pattern + attacker-controlled input confirmed | **Report** with severity |
| **MEDIUM** | Vulnerable pattern, input source unclear | **Note** as "Needs verification" |
| **LOW** | Theoretical, best practice, defense-in-depth | **Do not report** |

### Do Not Flag
- Test files (unless explicitly reviewing test security)
- Dead code, commented code, documentation strings
- Patterns using constants or server-controlled configuration
- Code paths that require prior authentication to reach
- Framework-mitigated patterns (e.g., ORM parameterized queries)

### Review Process
1. **Detect context**: What type of code? (API endpoints, frontend, file handling, etc.)
2. **Research before flagging**: Trace data flow — is input attacker-controlled or server-controlled?
3. **Verify exploitability**: Confirm the input is attacker-controlled — not server-controlled.
4. **Report HIGH confidence only**: Skip theoretical issues.

### Severity Classification
| Severity | Impact |
|----------|--------|
| **Critical** | Direct exploit, severe impact, no auth required |
| **High** | Exploitable with conditions, significant impact |
| **Medium** | Specific conditions required, moderate impact |
| **Low** | Defense-in-depth, minimal direct impact |

## Return

Findings grouped by severity (critical/high/medium/low) with file:line and remediation advice. If nothing found, say so in one line.
