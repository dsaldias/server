# Consumer mounting checklist

## Server

1. Add the library branch before generating the consumer files:

   ```bash
   go get github.com/dsaldias/server@version_shadcn
   go get -tool github.com/dsaldias/server/dataauth/xgenerar@version_shadcn
   go tool xgenerar init
   ```

   `xgenerar init` runs `go mod tidy` after generating `serverx.go`, which
   records `templ` and the other transitive dependencies automatically. If it
   reports that synchronization could not run, execute `go mod tidy` manually.
2. Construct the database handle and the auth repository dependencies required
   by the library.
3. Register the library's `RutasAdmin(db)` or the exported module route set at
   the `/adminx` prefix. Register it once to avoid duplicate handlers.
4. Serve the library assets, including the compiled component assets and the
   shared WebSocket client.
5. Keep the WebSocket endpoint and event payload fields compatible with the
   library. Chat events must include the conversation identifier when one is
   available.

## Layout

The bundled admin layout already renders the sidebar, header, role/unit
selector, profile actions, relogin dialog, notification UI, connection status,
and chat dialog. A consumer that uses `xlogin.InicioRender` does not need to
render those pieces again.

If a consumer supplies a custom authenticated layout, render the shared
components once and include `xchat.ChatDialog()` exactly once. Duplicate dialog
IDs or duplicate WebSocket clients cause stale signals and duplicated updates.

## Business module

Create the module's `xrutas.go` and append its routes to the authenticated admin
route set. Use `xutility/marco` for the common frame and return only the
business-specific table, filters, forms, and detail fragments. Keep repository
calls in the consumer's existing backend boundary and pass server-rendered
values into Templ components.

## Smoke test

- Open `/adminx` while unauthenticated and complete login.
- Select another unit/role and confirm menus reload.
- Refresh a list while keeping its page and filters.
- Open create, view, and edit dialogs more than once.
- Expire the session and complete relogin without a full-page loop.
- Open chat, select a conversation, send multiple messages, and receive
  multiple WebSocket messages while the history stays at the bottom.
