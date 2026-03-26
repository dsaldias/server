package utils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/permisos"
	"github.com/dsaldias/server/dataauth/sessionkey"

	"github.com/dgrijalva/jwt-go"
)

type JwtCustomClaim struct {
	USERID string `json:"id"`
	jwt.StandardClaims
}

type AuthData struct {
	Clains     *JwtCustomClaim
	SessionKey *model.SessionKey
	UnidadID   string
	RolID      string
	TOKEN      string `json:"token"`
}

var jwtSecret = []byte(getJwtSecret())

func getJwtSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "aSecret"
	}
	return secret
}

func jwtGenerate(userID string, tim time.Time) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &JwtCustomClaim{
		USERID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: tim.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	})

	token, err := t.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return token, nil
}

func JwtValidate(token string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(token, &JwtCustomClaim{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("there's a problem with the signing method")
		}
		return jwtSecret, nil
	})
}

type authString string

func AuthMiddleware(db *sql.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// auth := r.Header.Get("Authorization")
			unidad := r.Header.Get("UNIDAD")
			rol := r.Header.Get("ROL")
			skey := r.Header.Get("SESSIONKEY")

			// funcionalidad de cookie
			if skey == "" {
				cookie, err := r.Cookie("galletita_traviesa")
				if err == nil {
					skey = cookie.Value
				}
			}

			sk, er := sessionkey.GetyKey(db, skey)
			if er != nil {
				next.ServeHTTP(w, r)
				return
			}
			auth := sk.Apikey

			if auth == "" || len(auth) <= 7 {
				next.ServeHTTP(w, r)
				return
			}

			validate, err := JwtValidate(auth)
			if err != nil {
				txt := err.Error()
				if !strings.HasPrefix(txt, "token is expired by") {
					next.ServeHTTP(w, r)
					return
				}
			}

			customClaim, _ := validate.Claims.(*JwtCustomClaim)

			/* fecha := time.Unix(customClaim.ExpiresAt, 0)
			formato := fecha.Format("02/01/2006 15:04")
			fmt.Println(formato) */

			data := AuthData{}
			data.Clains = customClaim
			data.TOKEN = auth
			data.UnidadID = unidad
			data.RolID = rol
			data.SessionKey = sk

			ctx := context.WithValue(r.Context(), authString("auth"), &data)

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)

		})
	}
}

func GenerateToken(ctx context.Context, userID string) (string, time.Time, int32, error) {
	tokenduration := os.Getenv("TOKEN_DURATION_MIN")
	dur, err := strconv.Atoi(tokenduration)
	if err != nil {
		dur = 10
	}
	tim := time.Now().Add(time.Minute * time.Duration(dur))
	tok, err := jwtGenerate(userID, tim)
	if err != nil {
		return "", tim, 0, err
	}
	return tok, tim, int32(dur), nil
}

func CtxValue(ctx context.Context, db *sql.DB, metodo string) (*AuthData, error) {
	str := authString("auth")
	algo := ctx.Value(str)
	if algo == nil {
		return nil, errors.New("proporcione un token")
	}
	clains, _ := algo.(*AuthData)
	if clains == nil {
		return nil, errors.New("debes iniciar session")
	}
	validate, err := JwtValidate(clains.TOKEN)
	if err != nil || !validate.Valid {
		txt := err.Error()
		if strings.HasPrefix(txt, "token is expired by") {
			txt = strings.Replace(txt, "token is expired by", "Su sessión expiró hace ", 1)
			return nil, errors.New(txt)
		} else {
			return nil, errors.New(txt)
		}
	}
	if clains.SessionKey == nil {
		return nil, errors.New("no hay session key")
	}
	if !clains.SessionKey.UserEstado {
		return nil, errors.New("tu cuenta se encuentra suspendida")
	}

	if len(clains.UnidadID) == 0 {
		return nil, errors.New("falta la unidad en el header")
	}

	if len(metodo) > 0 {
		err = permisos.VerificarPermiso(db, clains.SessionKey.UsuarioID, clains.UnidadID, metodo)
		if err != nil {
			return nil, err
		}
	}

	return clains, nil
}

func MiddlewareCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "responseWriterCookie", w)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CtxSetCookie(ctx context.Context, token string, exp time.Time) {
	w := ctx.Value("responseWriterCookie").(http.ResponseWriter)
	http.SetCookie(w, &http.Cookie{
		Name:     "galletita_traviesa",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // true en producción con HTTPS
		// SameSite: http.SameSiteLaxMode,
		SameSite: http.SameSiteNoneMode, // front y back en dominios diferentes
		Expires:  exp,
	})
}
