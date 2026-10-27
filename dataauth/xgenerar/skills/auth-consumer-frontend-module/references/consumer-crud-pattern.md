# Consumer CRUD frontend pattern

This reference describes how to adapt the complete `data_web/usuarios` module
to a domain module in a consumer project. It is a frontend recipe, not a
backend generator.

## Contract mapping

Before writing templates, map the consumer contract into a small table:

| Frontend need | Required contract | Result |
|---|---|---|
| Table | list query and response | `PageData`, table columns |
| Filters | query input fields | Datastar signals and query parameters |
| Detail | detail query and ID | `visualizar.templ` and GET fragment |
| Create | create mutation/input | create form and POST handler |
| Edit | update mutation/input | edit form and POST handler |
| Catalog | list query or existing service | server-rendered selects/comboboxes |
| Delete/archive | explicit mutation | confirmation action only if supported |

If a row is not supported by the contract, omit that UI capability.

## `usuarios` adaptation map

Use the names below as a conceptual mapping, not as literal code to copy:

```text
UsuariosPageData       -> <Module>PageData
usuariosSignals        -> <module>Signals
UsuariosTableRegion    -> <Module>TableRegion
UsuariosTableContent   -> <Module>TableContent
UsuariosRender         -> <Module>Render
UsuariosListaHTML      -> <Module>ListaHTML
UsuarioFormularioRender-> <Entity>FormularioRender
UsuarioVisualizarRender-> <Entity>VisualizarRender
UsuarioEditarRender    -> <Entity>EditarRender
UsuarioCrear           -> <Entity>Crear
UsuarioEditar          -> <Entity>Editar
ADMIN_USUARIOS_*       -> ADMIN_<MODULE>_*
```

Choose a singular name for entity-level functions and a plural name for list,
page-data and route-prefix concepts, matching the local project's naming
style.

## Page flow

The initial GET should render only:

1. the administrative page shell;
2. the filters needed by the list;
3. the first page of table data;
4. a lightweight dialog loading shell.

The list refresh should:

1. read current signals;
2. normalize pagination and filters;
3. call the existing consumer list function;
4. render the complete table region;
5. patch only the region with SSE;
6. update the browser URL and loading signal.

The detail/form GET should:

1. read and validate the row identifier;
2. load only the requested entity and required catalogs;
3. render the modal content server-side;
4. patch the dialog content;
5. restore the triggering action and update the modal title/description.

The create/edit POST should:

1. read Datastar signals or the fallback form values;
2. normalize optional values;
3. invoke the existing consumer-facing operation;
4. patch only the form error on failure;
5. refresh the table region and close/reset the dialog on success.

## Consumer integration

The module should expose its own route collection:

```go
func Rutas(db *sql.DB) []*utilidad.Handlers2
```

The consumer then passes those routes through the framework's existing admin
extension point. The exact import and registration names depend on the
consumer's module path. Do not edit the framework's internal `RutasAdmin`
implementation to register the consumer module.

## Backend boundary check

Before finishing, inspect the diff and confirm that it contains only consumer
frontend files, route registration needed by the consumer, generated Templ
output and frontend tests. A diff touching GraphQL schema, generated models,
repositories, SQL, auth middleware or framework core is outside this skill's
scope and requires separate authorization.
