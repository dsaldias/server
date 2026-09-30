package unidades

import (
	"database/sql"
	"errors"

	"github.com/dsaldias/server/graph_auth/model"
)

func ListarV2(db *sql.DB, q model.QueryUnidades) (*model.UnidadPaginada, error) {
	if q.Pagina < 1 {
		return nil, errors.New("la página debe ser superior a 0")
	}
	if q.Size < 1 {
		return nil, errors.New("el tamaño debe ser superior a 0")
	}

	offset := (q.Pagina - 1) * q.Size

	sql := `select 
	id, 
	nombre, 
	descripcion, 
	orden,
	ST_X(ubicacion) AS latitud, 
	ST_Y(ubicacion) AS longitud, 
	fecha_registro 
	from rbac_unidades order by id, orden
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(sql, q.Size, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rs := []*model.Unidad{}
	for rows.Next() {
		u := model.Unidad{}
		er := parseRows(rows, &u)
		if er != nil {
			return nil, er
		}
		rs = append(rs, &u)
	}

	var total int32
	queryTotal := ` select count(*) from rbac_unidades `
	if err := db.QueryRow(queryTotal).Scan(&total); err != nil {
		return nil, err
	}
	paginas := calcularPaginas(total, int32(q.Size))
	pagina := model.Pagina{
		Pagina:    q.Pagina,
		Size:      q.Size,
		Paginas:   paginas,
		Registros: total,
	}

	dt := model.UnidadPaginada{
		Paginacion: &pagina,
		Datos:      rs,
	}

	return &dt, nil
}

func calcularPaginas(total, size int32) int32 {
	if size <= 0 {
		return 0
	}
	return (total + size - 1) / size
}
