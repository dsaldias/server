package xlogin

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
)

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

			Cargos []struct {
				Cargo struct {
					ID     int    `json:"id"`
					Nombre string `json:"nombre"`
				}
				Sede struct {
					ID     int    `json:"id"`
					Nombre string `json:"nombre"`
				}
			}
		} `json:"me"`
	} `json:"data"`
}

func getDataMe(token string) (string, error) {
	url := os.Getenv("EXTERNAL_ME")
	method := "POST"
	payload := "{\"query\": \"{me{datos_personales{id,nombres,pri_apellido,seg_apellido,usuario{id} } cargos{cargo{id nombre} sede{id nombre unidad_negocio}} }}\",\"variables\": {}}"
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

func GetMe(token string) (*Response, error) {
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
