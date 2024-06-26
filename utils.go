package repository_gorm

import (
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
)

func getIDField(entity any) string {
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		panic("entity must be a struct")
	}

	field, found := t.FieldByName("ID")
	if !found {
		field, found = t.FieldByName("Id")
		if !found {
			panic("entity must have field `ID` or `Id`")
		}
	}
	gormTag := field.Tag.Get("gorm")
	if gormTag == "" {
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			return strings.Split(jsonTag, ",")[0]
		}
	}
	tagSetting := schema.ParseTagSetting(gormTag, ";")
	if fieldName, ok := tagSetting["COLUMN"]; ok {
		return fieldName
	}

	return "id"
}

func getDeletedAtField(entity any) string {
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		panic("entity must be a struct")
	}

	field, found := t.FieldByName("DeletedAt")
	if !found {
		return ""
	}
	gormTag := field.Tag.Get("gorm")
	if gormTag == "" {
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			return strings.Split(jsonTag, ",")[0]
		}
	}
	tagSetting := schema.ParseTagSetting(gormTag, ";")
	if fieldName, ok := tagSetting["COLUMN"]; ok {
		return fieldName
	}

	return "deleted_at"
}
