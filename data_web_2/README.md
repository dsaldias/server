# Frontend SPA sin framework

Este directorio es un frontend independiente del backend. No usa Templ,
Datastar, React, Vue ni otro framework web: solo HTML, CSS y JavaScript
moderno con módulos ES.

## Estructura

- `index.html`: documento raíz de la SPA.
- `assets/js/api.js`: único cliente GraphQL y catálogo central de operaciones.
- `assets/js/router.js`: router del navegador.
- `assets/js/store.js`: estado compartido de sesión y contexto.
- `assets/js/components.js`: componentes HTML mínimos y utilidades visuales.
- `assets/js/main.js`: composición de páginas y eventos de la aplicación.
- `assets/css/app.css`: tokens visuales, layout y componentes CSS.

## Contrato con el backend

- La comunicación se hace mediante `POST /query_auth` con JSON GraphQL.
- Las cookies se envían con `credentials: include`.
- Las consultas y mutaciones se declaran en `assets/js/api.js`.
- Las páginas no deben llamar `fetch` directamente.
- La autenticación, autorización y validación siguen siendo responsabilidad de
  `dataauth`.

## Ejecución

El servidor se activa únicamente cuando existe `RUN_WEB_SERVER_PORT`:

```env
RUN_WEB_SERVER_PORT=3000
PORT=8038
ALLOWED_ORIGINS=http://localhost:3000
```

Con esa configuración, la SPA se sirve en `http://localhost:3000` y el API en
`http://localhost:8038`. `server.go` expone `config.js` dinámicamente para que
el frontend conozca la URL del API. También se puede definir explícitamente
`AUTH_API_URL` cuando el API está en otro host o detrás de un proxy.

## Reglas para ampliar el frontend

1. Agregar nuevas operaciones GraphQL únicamente en `api.js`.
2. Mantener las páginas como funciones que devuelven HTML y consumen el cliente
   centralizado.
3. Reutilizar `pageHead`, `table`, `modal`, `formField`, `status` y `toast`
   antes de crear una variante.
4. Usar eventos delegados en `main.js`; no agregar listeners duplicados después
   de cada render.
5. Escapar valores externos con `esc` antes de insertarlos en HTML.
6. No guardar tokens en `localStorage`; la sesión se mantiene mediante cookie.
7. Mantener estados de carga, error y vacío en cada página.
