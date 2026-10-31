package utils

import (
	"slices"
	"testing"
)

func TestGetJSONKeys(t *testing.T) {
	var jsonStr = `
	{
		"Name": "test",
		"TableName": "test",
		"TemplateID": "test",
		"TemplateInfo": "test",
		"Limit": 0
}`
	keys, err := GetJSONKeys(jsonStr)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Name", "TableName", "TemplateID", "TemplateInfo", "Limit"}
	if !slices.Equal(keys, want) {
		t.Fatalf("GetJSONKeys() = %v, want %v", keys, want)
	}
}
