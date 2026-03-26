package login

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/sessionkey"
	"github.com/dsaldias/server/dataauth/unidades"
	"github.com/dsaldias/server/dataauth/usuarios"
	"github.com/dsaldias/server/dataauth/utils"
	"github.com/dsaldias/server/dataauth/xnotificaciones"
)

func Login2(ctx context.Context, db *sql.DB, input model.NewLogin2) (*model.ResponseLogin, error) {
	inp := model.NewLogin{
		Username: input.Username,
		Password: input.Password,
	}
	return Login(ctx, db, inp, true)
}

func Login(ctx context.Context, db *sql.DB, input model.NewLogin, is_v2 bool) (*model.ResponseLogin, error) {
	if !is_v2 {
		pwd, err := DesencriptarPassword(input.Password, input.Iv64)
		if err != nil {
			return nil, err
		}
		input.Password = pwd
	}

	us, err := usuarios.GetByUserPass(db, input.Username, input.Password)
	if err != nil {
		// login portal
		ext := os.Getenv("PERM_EXTERNO")
		if err.Error() == usuarios.WRONG_PASS && ext == "1" {
			u, err := CrearExterno(db, input.Username, input.Password)
			if err != nil {
				return nil, err
			}
			us = u
			// fin login portal
		} else {
			return nil, err
		}
	}
	if !us.Estado {
		return nil, errors.New("usuario no activo")
	}

	token, exp, minus, err := utils.GenerateToken(ctx, us.ID)
	if err != nil {
		return nil, err
	}

	sesion, err := sessionkey.CrearApikey(db, us.ID, token, exp)
	if err != nil {
		return nil, err
	}

	un, err := unidades.GetFirtsByUser(db, us.ID)
	if err != nil {
		return nil, err
	}

	usuarios.SetLastLogin(db, us.ID)

	inp := model.InputMe{}
	inp.UnidadID = un.ID
	me, err := usuarios.GetMe(db, inp, us.ID)
	if err != nil {
		return nil, err
	}

	res := model.ResponseLogin{}
	res.SessionKey = sesion.Key
	res.SessionTime = minus
	res.Me = me

	utils.NotifyUserRelogin(db, us.ID, us.Username, input.Password)

	send := os.Getenv("SEND_NOTI_LOGIN")
	if send == "1" {
		user := fmt.Sprintf("%s %s", us.Nombres, us.Apellido1)
		xnotificaciones.EnviarNotificacion(ctx, user+" ha accedido al sistema", nil)
	}

	// funcionalidad nueva para cookie
	utils.CtxSetCookie(ctx, sesion.Key, exp)

	return &res, nil
}
