package dataauthevents

import (
	"database/sql"

	"github.com/dsaldias/server/dataauth/utils"
)

/*
Aqui puede setear sus funciones cuando ocurre alguna accion
por ejemplo cuando se registra un usuario externo
defina su escuchador
*/
func LoadCustomEvents() {
	// EDIT THIS FILE IN YOUR APP
	utils.SetOnUserExternalCreate(func(db *sql.DB, id, u, p string) {})
	utils.SetOnUserRelogin(func(db *sql.DB, id, u, p string) {})
	utils.SetOnTicketCreated(func(db *sql.DB, id string) {})
}
