package roles

import (
	"database/sql"
	"errors"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/menus"
	"github.com/dsaldias/server/dataauth/permisos"
)

func GetRoles(db *sql.DB) ([]*model.ResponseRoles, error) {
	sql := `
	SELECT  
		r.id,r.nombre,r.descripcion,r.jerarquia,r.fecha_registro,
		COUNT(DISTINCT rm.id) AS total_menus,
		COUNT(DISTINCT rp.metodo) AS total_permisos,
		COUNT(DISTINCT ru.usuario_id) AS total_usuarios
	FROM
		rbac_roles r
	LEFT JOIN 
		rbac_rol_menus rm ON r.id = rm.rol_id
	LEFT JOIN 
		rbac_rol_permiso rp ON r.id = rp.rol_id
	LEFT JOIN 
		rbac_rol_usuario_unidades ru ON r.id = ru.rol_id
	GROUP BY 
		r.id,r.nombre,r.descripcion,r.jerarquia,r.fecha_registro
	order by r.jerarquia asc, r.id;
	`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rs := []*model.ResponseRoles{}

	for rows.Next() {
		r := model.ResponseRoles{}
		er := parseRes(rows, &r)
		if er != nil {
			return nil, er
		}
		// r.FechaRegistro = utils.ToTZ(r.FechaRegistro)
		rs = append(rs, &r)
	}

	return rs, nil
}

func GetRolById(db *sql.DB, id string) (*model.Rol, error) {
	sq := "select id,nombre,descripcion,jerarquia,fecha_registro from rbac_roles where id = ?"
	row := db.QueryRow(sq, id)
	r := model.Rol{}
	err := parseRow(row, &r)
	if err == sql.ErrNoRows {
		return nil, errors.New("rol no existente")
	}

	r.Permisos, err = permisos.GetPermisosByRol(db, r.ID)
	if err != nil {
		return nil, err
	}
	r.Menus, err = menus.GetMenusbyRol(db, r.ID)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func GetRolesByUsuario(db *sql.DB, userid string) ([]*model.ResponseRolMe, error) {
	sql := `
	select 
	r.id as rol_id,
	r.nombre as rol_nombre,
	r.descripcion as rol_descripcion,
	r.jerarquia as rol_jerarquia,
	r.fecha_registro as rol_fecha_registro,
	un.id as unidad_id,
	un.nombre as unidad_nombre,
	un.descripcion as unidad_descripcion,
	un.orden as unidad_orden,
	un.fecha_registro as unidad_fecha_registro
	from rbac_roles r
	left join rbac_rol_usuario_unidades ruu on ruu.rol_id = r.id
	inner join rbac_unidades un on un.id = ruu.unidad_id 
	where ruu.usuario_id = ?
	order by r.jerarquia
	`
	rows, err := db.Query(sql, userid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rs := []*model.ResponseRolMe{}

	for rows.Next() {
		r := model.ResponseRolMe{Rol: &model.RolMe{}, Unidad: &model.Unidad{}}
		er := parseRolUnidad(rows, &r)
		if er != nil {
			return nil, er
		}
		rs = append(rs, &r)
	}

	return rs, nil
}

func GetRolUnidadesByUser(db *sql.DB, userid string) ([]*model.ResponseRolUnidad, error) {
	sql := `
	select 
	r.id as rol_id,
	r.nombre as rol_nombre,
	un.id as unidad_id,
	un.nombre as unidad_nombre
	from rbac_rol_usuario_unidades ruu 
	inner join rbac_roles r on r.id = ruu.rol_id 
	inner join rbac_unidades un on un.id = ruu.unidad_id 
	where ruu.usuario_id = ?
	`
	rows, err := db.Query(sql, userid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rs := []*model.ResponseRolUnidad{}
	for rows.Next() {
		ru := model.ResponseRolUnidad{}
		er := parseReRolUnidad(rows, &ru)
		if er != nil {
			return nil, er
		}

		rs = append(rs, &ru)
	}

	return rs, nil
}
