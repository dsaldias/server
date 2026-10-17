package xutility

import (
	"strconv"
	"strings"

	"github.com/dsaldias/server/graph_auth/model"
)

var AssetVersion = "dev"

func connectedUserName(user *model.Usuario) string {
	if user == nil {
		return "Usuario"
	}
	lastName := ""
	if user.Apellido2 != nil {
		lastName = *user.Apellido2
	}
	name := strings.TrimSpace(strings.Join([]string{user.Nombres, user.Apellido1, lastName}, " "))
	if name == "" {
		return user.Username
	}
	return name
}

func connectedUserCount(count int32) string {
	if count == 1 {
		return "1 conexión"
	}
	return strconv.Itoa(int(count)) + " conexiones"
}

func connectedUserUniqueCount(users []*model.Usuario) string {
	return strconv.Itoa(len(users))
}

func connectedUserTotalCount(users []*model.Usuario) string {
	var total int32
	for _, user := range users {
		if user != nil {
			total += user.Conexiones
		}
	}
	return strconv.Itoa(int(total))
}
