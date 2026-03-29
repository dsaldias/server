package roles

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/dsaldias/server/graph_auth/model"
)

func Crear(db *sql.DB, input model.NewRol) (*model.Rol, error) {
	if err := validar_campos(input); err != nil {
		return nil, err
	}
	sql := `insert into rbac_roles(nombre,descripcion,jerarquia) values (?,?,?)`
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	res, err := tx.Exec(sql, input.Nombre, input.Descripcion, input.Jerarquia)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	idd, _ := res.LastInsertId()
	id := strconv.FormatInt(idd, 10)

	sql = "insert into rbac_rol_permiso(rol_id, metodo) values %s"
	places := make([]string, len(input.Permisos))
	args := make([]interface{}, len(input.Permisos)*2)

	for i, p := range input.Permisos {
		places[i] = "(?,?)"
		args[i*2] = id
		args[i*2+1] = p
	}

	sql = fmt.Sprintf(sql, strings.Join(places, ", "))
	_, err = tx.Exec(sql, args...)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// ================================================
	sql = "insert into rbac_rol_menus(rol_id, menu_id) values %s"
	places = make([]string, len(input.Menus))
	args = make([]interface{}, len(input.Menus)*2)

	for i, m := range input.Menus {
		places[i] = "(?,?)"
		args[i*2] = id
		args[i*2+1] = m
	}

	sql = fmt.Sprintf(sql, strings.Join(places, ", "))
	_, err = tx.Exec(sql, args...)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// =================================================

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	r, err := GetRolById(db, id)
	if err != nil {
		return nil, err
	}

	return r, nil
}
