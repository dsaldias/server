package usuarios

import (
	"database/sql"
	"errors"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/menus"
	"github.com/dsaldias/server/dataauth/permisos"
	"github.com/dsaldias/server/dataauth/roles"
)

func GetMe(db *sql.DB, input model.InputMe, userid string) (*model.ResponseMe, error) {
	us, err := GetById(db, userid)
	if err != nil {
		return nil, err
	}

	user := model.ResponseMe{}
	user.Usuario = us

	user.Roles, err = roles.GetRolesByUsuario(db, us.ID)
	if err != nil {
		return nil, errors.Join(err, errors.New("error al cargar rbac_roles al usuario"))
	}

	user.PermisosSueltos, err = permisos.GetPermisosSueltosByUser(db, us.ID)
	if err != nil {
		return nil, errors.Join(err, errors.New("error al cargar rbac_permisos sueltos del usuario"))
	}

	user.Menus, err = menus.ListarByUserUnidad(db, input, us.ID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
