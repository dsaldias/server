---
name: add-file-handling
description: Add file upload and/or download to an entity, reusing archivos.SubirImagen and archivos.GetImagen from the auth server. Use when the user wants to add image/file upload to an entity or resolver.
---

# Add File Handling to Entity

The user wants to add file upload/download to: **$ARGUMENTS**

This project reuses `dataauth/archivos` from the auth server (`github.com/dsaldias/server/dataauth/archivos`).

## Available functions

```go
// Upload: receives base64 data URL, saves as WebP, returns file path
func SubirImagen(img64, prefix, idbol string) (string, error)
// img64:  base64 data URL (e.g. "data:image/jpeg;base64,...")
// prefix: file name prefix (e.g. "perfil", "producto", "documento")
// idbol:  unique identifier appended to filename (e.g. user ID)
// returns: path like "res/producto-42.webp"

// Download: reads file from disk, returns base64 data URI
func GetImagen(url string) (string, error)
// url: path returned by SubirImagen (e.g. "res/producto-42.webp")
// returns: "data:image/webp;base64,..." or placeholder if not found
```

Constraints:
- Max 2MB per file
- Accepted MIME types: jpeg, png, bmp, webp
- Output always WebP (converted automatically)
- Files stored in `res/` directory at project root

## Pattern: upload during Crear

```go
// In dataauth/<entity>/crear.go
import "<module>/dataauth/archivos"

func Crear(db *sql.DB, input model.New<Entity>) (*model.<Entity>, error) {
    // ... insert row first to get the ID ...
    id, _ := res.LastInsertId()
    idStr := fmt.Sprintf("%d", id)

    // Upload image after insert (need the ID for the filename)
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

## Pattern: upload during Actualizar

```go
// In dataauth/<entity>/update.go
func Actualizar(db *sql.DB, input model.Update<Entity>) (*model.<Entity>, error) {
    if input.Imagen != nil && *input.Imagen != "" {
        path, err := archivos.SubirImagen(*input.Imagen, "<entity>", input.ID)
        if err != nil {
            return nil, fmt.Errorf("error subiendo imagen: %w", err)
        }
        db.Exec(`UPDATE <table> SET imagen = ? WHERE id = ?`, path, input.ID)
    }
    // ... update other fields ...
    return GetById(db, input.ID)
}
```

## Pattern: download via resolver

Add a dedicated query resolver for fetching the file:

```go
// In graph_auth/schema.resolvers.go
func (r *queryResolver) Get<Entity>Imagen(ctx context.Context, url string) (string, error) {
    _, err := utils.CtxValue(ctx, r.DB, "get_<entity>_imagen")
    if err != nil {
        return "", err
    }
    return archivos.GetImagen(url)
}
```

GraphQL schema addition:
```graphql
# In Query type:
get_<entity>_imagen(url: String!): String!
```

## GraphQL input — add imagen field

```graphql
input New<Entity> {
  # ... existing fields ...
  imagen: String   # base64 data URL, optional
}

input Update<Entity> {
  id: ID!
  # ... existing fields ...
  imagen: String   # base64 data URL, optional — only update if provided
}

type <Entity> {
  # ... existing fields ...
  imagen: String   # file path stored in DB (use get_<entity>_imagen to fetch)
}
```

## SQL column

```sql
ALTER TABLE `<table>` ADD COLUMN `imagen` varchar(255) DEFAULT NULL;
```

## Important notes

- **Never store base64 in DB** — store only the file path returned by `SubirImagen`
- The `imagen` field in GraphQL input is the base64 data URL from the client
- The `imagen` field in GraphQL type is the file path — clients call `get_<entity>_imagen` to retrieve it
- `SubirImagen` overwrites the file if called with same prefix+id, so updates are safe
