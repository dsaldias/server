package usuarios

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/archivos"
)

func UpdatePerfil(db *sql.DB, input model.UpdatePerfil) (*model.Usuario, error) {
	user, err := GetById(db, input.ID)
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
	update rbac_usuarios set 
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

	if user.OauthID == nil {
		if input.Username != nil && len(*input.Username) > 0 {
			_, err = tx.Exec("update rbac_usuarios set username=? where id=?", input.Username, input.ID)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
		}

		if input.Password != nil && len(*input.Password) > 0 {
			_, err = tx.Exec("update rbac_usuarios set password=SHA2(?, 256) where id = ?", input.Password, input.ID)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	// subir foto perfil
	if input.Foto64 != nil {
		foto_url, err := archivos.SubirImagen(*input.Foto64, "perfil", input.ID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		sql = `update rbac_usuarios set foto_url=? where id=?`
		_, err = tx.Exec(sql, foto_url, input.ID)
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

	return GetById(db, input.ID)
}
