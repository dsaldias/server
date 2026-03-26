package usuarios

import (
	"database/sql"
	"os"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/utils"
)

func CrearOauth(db *sql.DB, input model.NewUsuarioOauth, isportal bool) (*model.Usuario, error) {

	if err := oauth_emails_permitidos(input.Correo); err != nil {
		return nil, err
	}

	var id, oauth *string
	xsql := "select id, oauth_id from usuarios where oauth_id=?"
	err := db.QueryRow(xsql, input.Username).Scan(&id, &oauth)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if id != nil {
		return GetById(db, *id)
	}

	rol := os.Getenv("DEFAULT_ROL_OAUTH")
	uni := os.Getenv("DEFAULT_UNIDAD_OAUTH")
	nombres, aps := splitName(input.Nombres)
	if !isportal {
		rol = os.Getenv("DEFAULT_ROL_EXTER")
	}

	data := model.NewUsuario{}
	data.Nombres = cut_string(nombres, 30)
	data.Apellido1 = cut_string(aps, 30)
	data.Celular = input.Celular
	data.Correo = input.Correo
	data.Username = input.Username
	data.Password = input.Password
	data.Roles = []*model.RolUnidad{{RolID: rol, UnidadID: uni}}

	if isportal {
		id_existe := ""
		xsql := "select id from usuarios where username=?"
		err := db.QueryRow(xsql, input.Username).Scan(&id_existe)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		if id_existe != "" {
			sqlx := "update usuarios set password=SHA2(?, 256) where id=?"
			_, err = db.Exec(sqlx, input.Password, id_existe)
			if err != nil {
				return nil, err
			}
			return GetById(db, id_existe)
		}
	}

	newus, er := Crear(db, data, &data.Username)
	if er != nil {
		return nil, er
	}

	if isportal {
		utils.NotifyUserExternalCreated(db, newus.ID, newus.Username, input.Password)
	}

	return newus, nil
}
