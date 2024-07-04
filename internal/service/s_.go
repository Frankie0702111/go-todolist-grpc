package service

import (
	"fmt"
	"go-todolist-grpc/api/pb"
	"reflect"
	"strings"

	"github.com/go-playground/validator"
)

type Server struct {
	pb.UnimplementedToDoListServer
}

func bindRequest(req interface{}, reqStruct interface{}) error {
	reqVal := reflect.ValueOf(req).Elem()
	val := reflect.ValueOf(reqStruct).Elem()

	// Create a reflection map to speed up field lookup
	reqFieldMap := make(map[string]reflect.Value, reqVal.NumField())
	for i := 0; i < reqVal.NumField(); i++ {
		field := reqVal.Type().Field(i)
		reqFieldMap[field.Name] = reqVal.Field(i)
	}

	// Iterate through reqStruct fields, retrieving corresponding req field values from the map
	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)
		fieldType := val.Type().Field(i)

		reqField, ok := reqFieldMap[fieldType.Name]
		if !ok {
			continue // Ignore fields that don't exist in req
		}

		// Handle pointer types
		if fieldVal.Kind() == reflect.Ptr {
			if reqField.IsNil() {
				continue // Ignore nil values
			}
			if fieldVal.IsNil() {
				// Initialize pointer
				fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
			}
			fieldVal = fieldVal.Elem()
			reqField = reqField.Elem()
		}

		// Assign values based on field type
		switch fieldVal.Kind() {
		case reflect.String:
			fieldVal.SetString(reqField.String())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldVal.SetInt(reqField.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fieldVal.SetUint(reqField.Uint())
		case reflect.Float32, reflect.Float64:
			fieldVal.SetFloat(reqField.Float())
		case reflect.Bool:
			fieldVal.SetBool(reqField.Bool())
		default:
			return fmt.Errorf("unsupported field type: %v", fieldVal.Kind())
		}
	}

	validate := validator.New()
	if err := validate.Struct(reqStruct); err != nil {
		return err
	}

	return nil
}

func ParseSortBy(str string) map[string]bool {
	rtn := map[string]bool{}
	if len(str) == 0 {
		return rtn
	}

	arr := strings.Split(str, ",")
	for _, v := range arr {
		if len(v) == 0 {
			continue
		}
		switch v[:1] {
		case "+":
			rtn[v[1:]] = true
		case "-":
			rtn[v[1:]] = false
		default:
			rtn[v] = true
		}
	}

	return rtn
}

type ReqId struct {
	Id int64 `json:"id" validate:"required,min=1"`
}
