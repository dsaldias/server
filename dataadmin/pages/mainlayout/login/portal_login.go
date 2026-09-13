package xlogin

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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

func LoginPortal(username string, password string) (string, error) {
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
