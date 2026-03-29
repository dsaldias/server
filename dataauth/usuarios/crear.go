package usuarios

import (
	"database/sql"
	"strconv"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/archivos"
)

func Crear(db *sql.DB, input model.NewUsuario, oauth_id *string) (*model.Usuario, error) {
	if err := validar_campos(input); err != nil {
		return nil, err
	}
	if err := permisos_obligatorios(input.Roles, input.PermisosSueltos); err != nil {
		return nil, err
	}
	if err := validarCadena(input.Username, "username"); err != nil {
		return nil, err
	}
	if err := validarCadena(input.Password, "password"); err != nil {
		return nil, err
	}
	point, err := Ubicacion(input.Latitud, input.Longitud)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	sql := `
	INSERT INTO rbac_usuarios(nombres, apellido1, apellido2, documento, celular, correo, sexo, direccion, username, password,oauth_id,ubicacion)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, SHA2(?, 256),?, ST_GeomFromText(?));
	`
	res, err := tx.Exec(sql,
		input.Nombres,
		input.Apellido1,
		input.Apellido2,
		input.Documento,
		input.Celular,
		input.Correo,
		input.Sexo,
		input.Direccion,
		input.Username,
		input.Password,
		oauth_id,
		point,
	)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	id, _ := res.LastInsertId()
	xid := strconv.FormatInt(id, 10)

	// asignar rbac_roles
	err = asignarRoles(tx, input.Roles, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// fin asignar rbac_roles

	// asignar rbac_permisos sueltos
	if len(input.PermisosSueltos) > 0 {
		err = asignarPermisos(tx, input.PermisosSueltos, id)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	// fin rbac_permisos sueltos

	// asignar rbac_menus sueltos
	if len(input.PermisosSueltos) > 0 {
		err = asignarMenus(tx, input.MenusSueltos, id)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	// fin rbac_menus sueltos

	// subir foto perfil
	if input.Foto64 != nil && len(*input.Foto64) > 0 {
		foto_url, err := archivos.SubirImagen(*input.Foto64, "perfil", xid)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		sql = `update rbac_usuarios set foto_url=? where id=?`
		_, err = tx.Exec(sql, foto_url, id)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	// fin subir foto perfil

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return GetById(db, strconv.FormatInt(id, 10))
}
