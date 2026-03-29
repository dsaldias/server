package unidades

import (
	"database/sql"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/usuarios"
)

func Actualizar(db *sql.DB, input model.UpdUnidad) (*model.Unidad, error) {
	point, err := usuarios.Ubicacion(input.Latitud, input.Longitud)
	if err != nil {
		return nil, err
	}
	sql := `update rbac_unidades set nombre=?, descripcion=?, ubicacion=ST_GeomFromText(?), orden=? where id=?`
	_, err = db.Exec(sql, input.Nombre, input.Descripcion, point, input.Orden, input.ID)
	if err != nil {
		return nil, err
	}
	return GetById(db, input.ID)
}
