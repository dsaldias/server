---
name: auth-v2-consumer
description: Integrate the auth_v2 RBAC frontend into a consumer Go/Templ application so the consumer can focus on business modules instead of authentication, roles, users, menus, chat, and admin layout.
metadata:
  short-description: Integrate auth_v2 into a consumer project
---

# auth_v2 consumer integration

Use this skill when adding the auth_v2 admin frontend to a consumer project.
The library owns authentication, session cookies, role/unit context, menus,
users, roles, units, notices, tickets, notifications, chat, WebSocket status,
shared shadcn-templ components, and the admin layout. Consumer code should own
only business-domain modules and their authorization-aware content.

## Integration contract

- Reuse the library's `data_web/xlogin` layout and `data_web/xutility` shared components.
- `go tool xgenerar init` synchronizes the consumer module after generating
  `serverx.go`, so transitive runtime dependencies such as
  `github.com/a-h/templ` and `github.com/starfederation/datastar-go` are
  recorded in the consumer's `go.sum`. It also
  registers the `templ` and `shadcn-templ` tools required by the generated
  `Taskfile.yml`, and generates consumer-local shadcn aliases under
  `<module>/app/front/componentes` and `<module>/app/front/utils`. If automatic
  synchronization cannot run, execute `go mod tidy` and register those tools
  manually.
- The generator also creates a minimal sample module under
  `app/front/web/test`. It is available at `/webc/test` and is intentionally
  disposable: use it to verify the reusable frame, table, filters and
  pagination before replacing it with a business module.
- Register the library's admin routes once, passing consumer module routes as
  additional routes so direct refreshes receive the full admin layout. Do not recreate login, relogin,
  profile, role selection, menu loading, chat, or WebSocket routes in the consumer.
- Render `xchat.ChatDialog()` once inside the main authenticated layout if the
  consumer uses the chat package independently of the bundled login layout.
- Mount the library's WebSocket client asset and preserve the event contract
  used by the library for connection status, notifications, alerts, and chat.
- Keep backend access behind the repository functions exposed by
  `dataauth/xrepo/repo.go`; do not duplicate authentication or RBAC queries in
  consumer handlers.
- Keep module endpoints explicit and operation-oriented. A consumer module
  should separate its full page, list fragment, modal fragments, and create,
  update, and delete operations.

## Consumer module shape

Follow the same structure used by the maintained CRUD modules:

- `index.templ`: page shell, title, filters, actions, and stable table region.
- `lista.templ`: table and pagination fragment only.
- `formulario.templ`: create/edit dialog and form.
- `visualizar.templ`: read-only detail dialog.
- `xcontroller.go`: shared data and helpers.
- `xhtml.go`: GET fragment handlers.
- `xoperaciones.go`: POST operation handlers.
- `xrutas.go`: module route declarations.

Compose the page with `data_web/xutility/marco` instead of copying table,
filter, pagination, dialog, or loading markup. Supply only the business
module's columns, filters, data, and operation handlers.

The generated consumer `components.json` must use aliases inside the consumer
module, not the library module path. The generator replaces `{{MODULE}}` with
the value from the consumer's `go.mod`, for example:

```json
"aliases": {
  "components": "my_server_app/app/front/componentes",
  "utils": "my_server_app/app/front/utils"
}
```

These aliases are for the consumer-side `shadcn-templ` CLI. They do not change
the Go import paths of reusable library components.

## Datastar rules

- Prefer server-rendered Templ fragments and Datastar SSE updates. Do not add
  module-specific JavaScript files.
- Patch only the stable region that changed. Keep dialogs outside table regions.
- Every network button must disable itself and show a loading state until the
  response completes or fails.
- Load modal content on demand and render edit values on the server.
- Preserve pagination and active filters across refreshes.
- Clearing a filter must clear both its signal and its request parameter.
- Keep a persistent fragment stable when only its children change. For example,
  in chat update only the messages fragment, not the entire form and dialog.

## Routes

Use query parameters for fragment lookup when that matches the library's
convention, for example `?userid=...`. Keep route names easy to discover:

- `GET /adminx/<module>`: full page.
- `GET /adminx/<module>/listahtml`: list fragment.
- `GET /adminx/<module>/formulariohtml`: create form.
- `GET /adminx/<module>/visualizarhtml?userid=...`: view fragment.
- `GET /adminx/<module>/editarhtml?userid=...`: edit form.
- `POST /adminx/<module>/crear`: create.
- `POST /adminx/<module>/editar`: update.

## Validation

After integration, run `templ generate`, `gofmt`, the relevant package tests,
`go test ./...`, and `git diff --check`. Verify a consumer can log in, relogin,
change role/unit context, navigate menus, use chat, receive WebSocket events,
and load a business module without duplicating RBAC UI.

For the detailed mounting checklist, read
[`references/integration.md`](references/integration.md).
