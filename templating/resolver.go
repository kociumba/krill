package templating

import (
	"fmt"
	"reflect"
	"strings"
)

func resolveTags(v any) (map[string]any, error) {
	result := make(map[string]any)
	val := reflect.ValueOf(v)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, got %s", val.Kind())
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		tomlTag := field.Tag.Get("toml")
		fieldName := field.Name
		fieldValue := val.Field(i)

		if tomlTag != "" {
			parts := strings.Split(tomlTag, ",")
			tomlName := parts[0]
			if tomlName != "" {
				fieldName = tomlName
			}
		}

		if fieldValue.Kind() == reflect.Struct {
			nestedMap, err := resolveTags(fieldValue.Interface())
			if err != nil {
				return nil, err
			}

			result[fieldName] = nestedMap
		} else if fieldValue.Kind() == reflect.Map {
			nestedMap := make(map[string]any)
			for _, key := range fieldValue.MapKeys() {
				mapValue := fieldValue.MapIndex(key).Interface()
				mapValReflect := reflect.ValueOf(mapValue)
				if mapValReflect.Kind() == reflect.Struct {
					nestedNested, err := resolveTags(mapValue)
					if err != nil {
						return nil, err
					}

					nestedMap[key.String()] = nestedNested
				} else if mapValReflect.Kind() == reflect.Map {
					nestedNested, err := resolveTags(mapValue)
					if err != nil {
						return nil, err
					}
					nestedMap[key.String()] = nestedNested
				} else {
					nestedMap[key.String()] = mapValue
				}
			}
			result[fieldName] = nestedMap
		} else {
			result[fieldName] = fieldValue.Interface()
		}
	}

	return result, nil
}
