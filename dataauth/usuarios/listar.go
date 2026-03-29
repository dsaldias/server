package usuarios

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/menus"
	"github.com/dsaldias/server/dataauth/permisos"
	"github.com/dsaldias/server/dataauth/roles"
	"github.com/dsaldias/server/dataauth/xnotificaciones"
)

var WRONG_PASS = "usuario o clave incorrectos"

func GetUsuarios(db *sql.DB, query model.QueryUsuarios) ([]*model.Usuario, error) {
	filter_by_rol := ""
	if query.Rol != nil {
		filter_by_rol = "where id in (select usuario_id from rbac_rol_usuario_unidades where rol_id='%s')"
		filter_by_rol = fmt.Sprintf(filter_by_rol, *query.Rol)
	}

	sql := `select id, nombres,apellido1,apellido2,documento,celular,correo,sexo,direccion,estado,username,last_login,oauth_id,foto_url,ST_X(ubicacion) AS latitud,ST_Y(ubicacion) AS longitud,fecha_registro,fecha_update from rbac_usuarios %s`
	sql = fmt.Sprintf(sql, filter_by_rol)
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cha := xnotificaciones.GetGlobal()

	us := []*model.Usuario{}
	for rows.Next() {
		u := model.Usuario{}
		er := parseRows(rows, &u)
		if er != nil {
			return nil, er
		}

		cons := len(cha.GetSubsByUser(u.ID))
		u.Conexiones = int32(cons)
		us = append(us, &u)
	}

	return us, nil
}

func GetUsuariosConectados(db *sql.DB) ([]*model.Usuario, error) {
	cha := xnotificaciones.GetGlobal()
	ids := cha.IdsConectados()
	if len(ids) == 0 {
		return nil, errors.New("no hay usuarios conectados")
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		select id, nombres,apellido1,apellido2,documento,celular,correo,sexo,direccion,estado,username,last_login,oauth_id,foto_url,ST_X(ubicacion) AS latitud,ST_Y(ubicacion) AS longitud,fecha_registro,fecha_update 
		from rbac_usuarios
		where id in (%s)
		order by last_login desc
	`, strings.Join(placeholders, ","))

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	us := []*model.Usuario{}
	for rows.Next() {
		u := model.Usuario{}
		er := parseRows(rows, &u)
		if er != nil {
			return nil, er
		}

		cons := len(cha.GetSubsByUser(u.ID))
		u.Conexiones = int32(cons)
		us = append(us, &u)
	}

	return us, nil
}

func GetById(db *sql.DB, id string) (*model.Usuario, error) {
	sq := `
		select 
		u.id,
		u.nombres, 
		u.apellido1, 
		u.apellido2, 
		u.documento,
		u.celular,
		u.correo,
		u.sexo,
		u.direccion,
		u.estado,
		u.username,
		u.last_login,
		u.oauth_id,
		u.foto_url,
		ST_X(u.ubicacion) AS latitud, 
		ST_Y(u.ubicacion) AS longitud,  
		u.fecha_registro,
		u.fecha_update
		from rbac_usuarios u 
		where u.id = ?`

	row := db.QueryRow(sq, id)

	us := model.Usuario{}
	err := parseRow(row, &us)

	if err == sql.ErrNoRows {
		return nil, errors.New("usuario no encontrado por id")
	}
	if err != nil {
		return nil, err
	}

	return &us, nil
}

func GetBy(db *sql.DB, id string) (*model.ResponseUsuario, error) {
	sq := `
		select 
		u.id,
		u.nombres, 
		u.apellido1, 
		u.apellido2, 
		u.documento,
		u.celular,
		u.correo,
		u.sexo,
		u.direccion,
		u.estado,
		u.username,
		u.last_login,
		u.oauth_id,
		u.foto_url,
		ST_X(u.ubicacion) AS latitud, 
		ST_Y(u.ubicacion) AS longitud,  
		u.fecha_registro,
		u.fecha_update
		from rbac_usuarios u 
		where u.id = ?`

	row := db.QueryRow(sq, id)

	us := model.ResponseUsuario{}
	err := parseRowRU(row, &us)

	if err == sql.ErrNoRows {
		return nil, errors.New("usuario no encontrado por id")
	}
	if err != nil {
		return nil, err
	}

	us.MenusSueltos, err = menus.MenusSueltos(db, id)
	if err != nil {
		return nil, err
	}

	us.PermisosSueltos, err = permisos.GetPermisosSueltosByUser(db, id)
	if err != nil {
		return nil, err
	}

	us.Roles, err = roles.GetRolUnidadesByUser(db, id)
	if err != nil {
		return nil, err
	}

	return &us, nil
}

func GetByUserPass(db *sql.DB, user, pass string) (*model.Usuario, error) {
	sq := `
		select 
		u.id,
		u.nombres, 
		u.apellido1, 
		u.apellido2, 
		u.documento,
		u.celular,
		u.correo,
		u.sexo,
		u.direccion,
		u.estado,
		u.username,
		u.last_login,
		u.oauth_id,
		u.foto_url,
		ST_X(u.ubicacion) AS latitud, 
		ST_Y(u.ubicacion) AS longitud,  
		u.fecha_registro,
		u.fecha_update
		from rbac_usuarios u 
		where u.username = ? 
		and u.password = SHA2( ?, 256)`

	row := db.QueryRow(sq, user, pass)

	us := model.Usuario{}
	err := parseRow(row, &us)

	if err == sql.ErrNoRows {
		return nil, errors.New(WRONG_PASS)
	}
	if err != nil {
		return nil, err
	}

	return &us, nil
}

func GetIdByUsernamePortal(db *sql.DB, username string) (string, error) {
	xsql := `select id from rbac_usuarios where username=?`
	id := ""
	err := db.QueryRow(xsql, username).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return id, nil
}
