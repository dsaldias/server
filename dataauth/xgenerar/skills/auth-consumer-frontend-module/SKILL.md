---
name: auth-consumer-frontend-module
description: Create or extend consumer-project admin frontend modules using the auth_v2 Templ, Datastar and shared UI patterns. Read existing backend contracts as needed, but never modify backend code, GraphQL, repositories, models, migrations, authentication or authorization.
metadata:
  short-description: Build auth_v2 consumer admin frontends
---

# Auth consumer frontend modules

Use this skill when a project that consumes `auth_v2` needs a new administrative
frontend module such as activities, products, reservations, reports or another
domain feature. The output belongs to the consumer project and must reuse the
framework's layout, components and Datastar interaction model.

This is a frontend-only workflow. Backend inspection is read-only and exists
only to discover the operations and types that the frontend can consume.

## Hard boundary

Never create, edit or regenerate any of the following as part of this skill:

- GraphQL schemas, resolvers or generated GraphQL files;
- backend repositories, services, SQL or migrations;
- backend models or generated models;
- authentication, sessions, authorization or permission logic;
- framework-owned users, roles, units or other core modules.

If a required query, mutation, field, repository function or permission is
missing, report the missing backend contract and stop before changing backend
code. The user must handle that backend work separately.

Reading backend files, generated types, service signatures and existing route
registration is allowed. Do not confuse reading a contract with implementing
it.

## Consumer/framework boundary

Treat `auth_v2` as the platform layer. It owns login, session context,
authorization, the administrative layout, shared components and core modules.
The consumer owns its domain pages, frontend handlers, templates and routes.

Consumer routes should be supplied through the framework's existing extension
point, normally `RutasAdmin(..., extraRoutes...)`. Do not add consumer routes
to the framework's internal route list unless the user explicitly requests a
framework change.

Before editing, inspect the consumer project and identify:

1. its Go module and import path;
2. the import and registration of `RutasAdmin`;
3. the consumer's existing `data_web` modules, if any;
4. the frontend-facing repository/client/service functions already available;
5. the backend query and mutation contract, read-only.

## Required contract before implementation

Do not scaffold a module until the available contract is explicit. Record:

- module name and URL prefix;
- list query and its pagination/filter inputs;
- detail query and identifier parameter;
- create and update mutations, if supported;
- delete/archive operation, if supported;
- response fields needed by the table and detail view;
- input fields, required fields and optional fields;
- related catalog data needed by forms;
- frontend-facing functions used to invoke each operation;
- whether the module is list-only, read-only or full CRUD.

Never invent a mutation because a conventional CRUD route would normally have
one. If the contract only supports listing and viewing, do not generate create
or edit UI.

## Module layout

For a CRUD-like consumer module, use the established layout:

```text
data_web/<module>/
├── index.templ
├── lista.templ
├── formulario.templ       # only when create/edit exists
├── visualizar.templ       # when detail view exists
├── xcontroller.go
├── xhtml.go
├── xoperaciones.go        # only when POST operations exist
└── xrutas.go
```

Keep all files in one Go package. The leading `x` is only a project naming
convention; it does not create a separate responsibility boundary.

Do not manually edit generated `*_templ.go` files. Regenerate them after
changing `.templ` files.

## Frontend responsibilities by file

- `index.templ`: page shell, `marco.AdminTable`, header actions, filters,
  signals and the dialog shell.
- `lista.templ`: stable table region, table content, empty/error states and
  pagination. Wrap the content in a module-specific `TableRegion` helper.
- `formulario.templ`: create/edit form, field bindings, validation messages and
  dialog content. Reuse one form when create and edit share an interaction
  model.
- `visualizar.templ`: read-only detail view and related sections.
- `xcontroller.go`: page data, signal structs, list loading, formatting helpers
  and SSE table patches. Call only existing consumer frontend services,
  repositories or clients.
- `xhtml.go`: GET fragment handlers, modal loading, detail/form lookup and SSE
  modal patches.
- `xoperaciones.go`: POST handlers that read form signals and invoke existing
  consumer operations. This file is presentation orchestration, not backend
  persistence logic.
- `xrutas.go`: module-local frontend route constants and route registration.

## Route contract

Use explicit operation-oriented routes under the consumer module prefix:

```text
GET  /adminx/<module>                         full page
GET  /adminx/<module>/listahtml              list fragment
GET  /adminx/<module>/formulariohtml         create form fragment
GET  /adminx/<module>/visualizarhtml?id=...  detail fragment
GET  /adminx/<module>/editarhtml?id=...      edit form fragment
POST /adminx/<module>/crear                   create request
POST /adminx/<module>/editar                  edit request
```

Only include routes supported by the contract. Use a module-specific query
parameter consistently for detail lookup. Do not overload one endpoint based
on a `mode` value or silently accept unrelated HTTP methods.

## Datastar and UI rules

- Use server-rendered Templ HTML, Datastar signals/events/requests and SSE.
- Do not add a module-specific JavaScript file or DOM synchronization script.
- Use existing `data_web/components` and `data_web/xutility` components.
- Compose `marco.AdminTable` for the page frame.
- Keep the table inside a stable module `TableRegion` and patch only that
  region during refreshes.
- Keep dialogs outside the table region.
- Load forms and detail views on demand; do not embed every modal in the first
  page response.
- Reset create/edit signals when opening a form.
- Disable every network-triggering button immediately and restore it on
  completion or failure.
- Preserve active pagination, filters and URL state during refreshes.
- Clearing a filter must clear both its signal and request parameter.
- Pagination must show immediate loading feedback and prevent duplicate clicks.
- Render catalog and relationship values on the server in returned fragments.
- Ask for confirmation before irreversible consumer operations.

Inline Datastar expressions are acceptable for small interaction state, such
as opening a request, setting a signal or displaying loading feedback. Keep
domain data population and important rendering decisions on the server.

## Recommended implementation order

1. Inspect the consumer frontend and read the relevant backend contract.
2. Write down the module contract and identify missing operations before
   creating files.
3. Choose the closest existing consumer module or `data_web/usuarios` pattern.
4. Create the page data and signals in `xcontroller.go`.
5. Add the explicit frontend routes in `xrutas.go`.
6. Build `index.templ` and `lista.templ` with the shared marco components.
7. Add on-demand detail/form dialogs only for supported operations.
8. Connect handlers to existing consumer-facing functions without adding
   backend logic.
9. Register the module through the consumer's `extraRoutes` integration.
10. Regenerate templates, format and test only the resulting frontend scope.

## Verification

Run the commands appropriate to the consumer project, normally:

```text
go tool templ generate ./data_web/...
gofmt -w <changed-go-files>
go test ./data_web/...
git diff --check
```

If generation or tests expose a missing backend operation, do not fix it in
the backend under this skill. Report the exact missing contract and leave the
frontend change in a clearly reviewable state.

For detailed mapping from the `usuarios` reference module to consumer modules,
read [references/consumer-crud-pattern.md](references/consumer-crud-pattern.md).
