package test

import (
	"database/sql"
	"net/http"
	"strconv"
)

type Controller struct {
	DB *sql.DB
}

type TestRow struct {
	ID       int
	Nombre   string
	Username string
	Estado   string
}

type PageData struct {
	Pagina    int
	Size      int
	Total     int
	TotalPage int
	Rows      []TestRow
}

var sampleRows = []TestRow{
	{ID: 1, Nombre: "Ana Pérez", Username: "ana.perez", Estado: "Activo"},
	{ID: 2, Nombre: "Bruno Salas", Username: "bruno.salas", Estado: "Activo"},
	{ID: 3, Nombre: "Carla Rojas", Username: "carla.rojas", Estado: "Inactivo"},
	{ID: 4, Nombre: "Diego Vargas", Username: "diego.vargas", Estado: "Activo"},
	{ID: 5, Nombre: "Elena Flores", Username: "elena.flores", Estado: "Activo"},
	{ID: 6, Nombre: "Felipe Cruz", Username: "felipe.cruz", Estado: "Inactivo"},
	{ID: 7, Nombre: "Gabriela León", Username: "gabriela.leon", Estado: "Activo"},
	{ID: 8, Nombre: "Hugo Medina", Username: "hugo.medina", Estado: "Activo"},
	{ID: 9, Nombre: "Irene Molina", Username: "irene.molina", Estado: "Activo"},
	{ID: 10, Nombre: "Jorge Pardo", Username: "jorge.pardo", Estado: "Inactivo"},
	{ID: 11, Nombre: "Karen Silva", Username: "karen.silva", Estado: "Activo"},
	{ID: 12, Nombre: "Luis Torres", Username: "luis.torres", Estado: "Activo"},
}

func (c *Controller) Listar(w http.ResponseWriter, r *http.Request) {
	data := pageData(r)
	if err := Index(data).Render(r.Context(), w); err != nil {
		http.Error(w, "No fue posible renderizar el módulo de prueba", http.StatusInternalServerError)
	}
}

func (c *Controller) ListaHTML(w http.ResponseWriter, r *http.Request) {
	data := pageData(r)
	if err := TestTable(data).Render(r.Context(), w); err != nil {
		http.Error(w, "No fue posible renderizar la lista de prueba", http.StatusInternalServerError)
	}
}

func pageData(r *http.Request) PageData {
	page := 1
	if value, err := strconv.Atoi(r.URL.Query().Get("pagina")); err == nil && value > 0 {
		page = value
	}
	size := 10
	if value, err := strconv.Atoi(r.URL.Query().Get("size")); err == nil && value > 0 {
		size = value
	}

	total := len(sampleRows)
	totalPages := (total + size - 1) / size
	if page > totalPages {
		page = totalPages
	}
	start := (page - 1) * size
	end := start + size
	if end > total {
		end = total
	}

	return PageData{Pagina: page, Size: size, Total: total, TotalPage: totalPages, Rows: sampleRows[start:end]}
}

func pageURL(page, size int) string {
	return "/webc/test?pagina=" + strconv.Itoa(page) + "&size=" + strconv.Itoa(size)
}

func testFirstRow(data PageData) string {
	if data.Total == 0 || len(data.Rows) == 0 {
		return "0"
	}
	return strconv.Itoa((data.Pagina-1)*data.Size + 1)
}

func testLastRow(data PageData) string {
	if len(data.Rows) == 0 {
		return "0"
	}
	return strconv.Itoa((data.Pagina-1)*data.Size + len(data.Rows))
}

func testStatusClass(status string) string {
	if status == "Activo" {
		return "inline-flex rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300"
	}
	return "inline-flex rounded-full bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground"
}
