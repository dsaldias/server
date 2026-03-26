---
name: add-file-handling
description: Add file upload and/or download to an entity, reusing archivos.SubirImagen and archivos.GetImagen from github.com/dsaldias/server. Use when the user wants to add image/file upload to an entity or resolver.
---

# Add File Handling to Entity

The user wants to add file upload/download to: **$ARGUMENTS**

Este proyecto reutiliza `dataauth/archivos` del servidor auth (`github.com/dsaldias/server/dataauth/archivos`).

## Funciones disponibles

```go
import "github.com/dsaldias/server/dataauth/archivos"

// Sube imagen: recibe base64 data URL, guarda como WebP, retorna el path
archivos.SubirImagen(img64, prefix, idbol string) (string, error)
// img64:  data URL base64 (ej: "data:image/jpeg;base64,...")
// prefix: prefijo del nombre de archivo (ej: "producto", "perfil")
// idbol:  identificador único (ej: ID del registro)
// retorna: path como "res/producto-42.webp"

// Descarga imagen: lee del disco, retorna base64 data URI
archivos.GetImagen(url string) (string, error)
// url: path retornado por SubirImagen
// retorna: "data:image/webp;base64,..." o imagen placeholder si no existe
```

Restricciones:
- Máximo 2MB por archivo
- MIME types aceptados: jpeg, png, bmp, webp
- Salida siempre WebP (conversión automática)
- Archivos guardados en directorio `res/` en la raíz del proyecto

## Patrón: upload al crear

```go
// dataauth/<entity>/crear.go
import "github.com/dsaldias/server/dataauth/archivos"

func Crear(db *sql.DB, input model.New<Entity>) (*model.<Entity>, error) {
    // ... insertar fila primero para obtener el ID ...
    id, _ := res.LastInsertId()
    idStr := fmt.Sprintf("%d", id)

    if input.Imagen != nil && *input.Imagen != "" {
        path, err := archivos.SubirImagen(*input.Imagen, "<entity>", idStr)
        if err != nil {
            return nil, fmt.Errorf("error subiendo imagen: %w", err)
        }
        db.Exec(`UPDATE <table> SET imagen = ? WHERE id = ?`, path, id)
    }

    return GetById(db, idStr)
}
```

## Patrón: upload al actualizar

```go
// dataauth/<entity>/update.go
func Actualizar(db *sql.DB, input model.Update<Entity>) (*model.<Entity>, error) {
    if input.Imagen != nil && *input.Imagen != "" {
        path, err := archivos.SubirImagen(*input.Imagen, "<entity>", input.ID)
        if err != nil {
            return nil, fmt.Errorf("error subiendo imagen: %w", err)
        }
        db.Exec(`UPDATE <table> SET imagen = ? WHERE id = ?`, path, input.ID)
    }
    // ... actualizar otros campos ...
    return GetById(db, input.ID)
}
```

## Patrón: descarga via resolver

```go
// graph/schema.resolvers.go
import (
    "github.com/dsaldias/server/dataauth/archivos"
    "github.com/dsaldias/server/dataauth/utils"
)

func (r *queryResolver) Get<Entity>Imagen(ctx context.Context, url string) (string, error) {
    _, err := utils.CtxValue(ctx, r.DB, "get_<entity>_imagen")
    if err != nil {
        return "", err
    }
    return archivos.GetImagen(url)
}
```

Schema GraphQL:
```graphql
# En Query:
get_<entity>_imagen(url: String!): String!
```

## Campos GraphQL para imagen

```graphql
input New<Entity> {
  # ... campos existentes ...
  imagen: String   # base64 data URL, opcional
}

input Update<Entity> {
  id: ID!
  # ... campos existentes ...
  imagen: String   # base64 data URL, opcional — solo actualiza si se provee
}

type <Entity> {
  # ... campos existentes ...
  imagen: String   # path guardado en DB — usar get_<entity>_imagen para obtener el base64
}
```

## Columna SQL

```sql
ALTER TABLE `<table>` ADD COLUMN `imagen` varchar(255) DEFAULT NULL;
```

## Reglas importantes

- **Nunca guardar base64 en DB** — guardar solo el path retornado por `SubirImagen`
- El campo `imagen` en el input GraphQL es el base64 data URL que envía el cliente
- El campo `imagen` en el tipo GraphQL es el path — el cliente llama `get_<entity>_imagen` para obtener el base64
- `SubirImagen` sobreescribe si se llama con mismo prefix+id, los updates son seguros
