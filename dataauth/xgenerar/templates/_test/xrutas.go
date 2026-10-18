package test

import (
	"database/sql"
	"net/http"

	"github.com/dsaldias/server/dataauth/utilidad"
)

const (
	Path      = "/webc/test"
	ListaHTML = Path + "/listahtml"
)

func Rutas(db *sql.DB) []*utilidad.Handlers2 {
	controller := Controller{DB: db}
	return []*utilidad.Handlers2{{Path: Path, H: http.HandlerFunc(controller.Listar)}, {Path: ListaHTML, H: http.HandlerFunc(controller.ListaHTML)}}
}
