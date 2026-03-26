package xnotificaciones

import "encoding/json"

type DataNotify struct {
	Color *string `json:"color"`
	Tipo  *string `json:"tipo"`
	Datos any     `json:"datos"`
}

func formatToJson(data DataNotify) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	jsonString := string(jsonBytes)
	return jsonString, nil
}
