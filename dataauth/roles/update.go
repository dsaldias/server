package roles

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/dsaldias/server/graph_auth/model"
)

func Actualizar(db *sql.DB, input model.UpdateRol) (*model.Rol, error) {
	sql := `update rbac_roles set nombre=?, descripcion=?,jerarquia=? where id=?`
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(sql, input.Nombre, input.Descripcion, input.Jerarquia, input.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec("delete from rbac_rol_permiso where rol_id = ?", input.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	sql = "replace into rbac_rol_permiso(rol_id, metodo) values %s"
	places := make([]string, len(input.Permisos))
	args := make([]interface{}, len(input.Permisos)*2)

	for i, p := range input.Permisos {
		places[i] = "(?,?)"
		args[i*2] = input.ID
		args[i*2+1] = p
	}

	sql = fmt.Sprintf(sql, strings.Join(places, ", "))
	_, err = tx.Exec(sql, args...)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// ====

	_, err = tx.Exec("delete from rbac_rol_menus where rol_id = ?", input.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	sql = "replace into rbac_rol_menus(rol_id, menu_id) values %s"
	places = make([]string, len(input.Menus))
	args = make([]interface{}, len(input.Menus)*2)

	for i, p := range input.Menus {
		places[i] = "(?,?)"
		args[i*2] = input.ID
		args[i*2+1] = p
	}

	sql = fmt.Sprintf(sql, strings.Join(places, ", "))
	_, err = tx.Exec(sql, args...)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// ====

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	r, err := GetRolById(db, input.ID)
	if err != nil {
		return nil, err
	}

	return r, nil
}
