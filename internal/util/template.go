package util

import (
	"bytes"
	"text/template"
)

func RenderTemplate(templateString string, data interface{}) (string, error) {
	tmpl, err := template.New("test").Parse(templateString)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, data)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}
