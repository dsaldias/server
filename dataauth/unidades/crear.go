package unidades

import (
	"database/sql"
	"strconv"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/usuarios"
)

func Crear(db *sql.DB, input model.NewUnidad) (*model.Unidad, error) {
	point, err := usuarios.Ubicacion(input.Latitud, input.Longitud)
	if err != nil {
		return nil, err
	}
	sql := `insert into unidades(nombre,descripcion,ubicacion,orden) values(?,?,ST_GeomFromText(?),?)`
	res, err := db.Exec(sql, input.Nombre, input.Descripcion, point, input.Orden)
	if err != nil {
		return nil, err
	}
	idd, _ := res.LastInsertId()
	id := strconv.FormatInt(idd, 10)
	return GetById(db, id)
}
