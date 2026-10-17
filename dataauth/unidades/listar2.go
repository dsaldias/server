package unidades

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/dsaldias/server/dataauth/utilidad"
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

	where := ""
	args := make([]any, 0, 3)
	if q.Nombre != nil {
		nombre := strings.TrimSpace(*q.Nombre)
		if nombre != "" {
			where = " where nombre like ?"
			args = append(args, "%"+nombre+"%")
		}
	}

	query := `select
	id, 
	nombre, 
	descripcion, 
	orden,
	ST_X(ubicacion) AS latitud, 
	ST_Y(ubicacion) AS longitud, 
	fecha_registro 
from rbac_unidades` + where + ` order by id, orden
LIMIT ? OFFSET ?
	`
	args = append(args, q.Size, offset)
	rows, err := db.Query(query, args...)
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
	queryTotal := `select count(*) from rbac_unidades` + where
	countArgs := args[:len(args)-2]
	if err := db.QueryRow(queryTotal, countArgs...).Scan(&total); err != nil {
		return nil, err
	}
	paginas := utilidad.CalcularPaginas(total, int32(q.Size))
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
