package xnotificaciones

import "fmt"

func VerConexiones() (string, error) {
	cha := GetGlobal()
	/* cha.Mu.Lock()
	defer cha.Mu.Unlock() */

	t := fmt.Sprintf("-> %+v\n", cha.subscriptores)
	return t, nil
}
