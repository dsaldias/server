package utils

import (
	"database/sql"
	"sync"
)

type UserCreatedCallback func(db *sql.DB, newUserID, userid, pwd string)
type TicketCreatedCallback func(db *sql.DB, id string)

var (
	userCreatedCallbacks   []UserCreatedCallback
	userReloginCallbacks   []UserCreatedCallback
	ticketCreatedCallbacks []TicketCreatedCallback
	userMu                 sync.Mutex
	ticketMu               sync.Mutex
)

// SetOnUserExternalCreate registra un callback para creación de usuarios.
func SetOnUserExternalCreate(callback UserCreatedCallback) {
	userMu.Lock()
	defer userMu.Unlock()
	userCreatedCallbacks = append(userCreatedCallbacks, callback)
}

func SetOnUserRelogin(callback UserCreatedCallback) {
	userMu.Lock()
	defer userMu.Unlock()
	userReloginCallbacks = append(userReloginCallbacks, callback)
}

// SetOnTicketCreated registra un callback para creación de tickets.
func SetOnTicketCreated(callback TicketCreatedCallback) {
	ticketMu.Lock()
	defer ticketMu.Unlock()
	ticketCreatedCallbacks = append(ticketCreatedCallbacks, callback)
}

// NotifyUserExternalCreated ejecuta todos los callbacks de usuarios.
func NotifyUserExternalCreated(db *sql.DB, userID, u, p string) {
	userMu.Lock()
	defer userMu.Unlock()
	for _, cb := range userCreatedCallbacks {
		safeExecuteCallback(func() { cb(db, userID, u, p) })
	}
}

func NotifyUserRelogin(db *sql.DB, userID, u, p string) {
	userMu.Lock()
	defer userMu.Unlock()
	for _, cb := range userReloginCallbacks {
		safeExecuteCallback(func() { cb(db, userID, u, p) })
	}
}

// NotifyTicketCreated ejecuta todos los callbacks de tickets.
func NotifyTicketCreated(db *sql.DB, id string) {
	ticketMu.Lock()
	defer ticketMu.Unlock()
	for _, cb := range ticketCreatedCallbacks {
		safeExecuteCallback(func() { cb(db, id) })
	}
}

// safeExecuteCallback evita que un panic en un callback afecte a los demás.
func safeExecuteCallback(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			// Loggear el error si es necesario
		}
	}()
	fn()
}
