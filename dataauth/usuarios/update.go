package usuarios

import (
	"database/sql"
	"strconv"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/archivos"
)

func Actualizar(db *sql.DB, input model.UpdateUsuario) (*model.Usuario, error) {
	/* if err := validar_campos(input); err != nil {
		return nil, err
	} */
	if err := permisos_obligatorios(input.Roles, input.PermisosSueltos); err != nil {
		return nil, err
	}
	us, err := GetById(db, input.ID)
	if err != nil {
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
	update usuarios set 
	nombres=?, 
	apellido1=?, 
	apellido2=?, 
	documento=?, 
	celular=?, 
	correo=?, 
	sexo=?, 
	direccion=?,
	ubicacion=ST_GeomFromText(?) 
	where id= ? 
	`
	_, err = tx.Exec(sql,
		input.Nombres,
		input.Apellido1,
		input.Apellido2,
		input.Documento,
		input.Celular,
		input.Correo,
		input.Sexo,
		input.Direccion,
		point,
		input.ID,
	)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if input.Password != nil && len(*input.Password) > 0 && (us.OauthID == nil || len(*us.OauthID) == 0) {
		_, err = tx.Exec("update usuarios set password=SHA2(?, 256) where id = ?", input.Password, input.ID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if len(input.Username) > 0 && (us.OauthID == nil || len(*us.OauthID) == 0) {
		_, err = tx.Exec("update usuarios set username=? where id = ?", input.Username, input.ID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	_, err = tx.Exec("delete from rol_usuario_unidades where usuario_id = ?", input.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec("delete from usuario_permiso where usuario_id = ?", input.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec("delete from menus_usuario where usuario_id = ?", input.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	id, _ := strconv.ParseInt(input.ID, 10, 64)

	// asignar roles
	err = asignarRoles(tx, input.Roles, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// fi asignar roles

	// asignar permisos sueltos
	if len(input.PermisosSueltos) > 0 {
		err = asignarPermisos(tx, input.PermisosSueltos, id)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	// fin permisos sueltos

	// asignar menus sueltos
	if len(input.MenusSueltos) > 0 {
		err = asignarMenus(tx, input.MenusSueltos, id)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	// fin menus sueltos

	// subir foto perfil
	if input.Foto64 != nil && len(*input.Foto64) > 0 {
		foto_url, err := archivos.SubirImagen(*input.Foto64, "perfil", input.ID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		sql = `update usuarios set foto_url=? where id=?`
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

func UpdatePassword(db *sql.DB, id, pass string) (*model.Usuario, error) {
	sql := `update usuarios set password=SHA2(?, 256) where id = ?`
	_, err := db.Exec(sql, pass, id)
	if err != nil {
		return nil, err
	}
	return GetById(db, id)
}

func SetLastLogin(db *sql.DB, userid string) {
	sql := "update usuarios set last_login=CURRENT_TIMESTAMP where id = ?"
	db.Exec(sql, userid)
}
