package permisos

import (
	"database/sql"
	"errors"
	"os"

	"github.com/dsaldias/server/graph_auth/model"
)

func GetPermisos(db *sql.DB) ([]*model.Permiso, error) {
	sql := `select metodo,nombre, descripcion,grupo,fecha_registro from rbac_permisos order by grupo,fecha_registro asc`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ps := []*model.Permiso{}
	for rows.Next() {
		p := model.Permiso{}
		er := parseRows(rows, &p)
		if er != nil {
			return nil, er
		}
		// p.FechaRegistro = utils.ToTZ(p.FechaRegistro)
		ps = append(ps, &p)
	}

	return ps, nil
}

func GetPermisosByRol(db *sql.DB, rol_id string) ([]*model.ResponsePermisoMe, error) {
	sql := `
	select p.metodo, p.nombre, p.descripcion,p.grupo, p.fecha_registro, rp.fecha_registro as fecha_asignado 
	from rbac_permisos p
	left join rbac_rol_permiso rp on rp.metodo = p.metodo
	where rp.rol_id = ?;
	`

	rows, err := db.Query(sql, rol_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perms := []*model.ResponsePermisoMe{}
	for rows.Next() {
		p := model.ResponsePermisoMe{}
		er := parse(rows, &p)
		if er != nil {
			return nil, er
		}
		perms = append(perms, &p)
	}

	return perms, nil
}

func GetPermisosSueltosByUser(db *sql.DB, userid string) ([]*model.ResponsePermisoMe, error) {
	sql := `
	select p.metodo, p.nombre, p.descripcion,p.grupo, p.fecha_registro, up.fecha_registro as fecha_asignado 
	from rbac_permisos p
	inner join rbac_usuario_permiso up on up.metodo  = p.metodo 
	where up.usuario_id = ?
	`

	rows, err := db.Query(sql, userid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perms := []*model.ResponsePermisoMe{}
	for rows.Next() {
		p := model.ResponsePermisoMe{}
		er := parse(rows, &p)
		if er != nil {
			return nil, er
		}
		perms = append(perms, &p)
	}

	return perms, nil
}

func VerificarPermiso(db *sql.DB, userid, unidadid, metodo string) error {
	sq := `
	SELECT 
		CASE 
			WHEN up.usuario_id IS NOT NULL THEN 'Directo'
			WHEN rp.rol_id IS NOT NULL THEN 'A través de rbac_roles' 
		END AS metodo_de_asignacion
	FROM rbac_usuarios u
	LEFT JOIN rbac_usuario_permiso up ON u.id = up.usuario_id AND up.metodo = ?
	LEFT JOIN rbac_rol_usuario_unidades ruu ON u.id = ruu.usuario_id
	LEFT JOIN rbac_rol_permiso rp ON ruu.rol_id = rp.rol_id AND rp.metodo = ?
	WHERE u.id = ? and ruu.unidad_id = ? AND (up.usuario_id IS NOT NULL OR rp.rol_id IS NOT NULL);
	`
	texto := ""
	err := db.QueryRow(sq, metodo, metodo, userid, unidadid).Scan(&texto)

	if err == sql.ErrNoRows {
		show_name := os.Getenv("AUTH_SHOW_NAME_PERMISO")
		if show_name == "1" {
			return errors.New("no tiene permiso para: " + metodo)
		} else {
			return errors.New("no tiene permiso")
		}
	}
	return err
}
