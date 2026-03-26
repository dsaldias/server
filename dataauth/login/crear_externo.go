package login

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/dsaldias/server/graph_auth/model"

	"github.com/dsaldias/server/dataauth/usuarios"
)

type loginData struct {
	AccessToken string `json:"accessToken"`
}

type data struct {
	Login loginData `json:"login"`
}

type jSONData struct {
	Data data `json:"data"`
}

type Response struct {
	Data struct {
		Me struct {
			DatosPersonales struct {
				ID          int     `json:"id"`
				Nombres     string  `json:"nombres"`
				PriApellido string  `json:"pri_apellido"`
				SegApellido *string `json:"seg_apellido"`
				Usuario     struct {
					ID int `json:"id"`
				} `json:"usuario"`
			} `json:"datos_personales"`
		} `json:"me"`
	} `json:"data"`
}

func loginPortal(username string, password string) (string, error) {
	url := os.Getenv("EXTERNAL_AUTH")
	method := "POST"

	payload := "{\"query\":\"mutation {login(login: \\\"%s\\\", clave: \\\"%s\\\"){accessToken}}\\n\",\"variables\":{}}"
	payload = fmt.Sprintf(payload, username, password)

	data := strings.NewReader(payload)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	jsonStr := string(body)
	var jsonData jSONData
	err = json.Unmarshal([]byte(jsonStr), &jsonData)
	if err != nil {
		return "", err
	}
	return jsonData.Data.Login.AccessToken, nil
}

func getMe(token string) (*Response, error) {
	jsonStr, err := getDataMe(token)
	if err != nil {
		return nil, err
	}

	var response Response
	err = json.Unmarshal([]byte(jsonStr), &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func getDataMe(token string) (string, error) {
	url := os.Getenv("EXTERNAL_ME")
	method := "POST"
	payload := "{\"query\": \"{me{datos_personales{id,nombres,pri_apellido,seg_apellido,usuario{id}} }}\",\"variables\": {}}"
	data := strings.NewReader(payload)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	jsonStr := string(body)

	return jsonStr, nil
}

func CrearExterno(db *sql.DB, u, p string) (*model.Usuario, error) {
	token, err := loginPortal(u, p)
	if err != nil {
		return nil, err
	}
	if len(token) == 0 {
		return nil, errors.New("credenciales incorrectos")
	}

	data, err := getMe(token)
	if err != nil {
		return nil, err
	}

	existe_id, err := usuarios.GetIdByUsernamePortal(db, u)
	if err != nil {
		return nil, err
	}
	if len(existe_id) > 0 {
		return usuarios.UpdatePassword(db, existe_id, p)
	}

	dp := data.Data.Me.DatosPersonales
	sa := ""
	if dp.SegApellido != nil {
		sa = *dp.SegApellido
	}
	n := fmt.Sprintf("%s %s %s", dp.Nombres, dp.PriApellido, sa)

	input := model.NewUsuarioOauth{}
	input.Username = u
	input.Password = p
	input.Nombres = n
	return usuarios.CrearOauth(db, input, true)
}
