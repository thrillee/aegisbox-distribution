package sp

import (
	"encoding/json"
	"fmt"
)

var (
	SP_HTTP_CALLBACK_URL = "HTTP_CALLBACK_URL"
	SP_HTTP_CALLBACK_KEY = "SP_HTTP_CALLBACK_KEY"
	SP_API_KEY           = "API_KEY"
)

func GetFromHttpConfig(httpConfigBytes []byte, fieldName string) (any, error) {
	var jsonData map[string]any
	err := json.Unmarshal(httpConfigBytes, &jsonData)
	if err != nil {
		return nil, fmt.Errorf("Invalid JSON data: size[%d]", len(httpConfigBytes))
	}

	fieldValue, ok := jsonData[fieldName]
	if !ok {
		return nil, fmt.Errorf("Field [%v] not found", fieldName)
	}
	return fieldValue, nil
}

func GetHttpCallBackURL(httpConfigBytes []byte) *string {
	value, err := GetFromHttpConfig(httpConfigBytes, SP_HTTP_CALLBACK_URL)
	if err != nil {
		return nil
	}
	result := value.(string)
	return &result
}

func GetHttpAPIKey(httpConfigBytes []byte) *string {
	value, err := GetFromHttpConfig(httpConfigBytes, SP_API_KEY)
	if err != nil {
		return nil
	}
	result := value.(string)
	return &result
}

func GetHttpCallBackKey(httpConfigBytes []byte) *string {
	value, err := GetFromHttpConfig(httpConfigBytes, SP_HTTP_CALLBACK_KEY)
	if err != nil {
		return nil
	}
	result := value.(string)
	return &result
}
