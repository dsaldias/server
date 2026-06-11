# Prompt para Desarrollar Proyectos Consumidores con github.com/dsaldias/server

Eres un agente de IA encargado de desarrollar un proyecto consumidor en Go usando la librería base `github.com/dsaldias/server`. Debes construir sobre esta base, respetar sus patrones y evitar reimplementar funcionalidades que ya existen.

Este repositorio es un framework base reutilizable para servidores Go con GraphQL/gqlgen, autenticación, autorización RBAC, administración de usuarios, roles, permisos, menús, unidades, notificaciones, tickets, archivos y utilidades HTTP. Los proyectos consumidores lo integran con:

```go
require github.com/dsaldias/server
```

Luego inicializan la estructura con:

```bash
go tool generar init
go tool generar db
```

El directorio `generar` pertenece al sistema de generación y no debe analizarse ni modificarse como parte de la lógica funcional del consumidor.

## 1. Descripción General del Framework Base

El objetivo del framework es entregar una base funcional para nuevos proyectos backend en Go, evitando comenzar desde cero cada vez. Proporciona un servidor HTTP con GraphQL, autenticación por sesión/JWT, RBAC por roles, permisos y unidades, gestión de usuarios, menús para frontends, notificaciones WebSocket/SSE, tickets de soporte, carga/lectura de imágenes y acceso REST delegado hacia operaciones GraphQL.

Problemas que resuelve:

- Inicialización rápida de un servidor Go con gqlgen.
- Autenticación centralizada y manejo de sesión.
- Control de permisos por operación GraphQL.
- Gestión común de usuarios, roles, unidades, permisos y menús.
- Exposición simultánea de API GraphQL del sistema base y API GraphQL propia del consumidor.
- Adaptador REST para consumir queries/mutations GraphQL mediante endpoints HTTP.
- Infraestructura común de WebSocket, SSE, CORS, cookies y rate limiting.
- Esquema SQL base para MySQL con tablas `rbac_*`.

Alcance funcional:

- Backend Go sobre `net/http`, `chi`, `gqlgen`, MySQL y `database/sql`.
- GraphQL principal del consumidor en `/query` y `/ws_app`.
- GraphQL de autenticación/base en `/query_auth` y `/ws`.
- Playground opcional en `/auth` y `/app`.
- Archivos estáticos desde `/res/*`.
- REST bridge en `/rest/query/{operationName}`, `/rest/mutation/{operationName}`, `/rest_auth/query/{operationName}` y `/rest_auth/mutation/{operationName}`.
- SSE en `/sse`.

## 2. Arquitectura General

La organización funcional detectada es:

- `dataauth`: paquete principal de la librería base. Contiene el arranque del servidor y los módulos funcionales.
- `dataauth/main.go`: función `Iniciar`, que configura router, middlewares, GraphQL auth, GraphQL app, WebSocket, SSE, playgrounds, handlers extra, estáticos y adaptador REST.
- `dataauth/utils`: middlewares, conexión a base de datos, validación de `.env`, rate limit, REST bridge, callbacks/eventos y helpers de contexto.
- `dataauth/login`: login normal, login v2 y login externo.
- `dataauth/usuarios`: CRUD y consultas de usuarios, perfil, OAuth, permisos sueltos, roles y menús sueltos.
- `dataauth/roles`: CRUD de roles y asignación de permisos/menús.
- `dataauth/permisos`: listado y verificación de permisos.
- `dataauth/menus`: listado de menús, menús por rol, menús por usuario/unidad.
- `dataauth/unidades`: CRUD/listado de unidades.
- `dataauth/sessionkey`: creación y búsqueda de claves de sesión.
- `dataauth/notificaciones`: avisos persistentes visibles por fecha.
- `dataauth/xnotificaciones`: notificaciones en tiempo real por WebSocket/SSE y registro de conexiones.
- `dataauth/tickets`: tickets de soporte y respuestas.
- `dataauth/dashboard`: reportes base.
- `dataauth/archivos`: subida/compresión/lectura de imágenes.
- `graph_auth`: schema, resolvers y modelos gqlgen de la API base.
- `sqls/database.sql`: tablas base, seeds, permisos iniciales, roles iniciales, menús reservados y usuario admin.

Convenciones de carpetas:

- Cada módulo funcional vive en su propio paquete dentro de `dataauth/<modulo>`.
- Las operaciones se separan por archivo: `crear.go`, `listar.go`, `update.go`, `delete.go`, `utils.go`.
- Los resolvers GraphQL delegan la lógica real a servicios de `dataauth/<modulo>`.
- Los modelos GraphQL generados se usan desde `github.com/dsaldias/server/graph_auth/model`.
- La base utiliza MySQL y tablas con prefijo `rbac_`.

Convenciones de nomenclatura:

- Queries/mutations GraphQL usan snake_case: `create_usuario`, `update_rol`, `usuario_by_id`.
- El permiso RBAC usa exactamente el nombre del método GraphQL en la columna `rbac_permisos.metodo`.
- Servicios Go exportados suelen usar nombres en español: `Crear`, `Actualizar`, `Listar`, `GetById`, `GetMe`.
- Inputs GraphQL siguen prefijos `New`, `Update`, `Upd`, `Query`: `NewUsuario`, `UpdateUsuario`, `UpdUnidad`, `QueryUsuarios`.
- Los IDs GraphQL se manejan como `string`.
- El consumidor debe reservar IDs de menú menores o iguales a 10 para la base. Sus menús deben comenzar desde 10 o superior.

Forma correcta de extender el sistema:

- Crear el schema GraphQL propio del consumidor en su proyecto.
- Generar resolvers propios con gqlgen.
- Crear paquetes por dominio siguiendo el patrón `crear.go`, `listar.go`, `update.go`, `utils.go`.
- En cada resolver protegido llamar a `utils.CtxValue(ctx, db, "nombre_operacion")`.
- Insertar en `rbac_permisos` una fila por cada operación protegida, usando como `metodo` el mismo nombre pasado a `CtxValue`.
- Crear menús propios en `rbac_menus` con IDs desde 10.
- Asociar permisos a roles mediante `rbac_rol_permiso`.
- Asociar menús a roles mediante `rbac_rol_menus`.
- Asociar roles a usuarios y unidades mediante `rbac_rol_usuario_unidades`.
- Usar `dataauth.Iniciar(...)` para montar el servidor base y la API del consumidor.

## 3. Funcionalidades ya Implementadas

Autenticación:

- Login con `login(input: NewLogin!)`.
- Login v2 con `login_v2(input: NewLogin2!)`.
- Sesión basada en `rbac_session_keys`.
- JWT firmado con `JWT_SECRET`.
- Soporte de cookie `galletita_traviesa`.
- Duración configurable con `TOKEN_DURATION_MIN`.
- Validación de cuenta activa.

Autorización RBAC:

- Permisos por operación GraphQL.
- Permisos asignados por rol.
- Permisos directos por usuario.
- Roles vinculados a usuarios por unidad.
- El header `UNIDAD` es obligatorio para operaciones protegidas.
- Header `SESSIONKEY` identifica la sesión.
- Header `ROL` se captura en contexto, aunque la verificación principal se hace por usuario, unidad y método.

Gestión de usuarios:

- Crear, actualizar, listar, buscar por ID, actualizar perfil.
- Usuarios OAuth.
- Login externo configurable.
- Roles por unidad.
- Permisos sueltos.
- Menús sueltos.
- Foto de perfil en base64 convertida a WebP.
- Ubicación geográfica como `POINT`.

Roles y permisos:

- CRUD de roles.
- Asignación de permisos y menús a roles.
- Listado global de permisos.
- Verificación de permiso con `permisos.VerificarPermiso`.

Menús:

- Menús jerárquicos con `padre_id`.
- Menús por rol.
- Menús sueltos por usuario.
- Menús por usuario y unidad para construir navegación frontend.

GraphQL:

- Schema base en `graph_auth/schema-auth.graphqls` y `graph_auth/schema.graphqls`.
- Resolvers base en `graph_auth/schema.resolvers.go`.
- Endpoint auth: `/query_auth`.
- Endpoint app consumidor: `/query`.
- WebSocket auth: `/ws`.
- WebSocket app consumidor: `/ws_app`.
- Subscriptions base: `notificaciones_subs`.

Base de datos:

- MySQL con `database/sql`.
- Conexión mediante variables `DB_USER`, `DB_PASS`, `DB_HOST`, `DB_NAME`.
- Pool configurable con `DB_CONN_LIFETIME_MIN`, `DB_MAX_OPEN`, `DB_MAX_IDLE`.
- Esquema base en `sqls/database.sql`.

Otras funcionalidades:

- CORS con `ALLOWED_ORIGINS`.
- Compresión HTTP.
- Rate limit opcional con `RATE_LIMIT=1`.
- Playgrounds con `PLAYGROUND=1`.
- REST bridge hacia GraphQL.
- Eventos/callbacks para creación externa de usuarios, relogin y tickets.
- Notificaciones WebSocket/SSE.
- Tickets de soporte.
- Reportes dashboard básicos.
- Archivos estáticos en `res`.

## 4. API Interna Disponible

### Arranque del Servidor

`dataauth.Iniciar(srv *handler.Server, schema *graphql.ExecutableSchema, db *sql.DB, handlers []*utils.Handlers2, router *chi.Mux)`

- Propósito: iniciar el servidor HTTP completo con GraphQL auth, GraphQL app, middlewares, WebSocket, SSE, playgrounds, REST bridge y estáticos.
- Entrada: servidor gqlgen del consumidor, schema ejecutable del consumidor, conexión DB, handlers HTTP extra, router opcional.
- Retorno: no retorna; ejecuta `http.ListenAndServe`.
- Uso recomendado: llamarlo desde `serverx.go` o entrypoint del consumidor después de crear el schema propio.

### Utilidades de Base de Datos y Entorno

`utils.Conexion() *sql.DB`

- Propósito: abrir conexión MySQL usando `.env`.
- Entrada: variables de entorno.
- Retorno: `*sql.DB`.
- Uso recomendado: conexión estándar de proyectos consumidores.

`utils.VerificarEnv()`

- Propósito: validar variables requeridas.
- Entrada: archivo `.env`.
- Retorno: aborta con `log.Fatalf` si faltan variables.
- Uso recomendado: antes de iniciar el servidor.

`utils.GetAllowedOrigins() []string`

- Propósito: parsear `ALLOWED_ORIGINS`.
- Entrada: variable separada por comas.
- Retorno: lista de origins o `["*"]`.
- Uso recomendado: CORS personalizado si el consumidor arma router propio.

### Autenticación y Contexto

`utils.AuthMiddleware(db *sql.DB) func(http.Handler) http.Handler`

- Propósito: leer `SESSIONKEY` o cookie, validar sesión/JWT y guardar `AuthData` en el contexto.
- Entrada: DB.
- Retorno: middleware HTTP.
- Uso recomendado: usarlo mediante `dataauth.Iniciar`; no duplicarlo.

`utils.CtxValue(ctx context.Context, db *sql.DB, metodo string) (*utils.AuthData, error)`

- Propósito: obtener sesión autenticada y validar permiso si `metodo` no está vacío.
- Entrada: contexto GraphQL/HTTP, DB, nombre de permiso.
- Retorno: datos de auth con claims, sesión, unidad, rol y token.
- Uso recomendado: primera línea de todo resolver protegido del consumidor.

Ejemplo:

```go
auth, err := utils.CtxValue(ctx, r.DB, "listar_clientes")
if err != nil {
    return nil, err
}
userid := auth.SessionKey.UsuarioID
```

`utils.GenerateToken(ctx context.Context, userID string) (string, time.Time, int32, error)`

- Propósito: generar JWT y expiración.
- Entrada: contexto y user ID.
- Retorno: token, expiración, duración en minutos, error.
- Uso recomendado: flujos de login custom solo si no basta `login.Login`.

`utils.JwtValidate(token string) (*jwt.Token, error)`

- Propósito: validar token JWT.
- Entrada: string token.
- Retorno: token parseado o error.
- Uso recomendado: validaciones avanzadas.

`utils.CtxSetCookie(ctx context.Context, token string, exp time.Time)`

- Propósito: setear cookie de sesión.
- Entrada: contexto con response writer, token, expiración.
- Retorno: ninguno.
- Uso recomendado: login o flujos que emitan sesión.

`utils.UaserIDMiddleware(db *sql.DB) transport.WebsocketInitFunc`

- Propósito: inicializar contexto de WebSocket con usuario.
- Entrada: DB.
- Retorno: init function para gqlgen WebSocket.
- Uso recomendado: se monta desde `dataauth.Iniciar`.

`utils.CtxUserIDWs(ctx context.Context, db *sql.DB, metodo string) string`

- Propósito: obtener usuario desde contexto WebSocket y validar permiso.
- Entrada: contexto, DB, método.
- Retorno: user ID.
- Uso recomendado: subscriptions custom protegidas.

### REST y HTTP

`utils.RestToGraphQlHandler(schema graphql.ExecutableSchema) http.HandlerFunc`

- Propósito: exponer operaciones GraphQL como REST.
- Entrada: schema gqlgen.
- Retorno: handler HTTP.
- Uso recomendado: habilitar REST para schemas custom si se arma router manual; `Iniciar` ya lo hace para auth y app.

`utils.Handlers2`

- Propósito: registrar handlers HTTP extra.
- Campos: `Path string`, `H http.Handler`.
- Uso recomendado: pasar slice a `dataauth.Iniciar` para endpoints extra del consumidor.

`utils.NewRateLimiter(limit int, window time.Duration) *utils.RateLimiter`

- Propósito: crear rate limiter por cliente.
- Entrada: límite y ventana.
- Retorno: rate limiter.
- Uso recomendado: usar `RATE_LIMIT=1` con `Iniciar`, o montar manualmente si se requiere control custom.

### Usuarios

`usuarios.Crear(db *sql.DB, input model.NewUsuario, oauth_id *string) (*model.Usuario, error)`

- Propósito: crear usuario con roles, permisos y menús.
- Entrada: DB, input, OAuth ID opcional.
- Retorno: usuario creado.
- Uso recomendado: alta administrativa de usuarios.

`usuarios.Actualizar(db *sql.DB, input model.UpdateUsuario) (*model.Usuario, error)`

- Propósito: actualizar usuario y sus asignaciones.
- Entrada: DB, input.
- Retorno: usuario actualizado.
- Uso recomendado: mantenimiento administrativo.

`usuarios.UpdatePerfil(db *sql.DB, input model.UpdatePerfil) (*model.Usuario, error)`

- Propósito: actualizar datos de perfil.
- Entrada: DB, input.
- Retorno: usuario actualizado.
- Uso recomendado: edición de perfil propio.

`usuarios.GetUsuarios(db *sql.DB, query model.QueryUsuarios) ([]*model.Usuario, error)`

- Propósito: listar usuarios filtrando opcionalmente por rol.
- Entrada: DB, query.
- Retorno: usuarios.
- Uso recomendado: pantallas administrativas.

`usuarios.GetUsuariosConectados(db *sql.DB) ([]*model.Usuario, error)`

- Propósito: listar usuarios con conexiones activas.
- Entrada: DB.
- Retorno: usuarios.
- Uso recomendado: monitoreo.

`usuarios.GetById(db *sql.DB, id string) (*model.Usuario, error)`

- Propósito: obtener usuario simple.
- Entrada: DB, ID.
- Retorno: usuario.
- Uso recomendado: consultas internas.

`usuarios.GetBy(db *sql.DB, id string) (*model.ResponseUsuario, error)`

- Propósito: obtener usuario con roles, permisos y menús.
- Entrada: DB, ID.
- Retorno: usuario extendido.
- Uso recomendado: edición administrativa.

`usuarios.GetMe(db *sql.DB, input model.InputMe, userid string) (*model.ResponseMe, error)`

- Propósito: obtener usuario autenticado, roles, permisos sueltos y menús para una unidad.
- Entrada: DB, unidad, user ID.
- Retorno: datos completos de sesión.
- Uso recomendado: carga inicial de frontend después de login.

`usuarios.GetByUserPass(db *sql.DB, user, pass string) (*model.Usuario, error)`

- Propósito: validar credenciales.
- Entrada: username y password.
- Retorno: usuario.
- Uso recomendado: flujos de login internos.

`usuarios.UpdatePassword(db *sql.DB, id, pass string) (*model.Usuario, error)`

- Propósito: actualizar password.
- Entrada: user ID y password.
- Retorno: usuario.
- Uso recomendado: cambio de clave.

`usuarios.SetLastLogin(db *sql.DB, userid string)`

- Propósito: registrar último login.
- Entrada: DB, user ID.
- Retorno: ninguno.
- Uso recomendado: login custom.

`usuarios.CrearOauth(db *sql.DB, input model.NewUsuarioOauth, isportal bool) (*model.Usuario, error)`

- Propósito: crear usuario OAuth.
- Entrada: DB, input, flag portal.
- Retorno: usuario.
- Uso recomendado: integración OAuth.

`usuarios.CrearExterno(db *sql.DB, u, p string) (*model.Usuario, error)`

- Propósito: crear/sincronizar usuario desde sistema externo.
- Entrada: DB, username, password.
- Retorno: usuario.
- Uso recomendado: integración con auth externo.

### Login y Sesiones

`login.Login(ctx context.Context, db *sql.DB, input model.NewLogin, is_v2 bool) (*model.ResponseLogin, error)`

- Propósito: autenticar usuario con password cifrado/AES según flujo original.
- Entrada: contexto, DB, credenciales, flag v2.
- Retorno: sesión y datos `me`.
- Uso recomendado: usar resolver base `login`.

`login.Login2(ctx context.Context, db *sql.DB, input model.NewLogin2) (*model.ResponseLogin, error)`

- Propósito: autenticar usuario con input simplificado.
- Entrada: contexto, DB, credenciales.
- Retorno: sesión y datos `me`.
- Uso recomendado: usar resolver base `login_v2`.

`login.DesencriptarPassword(textoCifradoBase64 string, ivBase64 string) (string, error)`

- Propósito: descifrar password cliente.
- Entrada: texto cifrado base64 e IV base64.
- Retorno: password plano o error.
- Uso recomendado: compatibilidad con login original.

`sessionkey.CrearApikey(db *sql.DB, userid, apikey string, exp time.Time) (*model.SessionKey, error)`

- Propósito: persistir una sesión.
- Entrada: DB, user ID, JWT/API key, expiración.
- Retorno: session key.
- Uso recomendado: login custom.

`sessionkey.GetyKey(db *sql.DB, key string) (*model.SessionKey, error)`

- Propósito: recuperar sesión por key.
- Entrada: DB, session key.
- Retorno: sesión.
- Uso recomendado: middlewares y validaciones.

### Roles, Permisos, Menús y Unidades

`roles.Crear(db *sql.DB, input model.NewRol) (*model.Rol, error)`

- Propósito: crear rol y asociar permisos/menús.
- Entrada: DB, input.
- Retorno: rol completo.
- Uso recomendado: administración RBAC.

`roles.Actualizar(db *sql.DB, input model.UpdateRol) (*model.Rol, error)`

- Propósito: actualizar rol, permisos y menús.
- Entrada: DB, input.
- Retorno: rol completo.
- Uso recomendado: administración RBAC.

`roles.GetRoles(db *sql.DB) ([]*model.ResponseRoles, error)`

- Propósito: listar roles con contadores.
- Entrada: DB.
- Retorno: roles.
- Uso recomendado: pantallas de roles.

`roles.GetRolById(db *sql.DB, id string) (*model.Rol, error)`

- Propósito: obtener rol con permisos y menús.
- Entrada: DB, ID.
- Retorno: rol.
- Uso recomendado: edición de rol.

`roles.GetRolesByUsuario(db *sql.DB, userid string) ([]*model.ResponseRolMe, error)`

- Propósito: roles del usuario por unidad.
- Entrada: DB, user ID.
- Retorno: roles.
- Uso recomendado: respuesta `me`.

`roles.GetRolUnidadesByUser(db *sql.DB, userid string) ([]*model.ResponseRolUnidad, error)`

- Propósito: listar relaciones rol-unidad de un usuario.
- Entrada: DB, user ID.
- Retorno: relaciones.
- Uso recomendado: edición de usuario.

`permisos.GetPermisos(db *sql.DB) ([]*model.Permiso, error)`

- Propósito: listar permisos registrados.
- Entrada: DB.
- Retorno: permisos.
- Uso recomendado: administración RBAC.

`permisos.GetPermisosByRol(db *sql.DB, rol_id string) ([]*model.ResponsePermisoMe, error)`

- Propósito: permisos asignados a rol.
- Entrada: DB, rol ID.
- Retorno: permisos.
- Uso recomendado: detalle de rol.

`permisos.GetPermisosSueltosByUser(db *sql.DB, userid string) ([]*model.ResponsePermisoMe, error)`

- Propósito: permisos directos del usuario.
- Entrada: DB, user ID.
- Retorno: permisos.
- Uso recomendado: detalle de usuario.

`permisos.VerificarPermiso(db *sql.DB, userid, unidadid, metodo string) error`

- Propósito: validar permiso directo o por rol para usuario y unidad.
- Entrada: DB, user ID, unidad ID, método.
- Retorno: error si no tiene permiso.
- Uso recomendado: normalmente mediante `utils.CtxValue`.

`menus.Listar(db *sql.DB) ([]*model.Menus, error)`

- Propósito: listar menús.
- Entrada: DB.
- Retorno: menús.
- Uso recomendado: administración de navegación.

`menus.GetMenusbyRol(db *sql.DB, rol_id string) ([]*model.Menus, error)`

- Propósito: menús asociados a rol.
- Entrada: DB, rol ID.
- Retorno: menús.
- Uso recomendado: edición de rol.

`menus.MenusSueltos(db *sql.DB, userid string) ([]*model.Menus, error)`

- Propósito: menús directos del usuario.
- Entrada: DB, user ID.
- Retorno: menús.
- Uso recomendado: edición de usuario.

`menus.ListarByUserUnidad(db *sql.DB, input model.InputMe, userid string) ([]*model.Menus, error)`

- Propósito: menús visibles para usuario en una unidad.
- Entrada: DB, unidad, user ID.
- Retorno: menús.
- Uso recomendado: navegación del frontend autenticado.

`unidades.Crear(db *sql.DB, input model.NewUnidad) (*model.Unidad, error)`

- Propósito: crear unidad/sede/departamento.
- Entrada: DB, input.
- Retorno: unidad.
- Uso recomendado: administración de unidades.

`unidades.Actualizar(db *sql.DB, input model.UpdUnidad) (*model.Unidad, error)`

- Propósito: actualizar unidad.
- Entrada: DB, input.
- Retorno: unidad.
- Uso recomendado: administración de unidades.

`unidades.Listar(db *sql.DB) ([]*model.Unidad, error)`

- Propósito: listar unidades.
- Entrada: DB.
- Retorno: unidades.
- Uso recomendado: selección de unidad y administración.

`unidades.GetById(db *sql.DB, id string) (*model.Unidad, error)`

- Propósito: obtener unidad.
- Entrada: DB, ID.
- Retorno: unidad.
- Uso recomendado: consultas internas.

`unidades.GetFirtsByUser(db *sql.DB, userid string) (*model.Unidad, error)`

- Propósito: obtener primera unidad asignada a usuario.
- Entrada: DB, user ID.
- Retorno: unidad.
- Uso recomendado: defaults de login o navegación.

### Notificaciones, Tickets y Tiempo Real

`notificaciones.Crear(db *sql.DB, input model.NewNotificacion, userid string) (*model.Notificacion, error)`

- Propósito: crear aviso persistente.
- Entrada: DB, input, creador.
- Retorno: notificación.
- Uso recomendado: administración de avisos.

`notificaciones.Actualizar(db *sql.DB, input model.UpdNotificacion, userid string) (*model.Notificacion, error)`

- Propósito: actualizar aviso.
- Entrada: DB, input, usuario.
- Retorno: notificación.
- Uso recomendado: administración de avisos.

`notificaciones.Get(db *sql.DB, id string) (*model.Notificacion, error)`

- Propósito: obtener aviso.
- Entrada: DB, ID.
- Retorno: aviso.
- Uso recomendado: detalle.

`notificaciones.GetNotificacionesActivas(db *sql.DB) ([]*model.Notificacion, error)`

- Propósito: listar avisos vigentes.
- Entrada: DB.
- Retorno: avisos.
- Uso recomendado: mostrar avisos al iniciar sesión.

`xnotificaciones.InitializeGlobal()`

- Propósito: inicializar hub global de notificaciones.
- Entrada: ninguna.
- Retorno: ninguno.
- Uso recomendado: ya lo ejecuta `Iniciar`.

`xnotificaciones.GetGlobal() *xnotificaciones.Chan`

- Propósito: acceder al hub global.
- Entrada: ninguna.
- Retorno: canal global.
- Uso recomendado: emisiones custom.

`xnotificaciones.EnviarNotificacion(ctx context.Context, titulo string, datos *xnotificaciones.DataNotify) (bool, error)`

- Propósito: enviar notificación por WebSocket.
- Entrada: contexto, título, datos opcionales.
- Retorno: éxito/error.
- Uso recomendado: avisos en tiempo real desde módulos del consumidor.

`xnotificaciones.EnviarSSENotificacion(ctx context.Context, titulo string, datos *xnotificaciones.DataNotify) (bool, error)`

- Propósito: enviar notificación por SSE.
- Entrada: contexto, título, datos.
- Retorno: éxito/error.
- Uso recomendado: clientes SSE.

`xnotificaciones.NotificacionesSubs(ctx context.Context, userid string) (<-chan *model.XNotificacion, error)`

- Propósito: crear subscription GraphQL para usuario.
- Entrada: contexto, user ID.
- Retorno: canal de notificaciones.
- Uso recomendado: subscriptions custom.

`xnotificaciones.VerConexiones() (string, error)`

- Propósito: ver métricas de conexiones.
- Entrada: ninguna.
- Retorno: texto con conexiones.
- Uso recomendado: monitoreo.

`tickets.Crear(ctx context.Context, db *sql.DB, input model.NewTicket, userid string) (*model.Ticket, error)`

- Propósito: crear ticket de soporte.
- Entrada: contexto, DB, problema, user ID.
- Retorno: ticket.
- Uso recomendado: soporte en proyectos consumidores.

`tickets.Update(ctx context.Context, db *sql.DB, input model.NewTicketRespuesta, userid string) (*model.Ticket, error)`

- Propósito: responder ticket.
- Entrada: contexto, DB, respuesta, user ID.
- Retorno: ticket.
- Uso recomendado: flujo de soporte.

`tickets.Cerrar(ctx context.Context, db *sql.DB, id string) (*model.Ticket, error)`

- Propósito: cerrar ticket.
- Entrada: contexto, DB, ID.
- Retorno: ticket.
- Uso recomendado: cierre administrativo o del cliente.

`tickets.AllTickets(db *sql.DB, q model.QueryTickets) ([]*model.RespTickets, error)`

- Propósito: listar tickets con filtro por estado.
- Entrada: DB, query.
- Retorno: tickets.
- Uso recomendado: panel soporte.

`tickets.MisTickets(db *sql.DB, userid string) ([]*model.RespTickets, error)`

- Propósito: listar tickets del usuario.
- Entrada: DB, user ID.
- Retorno: tickets.
- Uso recomendado: portal de usuario.

`tickets.Get(ctx context.Context, db *sql.DB, id string) (*model.Ticket, error)`

- Propósito: ver ticket con respuestas.
- Entrada: contexto, DB, ID.
- Retorno: ticket.
- Uso recomendado: detalle de ticket.

`tickets.Respuestas(db *sql.DB, idticket string) ([]*model.TicketsRespuestas, error)`

- Propósito: listar respuestas de ticket.
- Entrada: DB, ticket ID.
- Retorno: respuestas.
- Uso recomendado: detalle de ticket.

### Archivos y Dashboard

`archivos.SubirImagen(img64, prefix, idbol string) (string, error)`

- Propósito: guardar imagen base64 como WebP reducido en `res`.
- Entrada: imagen base64, prefijo, ID.
- Retorno: path.
- Uso recomendado: fotos o imágenes del consumidor.

`archivos.GetImagen(url string) (string, error)`

- Propósito: leer imagen y devolver base64/data URL.
- Entrada: URL/path.
- Retorno: imagen base64.
- Uso recomendado: recuperar imágenes guardadas.

`dashboard.Reporte1(db *sql.DB) ([]*model.ResponseReporte1, error)`

- Propósito: reporte agregado base.
- Entrada: DB.
- Retorno: nombre/valor.
- Uso recomendado: dashboard administrativo base.

`dashboard.Reporte2(db *sql.DB) ([]*model.ResponseReporte2, error)`

- Propósito: reporte por fecha.
- Entrada: DB.
- Retorno: fecha/valor.
- Uso recomendado: dashboard administrativo base.

`dashboard.Reporte2b(db *sql.DB) ([]*model.ResponseReporte2b, error)`

- Propósito: reporte por mes.
- Entrada: DB.
- Retorno: mes/valor.
- Uso recomendado: dashboard administrativo base.

### Eventos

`utils.SetOnUserExternalCreate(callback utils.UserCreatedCallback)`

- Propósito: registrar callback cuando se crea usuario externo.
- Entrada: función `func(db *sql.DB, newUserID, userid, pwd string)`.
- Retorno: ninguno.
- Uso recomendado: enganchar lógica del consumidor después de alta externa.

`utils.SetOnUserRelogin(callback utils.UserCreatedCallback)`

- Propósito: registrar callback en relogin.
- Entrada: función callback.
- Retorno: ninguno.
- Uso recomendado: sincronización externa.

`utils.SetOnTicketCreated(callback utils.TicketCreatedCallback)`

- Propósito: registrar callback al crear ticket.
- Entrada: función `func(db *sql.DB, id string)`.
- Retorno: ninguno.
- Uso recomendado: notificaciones o integraciones del consumidor.

`utils.NotifyUserExternalCreated`, `utils.NotifyUserRelogin`, `utils.NotifyTicketCreated`

- Propósito: disparar callbacks registrados.
- Entrada: DB y datos del evento.
- Retorno: ninguno.
- Uso recomendado: normalmente los invoca la base.

## 5. Funcionalidades que NO Deben Reimplementarse

No reimplementes:

- Login, sesión, JWT, cookies ni validación de `SESSIONKEY`.
- Middleware de autenticación.
- RBAC por roles, permisos, unidades y usuarios.
- Verificación de permisos por operación.
- CRUD base de usuarios, roles, unidades, notificaciones y tickets.
- Listado/asignación de permisos.
- Sistema de menús base.
- GraphQL auth en `/query_auth`.
- WebSocket/SSE base de notificaciones.
- Conexión MySQL estándar.
- Adaptador REST hacia GraphQL.
- Carga/lectura básica de imágenes.
- Estructura SQL `rbac_*`.

Cuando el consumidor necesite una capacidad relacionada, debe reutilizar paquetes existentes y agregar solo la lógica específica del dominio.

## 6. Guía para Proyectos Consumidores

Para construir un nuevo módulo:

1. Define tipos, queries y mutations en el schema GraphQL del consumidor.
2. Ejecuta gqlgen para regenerar modelos y resolvers del consumidor.
3. Crea un paquete por dominio, por ejemplo `clientes`.
4. Divide la lógica en archivos pequeños:
   - `crear.go`
   - `listar.go`
   - `update.go`
   - `delete.go` si aplica
   - `utils.go` para parseos/validaciones privadas
5. En resolvers protegidos, valida sesión y permiso:

```go
tok, err := utils.CtxValue(ctx, r.DB, "create_cliente")
if err != nil {
    return nil, err
}
userid := tok.SessionKey.UsuarioID
```

6. Registra el permiso:

```sql
INSERT INTO rbac_permisos (metodo, nombre, grupo, descripcion)
VALUES ('create_cliente', 'crear cliente', 'clientes', 'Crear cliente');
```

7. Asocia permisos a roles:

```sql
INSERT INTO rbac_rol_permiso (rol_id, metodo)
VALUES (1, 'create_cliente');
```

8. Si el módulo debe aparecer en frontend, crea menú con ID >= 10:

```sql
INSERT INTO rbac_menus (id, label, path, icon, grupo, color, orden, padre_id)
VALUES (10, 'Clientes', '/clientes', 'groups', 2, 'primary', 1, NULL);
```

9. Asocia menú a rol:

```sql
INSERT INTO rbac_rol_menus (rol_id, menu_id)
VALUES (1, 10);
```

10. Usa transacciones cuando una operación escriba en varias tablas.
11. Usa `database/sql` y consultas parametrizadas con `?`.
12. Devuelve modelos GraphQL del consumidor o modelos base cuando corresponda.

Patrón de resolver recomendado:

```go
func (r *mutationResolver) CreateCliente(ctx context.Context, input model.NewCliente) (*model.Cliente, error) {
    tok, err := utils.CtxValue(ctx, r.DB, "create_cliente")
    if err != nil {
        return nil, err
    }
    return clientes.Crear(r.DB, input, tok.SessionKey.UsuarioID)
}
```

Patrón de servicio recomendado:

```go
func Crear(db *sql.DB, input model.NewCliente, userid string) (*model.Cliente, error) {
    tx, err := db.Begin()
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    res, err := tx.Exec(`insert into clientes(nombre, creado_por_id) values (?, ?)`, input.Nombre, userid)
    if err != nil {
        return nil, err
    }

    id, err := res.LastInsertId()
    if err != nil {
        return nil, err
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }

    return GetById(db, strconv.FormatInt(id, 10))
}
```

## 7. Buenas Prácticas

- Mantén los nombres de permisos sincronizados con los nombres de operaciones GraphQL.
- No omitas `utils.CtxValue` en resolvers que requieran autenticación.
- Para operaciones que solo requieren usuario autenticado y no permiso específico, usa `utils.CtxValue(ctx, db, "")`.
- No uses IDs de menú reservados por la base.
- No modifiques tablas `rbac_*` salvo para insertar permisos, menús o relaciones necesarias.
- No dupliques lógica de autenticación o RBAC en módulos del consumidor.
- Usa paquetes por dominio y funciones de servicio pequeñas.
- Mantén los resolvers delgados; deben validar contexto y delegar al paquete de dominio.
- Usa transacciones en operaciones compuestas.
- Cierra `rows` con `defer rows.Close()`.
- Maneja `sql.ErrNoRows` explícitamente cuando corresponda.
- Usa variables de entorno ya soportadas por la base.
- Reutiliza `utils.Handlers2` para endpoints HTTP adicionales.
- Reutiliza `xnotificaciones` para eventos en tiempo real.
- Reutiliza `archivos.SubirImagen` para imágenes base64.
- Reutiliza `dataauth.Iniciar` como punto único de montaje del servidor.

## Variables de Entorno Requeridas

El `.env` del consumidor debe incluir al menos:

```text
PORT=
DB_USER=
DB_PASS=
DB_HOST=
DB_NAME=
PERM_EXTERNO=
EXTERNAL_AUTH=
EXTERNAL_ME=
PLAYGROUND=
RATE_LIMIT=
DECODE_PASS_KEY=
TOKEN_DURATION_MIN=
AUTH_SHOW_NAME_PERMISO=
SEND_NOTI_LOGIN=
DEFAULT_UNIDAD_OAUTH=
DEFAULT_ROL_OAUTH=
DEFAULT_ROL_EXTER=
OAUTH_EMAILS_PERM=
DB_CONN_LIFETIME_MIN=
DB_MAX_OPEN=
DB_MAX_IDLE=
ALLOWED_ORIGINS=
JWT_SECRET=
```

`JWT_SECRET` tiene fallback interno, pero debe definirse explícitamente en producción.

## Endpoints Base

- `POST /query_auth`: GraphQL de autenticación y administración base.
- `GET /ws`: WebSocket para subscriptions base.
- `POST /query`: GraphQL del consumidor.
- `GET /ws_app`: WebSocket para subscriptions del consumidor.
- `GET /sse`: Server-Sent Events.
- `GET /auth`: playground auth si `PLAYGROUND=1`.
- `GET /app`: playground app si `PLAYGROUND=1`.
- `/res/*`: archivos estáticos.
- `/rest_auth/query/{operationName}` y `/rest_auth/mutation/{operationName}`: REST hacia GraphQL auth.
- `/rest/query/{operationName}` y `/rest/mutation/{operationName}`: REST hacia GraphQL app.

# Especificación del Proyecto Consumidor

A partir de este punto debes desarrollar el siguiente proyecto consumidor utilizando exclusivamente las herramientas y patrones disponibles en esta librería base:

[DESCRIPCIÓN DEL PROYECTO AQUÍ]
