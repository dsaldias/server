---
name: setup-project
description: Guide the first-time setup of a new consumer project that imports github.com/dsaldias/server as its RBAC auth backend. Use when a user runs generar init or asks how to start a new project with this auth server.
---

# First-Time Project Setup

Walk the user through setting up a new project that uses `github.com/dsaldias/server` as its RBAC auth server.

## Prerequisites

The user needs:
- Go installed
- MySQL/MariaDB running
- This auth server available as a Go module

## Step 1 — Initialize the project

```bash
mkdir my-project && cd my-project
go mod init github.com/<user>/<my-project>
go get github.com/dsaldias/server
```

## Step 2 — Run the scaffolding CLI

From inside the new project:
```bash
go tool generar init
```

This generates:
- `serverx.go` — Chi router + GraphQL server setup
- `.env` — environment variables template
- `app/onevents.go` — hooks for custom events (user create, relogin, ticket)

## Step 3 — Create the database

Apply the auth server's base schema first — this creates all RBAC tables (usuarios, roles, permisos, unidades, sessionkeys, menus, tickets, etc.):

```bash
mysql -u root -p <your_db_name> < $(go env GOPATH)/pkg/mod/github.com/dsaldias/server@*/sqls/database.sql
```

Or if you have the server source locally:
```bash
mysql -u root -p <your_db_name> < /path/to/auth_v2/server/sqls/database.sql
```

## Step 4 — Configure `.env`

Copy `.env` and fill in:
```env
PORT=8038
DB_HOST=localhost
DB_USER=root
DB_PASS=yourpassword
DB_NAME=yourdbname
TOKEN_DURATION_MIN=60
PLAYGROUND=1
RATE_LIMIT=1
DECODE_PASS_KEY=some-secret-key-32chars
ALLOWED_ORIGINS=http://localhost:3000
DEFAULT_ROL_OAUTH=2
DEFAULT_UNIDAD_OAUTH=1
```

Key variables:
| Variable | Purpose |
|----------|---------|
| `TOKEN_DURATION_MIN` | JWT expiry in minutes |
| `PLAYGROUND` | Enable GraphQL UI at `/auth` and `/app` |
| `RATE_LIMIT` | Toggle 18 req/sec rate limiter |
| `DECODE_PASS_KEY` | Key used to hash/verify passwords |
| `PERM_EXTERNO` | Allow users to self-register |
| `EXTERNAL_AUTH` | URL to delegate auth to external service |

## Step 5 — Add your own entities

Use `/new-entity <entity-name>` for each new table you need to add.

Each entity follows the pattern:
```
dataauth/<entity>/
  crear.go     → Crear(db, input)
  listar.go    → Get<Entities>(db), GetById(db, id)
  update.go    → Actualizar(db, input)
  delete.go    → Eliminar(db, id)
  utils.go     → parseRow, parseRows (unexported)
```

## Step 6 — Register custom hooks (optional)

Edit `app/onevents.go` to react to auth events:
```go
func LoadCustomEvents() {
    // Called when a new user registers externally
    utils.SetOnUserExternalCreate(func(db *sql.DB, id, username, pass string) {
        // e.g., create default records for new users
    })

    // Called on every login
    utils.SetOnUserRelogin(func(db *sql.DB, id, username, pass string) {
        // e.g., sync external data on login
    })

    // Called when a support ticket is created
    utils.SetOnTicketCreated(func(db *sql.DB, ticketID string) {
        // e.g., send notification email
    })
}
```

## Step 7 — Run

```bash
go run serverx.go
```

Access the GraphQL playground at `http://localhost:8038/auth`

## Project structure after setup

```
my-project/
├── serverx.go              # Entry point (generated)
├── .env                    # Environment config
├── go.mod                  # Imports github.com/dsaldias/server
├── app/
│   └── onevents.go         # Custom event hooks
├── dataauth/               # Your entities (create with /new-entity)
│   └── productos/
│       ├── crear.go
│       ├── listar.go
│       ├── update.go
│       ├── delete.go
│       └── utils.go
├── graph_auth/             # GraphQL schemas and resolvers
│   ├── schema.graphqls     # Your operations
│   ├── productos.graphqls  # Per-entity types
│   └── schema.resolvers.go # Resolver implementations
└── sqls/                   # DB migrations for your entities
    └── productos.sql
```

## RBAC quick reference

- Every resolver must start with: `utils.CtxValue(ctx, r.DB, "operation_name")`
- Permission names map 1:1 to GraphQL operation names (e.g., `create_producto`)
- Permissions are granted per user or per role in the admin UI at `/auth`
- Session tokens sent via `SESSIONKEY` header or `galletita_traviesa` cookie
