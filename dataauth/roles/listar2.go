package roles

import (
	"database/sql"
	"errors"

	"github.com/dsaldias/server/dataauth/utils"
	"github.com/dsaldias/server/graph_auth/model"
)

func GetRoles2(db *sql.DB, q model.QueryRoles) (*model.RolPaginado, error) {
	if q.Pagina < 1 {
		return nil, errors.New("la página debe ser superior a 0")
	}
	if q.Size < 1 {
		return nil, errors.New("el tamaño debe ser superior a 0")
	}

	offset := (q.Pagina - 1) * q.Size

	sql := `
	SELECT  
		r.id,r.nombre,r.descripcion,r.jerarquia,r.fecha_registro,
		COUNT(DISTINCT rm.id) AS total_menus,
		COUNT(DISTINCT rp.metodo) AS total_permisos,
		COUNT(DISTINCT ru.usuario_id) AS total_usuarios
	FROM
		rbac_roles r
	LEFT JOIN 
		rbac_rol_menus rm ON r.id = rm.rol_id
	LEFT JOIN 
		rbac_rol_permiso rp ON r.id = rp.rol_id
	LEFT JOIN 
		rbac_rol_usuario_unidades ru ON r.id = ru.rol_id
	GROUP BY 
		r.id,r.nombre,r.descripcion,r.jerarquia,r.fecha_registro
	order by r.jerarquia asc, r.id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(sql, q.Size, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rs := []*model.ResponseRoles{}

	for rows.Next() {
		r := model.ResponseRoles{}
		er := parseRes(rows, &r)
		if er != nil {
			return nil, er
		}
		rs = append(rs, &r)
	}

	var total int32
	queryTotal := ` select count(*) from rbac_roles `
	if err := db.QueryRow(queryTotal).Scan(&total); err != nil {
		return nil, err
	}
	paginas := utils.CalcularPaginas(total, int32(q.Size))
	pagina := model.Pagina{
		Pagina:    q.Pagina,
		Size:      q.Size,
		Paginas:   paginas,
		Registros: total,
	}

	dt := model.RolPaginado{
		Paginacion: &pagina,
		Datos:      rs,
	}

	return &dt, nil
}
