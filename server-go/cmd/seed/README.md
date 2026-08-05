# Database Seeders

This folder holds the **committed reference seeder** (`cmd/seed`). This README
also explains how to write your **own local demo seeders** for testing —
without leaking data into the repository.

---

## ⚠️ The golden rule: never commit dummy/PII-shaped data

Demo seeders create realistic-looking people: names, emails, **NIK/NPWP**,
**bank accounts**, salaries, addresses. Even when the values are fake, committing
them:

- looks like a **data leak** to anyone browsing the repo,
- pollutes `git blame`/history permanently, and
- confuses the next contributor about what is "real".

So we split seeders into two kinds:

| Kind | Example | Committed? | Contents |
|------|---------|-----------|----------|
| **Reference seeder** | `cmd/seed` | ✅ yes | Only non-personal reference data (industries, skills) + an admin created **from env vars**, never hard-coded |
| **Demo / dummy-data seeder** | `cmd/seeddemo`, `cmd/seedhris` | ❌ **no** | Fake companies, employees, jobseekers, PII-shaped fields |

Demo-seeder folders are **git-ignored** (see the root `.gitignore`):

```
server-go/cmd/seeddemo/
server-go/cmd/seedhris/
server-go/cmd/seed-local*/
```

If you create a new demo seeder, name its folder `seed-local<something>` (or add
it to `.gitignore`) so it can never be `git add`-ed by accident.

---

## How to create a demo seeder — step by step

### 1. Create the command folder (git-ignored name)

```
server-go/cmd/seed-local-mydata/main.go
```

### 2. Start from this template

```go
// Command seed-local-mydata inserts demo data for local testing ONLY.
// It is git-ignored and must never be committed (contains PII-shaped data).
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"skillpass-server-go/internal/models"
)

func main() {
	// Reads server-go/.env (DATABASE_URL). Same pattern as cmd/seed & cmd/migrate.
	_ = godotenv.Load(".env", "../.env")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx := context.Background()
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer sqlDB.Close()
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("ping: %v", err)
	}

	db := bun.NewDB(sqlDB, pgdialect.New())
	defer db.Close()

	// --- Idempotency: make re-runs safe. Either guard, or reset-then-insert. ---
	// Guard example: bail out if data already exists.
	exists, err := db.NewSelect().Model((*models.User)(nil)).
		Where("email = ?", "demo01@example.test").Exists(ctx)
	if err != nil {
		log.Fatalf("check: %v", err)
	}
	if exists {
		fmt.Println("Demo data already seeded. Nothing to do.")
		return
	}

	now := time.Now()

	// --- Insert using the Bun models in internal/models. ---
	users := []models.User{
		{Email: "demo01@example.test", Username: "demo01", PasswordHash: "<bcrypt>", Role: "jobseeker", Name: "Demo One", IsVerified: true, CreatedAt: now},
	}
	if _, err := db.NewInsert().Model(&users).Exec(ctx); err != nil {
		log.Fatalf("insert users: %v", err)
	}
	// users[i].ID is now populated — use it for related rows (profiles, employees…).

	fmt.Printf("Seeded %d users.\n", len(users))
}
```

### 3. Conventions to follow

- **Emails/domains**: use an obviously-fake domain like `@example.test` or
  `@skillpass.test`. Never use real people's data.
- **Passwords**: hash once with `bcrypt.GenerateFromPassword` and reuse the
  string for all demo users (same hash works for the same password) — fast and
  fine for local data.
- **Idempotency**: pick one —
  - **Guard**: check for a marker row and exit if present, or
  - **Reset-then-insert**: `db.NewDelete().Model((*models.User)(nil)).Where("email LIKE 'demo%@example.test'").Exec(ctx)` first (FK `ON DELETE CASCADE` cleans up children), then insert.
- **Foreign keys**: insert parents first, read back their generated IDs (Bun
  populates the slice), then insert children referencing those IDs.
- **Enums**: always set enum columns explicitly (e.g. `Status: "open"`,
  `employment_status: "active"`) — the Go zero value `""` is not a valid enum
  and the insert will fail.

### 4. (Optional) add a convenience script

If you want `bun run …` shortcuts, add them to the **root `package.json`** —
but **do not commit those lines** if they point at a git-ignored seeder
(otherwise the repo references a folder that isn't there):

```jsonc
"db:seed:local": "go -C server-go run ./cmd/seed-local-mydata/",
```

Otherwise just run it directly:

```bash
go -C server-go run ./cmd/seed-local-mydata/
```

### 5. Prerequisites to run

```bash
docker compose up db -d      # Postgres must be running
bun run db:migrate           # schema must be up to date
```

---

## Checklist before you commit anything

- [ ] My seeder folder is git-ignored (name matches `.gitignore`, or I added it).
- [ ] `git status` does **not** show my seeder or any dummy data as staged.
- [ ] No real names, emails, NIK/NPWP, bank numbers, or salaries anywhere.
- [ ] The committed reference seeder (`cmd/seed`) still only has non-personal data.

---

## The committed reference seeder (`cmd/seed`)

`cmd/seed/main.go` seeds only safe reference data and is intended to be run on
every environment:

- **Industry categories** and **skills** (`ON CONFLICT (name) DO NOTHING`).
- An **admin user**, only if `ADMIN_EMAIL` **and** `ADMIN_PASSWORD` are set in
  the environment — credentials are never hard-coded.

Run it with:

```bash
bun run db:seed
```

---

## Phase 2 demo seeder (Sprints 1–4) — how to build it

The Phase 2 demo data lives in a **git-ignored** local seeder, `cmd/seedphase2`
(ignored via `.gitignore`, like `seeddemo`/`seedhris`). It is **not committed**
because it produces PII-shaped data (documents, biometrics, bank amounts). Here
is how to recreate it.

### What it seeds (aligned with `seedhris`'s employees)
| Sprint | Data |
|--------|------|
| S1 Documents | a handbook + per-employee documents (real files via the storage layer, `scan_status='clean'`) + access logs |
| S2 Face | one **active enrollment per active employee** (encrypted embedding) + verification logs |
| S3 Identity | a **DID + signed employment credential** per employee, and an **integrity anchor** per document |
| S4 Payroll | varied `ptkp_status`, a `bpjs_config` row per company, and a **computed payroll run** per company |

### Key techniques (so the seeded data actually works)
1. **Reset-then-insert** for idempotency; scope to the demo companies
   (`WHERE u.email LIKE 'hr%@skillpass.test'`). Delete in FK-safe order and
   `os.RemoveAll` the document files under `DOCUMENTS_DIR`.
2. **Documents**: write files through `storage.NewLocalStoreAt(DOCUMENTS_DIR).Save(key, r)`
   using the same key format the app uses (`documents/<companyID>/<docID>.txt`),
   then insert the `documents` row with the file's real SHA-256.
3. **Face embeddings must be decryptable by the app.** Encrypt with
   **AES-256-GCM using `sha256(JWT_SECRET)`** — identical to
   `internal/face/crypto.go`. If the key differs, verification fails.
4. **Reuse the real services** so production code paths run:
   - `identity.NewService(db, identity.NewSigner(JWT_SECRET))` → `IssueDID`,
     `IssueCredential`, `Anchor`.
   - `payroll.NewService(db)` → `CreateRun` then `CalculateRun` (this runs the
     **real PPh21 TER + BPJS** engine, producing correct payslips + bank transfers).
5. **Vary `ptkp_status`** across employees so PPh21 exercises all three TER
   categories (A/B/C).

### Run it
```bash
docker compose up db face-service -d
bun run db:migrate
bun run db:seed          # reference data
bun run db:seed:demo     # companies + jobseekers
bun run db:seed:hris     # employees + RBAC
bun run db:seed:phase2   # <- the Phase 2 seeder (add this script locally; do not commit it)
```
Add the `db:seed:phase2` script to the root `package.json` **locally only** —
don't commit it, since it points at a git-ignored folder.
