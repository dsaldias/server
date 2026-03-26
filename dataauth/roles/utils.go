package roles

import (
	"database/sql"
	"errors"

	"github.com/dsaldias/server/graph_auth/model"
)

func parseRow(row *sql.Row, t *model.Rol) error {
	return row.Scan(
		&t.ID,
		&t.Nombre,
		&t.Descripcion,
		&t.Jerarquia,
		&t.FechaRegistro,
	)
}

func parseRes(rows *sql.Rows, t *model.ResponseRoles) error {
	return rows.Scan(
		&t.ID,
		&t.Nombre,
		&t.Descripcion,
		&t.Jerarquia,
		&t.FechaRegistro,
		&t.Menus,
		&t.Permisos,
		&t.Usuarios,
	)
}

func parseRolUnidad(rows *sql.Rows, t *model.ResponseRolMe) error {
	return rows.Scan(
		&t.Rol.ID,
		&t.Rol.Nombre,
		&t.Rol.Descripcion,
		&t.Rol.Jerarquia,
		&t.Rol.FechaRegistro,
		&t.Unidad.ID,
		&t.Unidad.Nombre,
		&t.Unidad.Descripcion,
		&t.Unidad.Orden,
		&t.Unidad.FechaRegistro,
	)
}

func parseReRolUnidad(rows *sql.Rows, t *model.ResponseRolUnidad) error {
	return rows.Scan(
		&t.RolID,
		&t.RolName,
		&t.UnidadID,
		&t.UnidadName,
	)
}

func validar_campos(input model.NewRol) error {
	if len(input.Nombre) > 50 {
		return errors.New("nombre excede los 50 caracteres")
	}

	if input.Descripcion != nil && len(*input.Descripcion) > 100 {
		return errors.New("descripcion excede los 100 caracteres")
	}
	return nil
}
