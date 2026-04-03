---
name: new-entity
description: Create a new entity (database table) following the dataauth/ CRUD pattern — folder with crear.go, listar.go, update.go, delete.go, utils.go plus GraphQL schema. Use when the user asks to add a new table, entity, model, or resource to the project.
---

# New Entity Scaffolding

The user wants to create a new entity: **$ARGUMENTS**

This project uses `github.com/dsaldias/server` as its RBAC auth backend. Follow the exact same patterns from that server's `dataauth/` package.

## Step 1 — Read context first

Before generating anything:
1. Read `go.mod` to get this project's module name
2. Ask the user for the DB table columns if not provided, or make reasonable assumptions and state them

## Step 2 — Create `app/<entity>/` folder with these files

### `utils.go`
```go
package <entity>

import (
    "database/sql"
    "<this-module>/graph/model"
)

func parseRow(row *sql.Row, t *model.<Entity>) error {
    return row.Scan(
        &t.ID,
        // ... fields matching SELECT order
    )
}

func parseRows(rows *sql.Rows, t *model.<Entity>) error {
    return rows.Scan(
        &t.ID,
        // ... fields matching SELECT order
    )
}
```

### `listar.go`
```go
package <entity>

import (
    "database/sql"
    "<this-module>/graph/model"
)

func Get<Entities>(db *sql.DB) ([]*model.<Entity>, error) {
    rows, err := db.Query(`SELECT id, ... FROM <table> WHERE activo = 1`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var result []*model.<Entity>
    for rows.Next() {
        t := &model.<Entity>{}
        if err := parseRows(rows, t); err != nil {
            return nil, err
        }
        result = append(result, t)
    }
    return result, nil
}

func GetById(db *sql.DB, id string) (*model.<Entity>, error) {
    row := db.QueryRow(`SELECT id, ... FROM <table> WHERE id = ?`, id)
    t := &model.<Entity>{}
    if err := parseRow(row, t); err != nil {
        return nil, err
    }
    return t, nil
}
```

### `crear.go`
```go
package <entity>

import (
    "database/sql"
    "fmt"
    "<this-module>/graph/model"
)

func Crear(db *sql.DB, input model.New<Entity>) (*model.<Entity>, error) {
    if err := validarCampos(input); err != nil {
        return nil, err
    }
    res, err := db.Exec(
        `INSERT INTO <table> (...) VALUES (?, ?, ...)`,
        input.Field1, input.Field2,
    )
    if err != nil {
        return nil, err
    }
    id, err := res.LastInsertId()
    if err != nil {
        return nil, err
    }
    return GetById(db, fmt.Sprintf("%d", id))
}

func validarCampos(input model.New<Entity>) error {
    if input.RequiredField == "" {
        return fmt.Errorf("campo requerido: required_field")
    }
    return nil
}
```

### `update.go`
```go
package <entity>

import (
    "database/sql"
    "<this-module>/graph/model"
)

func Actualizar(db *sql.DB, input model.Update<Entity>) (*model.<Entity>, error) {
    _, err := db.Exec(
        `UPDATE <table> SET field1 = ?, field2 = ? WHERE id = ?`,
        input.Field1, input.Field2, input.ID,
    )
    if err != nil {
        return nil, err
    }
    return GetById(db, input.ID)
}
```

### `delete.go`
```go
package <entity>

import (
    "database/sql"
    "fmt"
)

func Eliminar(db *sql.DB, id string) (bool, error) {
    _, err := db.Exec(`UPDATE <table> SET activo = 0 WHERE id = ?`, id)
    if err != nil {
        return false, fmt.Errorf("error al eliminar: %w", err)
    }
    return true, nil
}
```

> Soft delete (`activo = 0`) por defecto salvo que el usuario pida hard delete.

## Step 3 — Add GraphQL schema

### Agregar tipos en `graph/schema.graphqls` (o un archivo dedicado `graph/<entity>.graphqls`)
```graphql
type <Entity> {
  id: ID!
  field1: String!
  field2: String
  activo: Boolean!
}

input New<Entity> {
  field1: String!
  field2: String
}

input Update<Entity> {
  id: ID!
  field1: String
  field2: String
}
```

### Agregar operaciones al tipo Query/Mutation
```graphql
# Query:
<entities>: [<Entity>!]!
<entity>_by_id(id: ID!): <Entity>

# Mutation:
create_<entity>(input: New<Entity>!): <Entity>!
update_<entity>(input: Update<Entity>!): <Entity>!
delete_<entity>(id: ID!): Boolean!
```

## Step 4 — Regenerar gqlgen

```bash
go run github.com/99designs/gqlgen generate
```

> **NUNCA editar `graph/models_gen.go` manualmente** — es auto-generado y se sobreescribe con este comando. Para cambiar tipos, editar los archivos `.graphqls` y regenerar.

## Step 5 — Agregar resolvers

**REGLA OBLIGATORIA**: todo resolver debe comenzar con el check de permiso.

```go
import (
    "<this-module>/app/<entity>"
    "github.com/dsaldias/server/dataauth/utils"
)

func (r *queryResolver) <Entities>(ctx context.Context) ([]*model.<Entity>, error) {
    _, err := utils.CtxValue(ctx, r.DB, "<entities>")
    if err != nil {
        return nil, err
    }
    return <entity>.Get<Entities>(r.DB)
}

func (r *mutationResolver) Create<Entity>(ctx context.Context, input model.New<Entity>) (*model.<Entity>, error) {
    _, err := utils.CtxValue(ctx, r.DB, "create_<entity>")
    if err != nil {
        return nil, err
    }
    return <entity>.Crear(r.DB, input)
}

func (r *mutationResolver) Update<Entity>(ctx context.Context, input model.Update<Entity>) (*model.<Entity>, error) {
    _, err := utils.CtxValue(ctx, r.DB, "update_<entity>")
    if err != nil {
        return nil, err
    }
    return <entity>.Actualizar(r.DB, input)
}

func (r *mutationResolver) Delete<Entity>(ctx context.Context, id string) (bool, error) {
    _, err := utils.CtxValue(ctx, r.DB, "delete_<entity>")
    if err != nil {
        return nil, err
    }
    return <entity>.Eliminar(r.DB, id)
}
```

## Step 6 — Migración SQL

Agregar la tabla al archivo `sqls/database-<modulo>.sql` del proyecto (hay un único archivo SQL por proyecto). Para determinar el nombre exacto del archivo, leer `go.mod` y tomar el último segmento del módulo.

Convenciones:
- Nombre de tabla: prefijo `app_` + nombre en español en plural (ej: `app_productos`)
- Siempre incluir: `id`, `activo`, `fecha_registro`
- Charset: `utf8mb4 COLLATE utf8mb4_unicode_ci`

```sql
CREATE TABLE IF NOT EXISTS `app_<entidad>` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `campo1` varchar(255) NOT NULL,
  `campo2` varchar(255) DEFAULT NULL,
  `activo` tinyint NOT NULL DEFAULT 1,
  `fecha_registro` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

Agregar también los permisos correspondientes:
```sql
INSERT INTO `rbac_permisos` (`metodo`, `nombre`, `grupo`, `descripcion`)
VALUES
('<entidades>', 'listar <entidades>', '<entidades>', 'Lista todos los <entidades>'),
('create_<entidad>', 'crear <entidad>', '<entidades>', 'Crea un <entidad>'),
('update_<entidad>', 'actualizar <entidad>', '<entidades>', 'Actualiza un <entidad>'),
('delete_<entidad>', 'eliminar <entidad>', '<entidades>', 'Elimina un <entidad>');
```
```sql
CREATE TABLE IF NOT EXISTS `<table>` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `field1` varchar(255) NOT NULL,
  `field2` varchar(255) DEFAULT NULL,
  `activo` tinyint(1) NOT NULL DEFAULT 1,
  `fecha_registro` timestamp NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## Convenciones de nombres

| Concepto | Convención | Ejemplo |
|----------|-----------|---------|
| Package | minúsculas, singular | `package productos` |
| Carpeta | minúsculas, igual que tabla sin prefijo | `app/productos/` |
| Tabla SQL | prefijo `app_` + plural en español | `app_productos` |
| Columnas SQL | en español | `nombre`, `descripcion`, `activo` |
| Crear | `Crear` | `func Crear(db, input)` |
| Actualizar | `Actualizar` | `func Actualizar(db, input)` |
| Listar | `Get<Entidades>` | `func GetProductos(db)` |
| Uno por ID | `GetById` | `func GetById(db, id)` |
| Eliminar | `Eliminar` | `func Eliminar(db, id)` |
| Parse fila | `parseRow` (no exportado) | `func parseRow(row, t)` |
| Parse filas | `parseRows` (no exportado) | `func parseRows(rows, t)` |
| Nombre permiso | nombre de operación GraphQL | `"create_producto"` |
