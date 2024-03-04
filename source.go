package main

import (
	"encoding/json"

	_ "github.com/go-kratos/kratos/v2/encoding/json"
	_ "github.com/go-kratos/kratos/v2/encoding/yaml"
	"gopkg.in/yaml.v3"
)

func isJSON(data []byte) bool {
	return json.Valid(data)
}

func isYAML(data []byte) bool {
	return yaml.Unmarshal(data, &struct{}{}) == nil
}

func format(data []byte) string {
	if isYAML(data) {
		return "yaml"
	}
	// if isJSON(data) {
	// 	return "json"
	// }
	return "json"
}
