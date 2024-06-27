package model

import (
	"database/sql"
	"go-todolist-grpc/internal/pkg/db/builder"
	"go-todolist-grpc/internal/pkg/db/condition"
	"go-todolist-grpc/internal/pkg/db/field"
	"go-todolist-grpc/internal/pkg/util"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type DBExecutable interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	gorm.ConnPool
}

// GiveColString wraps string
func GiveColString(v string) field.String {
	return field.String{
		Val:   v,
		Given: true,
	}
}

// GiveColStringArray wraps string array
func GiveColStringArray(v []string) field.StringArray {
	return field.StringArray{
		Val:   v,
		Given: true,
	}
}

// GiveColNullString wraps null string
func GiveColNullString(v *string) field.NullString {
	if v != nil {
		return field.NullString{
			Val:   *v,
			Given: true,
		}
	}
	return field.NullString{
		Given:  true,
		IsNull: true,
	}
}

// GiveColInt wraps int
func GiveColInt(v int) field.Int {
	return field.Int{
		Val:   v,
		Given: true,
	}
}

// GiveColInt64 wraps int64
func GiveColInt64(v int64) field.Int64 {
	return field.Int64{
		Val:   v,
		Given: true,
	}
}

// GiveColInt32Array wraps int32 array
func GiveColInt32Array(v pq.Int32Array) field.Int32Array {
	return field.Int32Array{
		Val:   v,
		Given: true,
	}
}

// GiveColInt64Array wraps int32 array
func GiveColInt64Array(v pq.Int64Array) field.Int64Array {
	return field.Int64Array{
		Val:   v,
		Given: true,
	}
}

// GiveColNullInt wraps null int
func GiveColNullInt(v *int) field.NullInt {
	if v != nil {
		return field.NullInt{
			Val:   *v,
			Given: true,
		}
	}
	return field.NullInt{
		Given:  true,
		IsNull: true,
	}
}

// GiveColFloat32 wraps float32
func GiveColFloat32(v float32) field.Float32 {
	return field.Float32{
		Val:   v,
		Given: true,
	}
}

// GiveColFloat64 wraps float64
func GiveColFloat64(v float64) field.Float64 {
	return field.Float64{
		Val:   v,
		Given: true,
	}
}

// GiveColNullFloat32 wraps null float32
func GiveColNullFloat32(v *float32) field.NullFloat32 {
	if v != nil {
		return field.NullFloat32{
			Val:   *v,
			Given: true,
		}
	}
	return field.NullFloat32{
		Given:  true,
		IsNull: true,
	}
}

// GiveColNullFloat64 wraps null float64
func GiveColNullFloat64(v *float64) field.NullFloat64 {
	if v != nil {
		return field.NullFloat64{
			Val:   *v,
			Given: true,
		}
	}
	return field.NullFloat64{
		Given:  true,
		IsNull: true,
	}
}

// GiveColBool wraps bool
func GiveColBool(v bool) field.Bool {
	return field.Bool{
		Val:   v,
		Given: true,
	}
}

// GiveColNullBool wraps null bool
func GiveColNullBool(v *bool) field.NullBool {
	if v != nil {
		return field.NullBool{
			Val:   *v,
			Given: true,
		}
	}
	return field.NullBool{
		Given:  true,
		IsNull: true,
	}
}

// GiveColTime wraps time.Time
func GiveColTime(v time.Time) field.Time {
	return field.Time{
		Val:   v,
		Given: true,
	}
}

// GiveColNullTime wraps null time.Time
func GiveColNullTime(v *time.Time) field.NullTime {
	if v != nil {
		return field.NullTime{
			Val:   *v,
			Given: true,
		}
	}
	return field.NullTime{
		Given:  true,
		IsNull: true,
	}
}

func ParseConditionParams(paramsGetter func(string) map[string]string, holder interface{}) {
	e := reflect.ValueOf(holder).Elem()
	for i := 0; i < e.NumField(); i++ {
		paramKey, _ := e.Type().Field(i).Tag.Lookup("db_col")

		if dbAlias, ok := e.Type().Field(i).Tag.Lookup("db_alias"); ok {
			paramKey = dbAlias + "." + paramKey
		}

		// json tag override all tags
		if json, ok := e.Type().Field(i).Tag.Lookup("json"); ok {
			paramKey = json
		}

		param := paramsGetter(paramKey)
		valueField := e.Field(i)
		iValue := valueField.Interface()
		switch iValue.(type) {
		case *condition.Time:
			conTime := &condition.Time{}
			isValidCon := false
			if v, valOK := param["gt"]; valOK {
				if t := util.MsTimestampStrToTime(v); t != nil {
					isValidCon = true
					conTime.GT = t
				}
			}
			if v, valOK := param["gte"]; valOK {
				if t := util.MsTimestampStrToTime(v); t != nil {
					isValidCon = true
					conTime.GTE = t
				}
			}
			if v, valOK := param["lt"]; valOK {
				if t := util.MsTimestampStrToTime(v); t != nil {
					isValidCon = true
					conTime.LT = t
				}
			}
			if v, valOK := param["lte"]; valOK {
				if t := util.MsTimestampStrToTime(v); t != nil {
					isValidCon = true
					conTime.LTE = t
				}
			}
			if v, valOK := param["eq"]; valOK {
				if t := util.MsTimestampStrToTime(v); t != nil {
					isValidCon = true
					conTime.EQ = t
				}
			}
			if v, valOK := param["null"]; valOK {
				t := true
				f := false
				switch v {
				case "true":
					isValidCon = true
					conTime.IsNull = &t
				case "false":
					isValidCon = true
					conTime.IsNull = &f
				}
			}
			if isValidCon {
				valueField.Set(reflect.ValueOf(conTime))
			}
		case *condition.Int:
			conInt := &condition.Int{}
			isValidCon := false
			if v, valOK := param["gt"]; valOK {
				if intV, intErr := strconv.Atoi(v); intErr == nil {
					isValidCon = true
					conInt.GT = &intV
				}
			}
			if v, valOK := param["gte"]; valOK {
				if intV, intErr := strconv.Atoi(v); intErr == nil {
					isValidCon = true
					conInt.GTE = &intV
				}
			}
			if v, valOK := param["lt"]; valOK {
				if intV, intErr := strconv.Atoi(v); intErr == nil {
					isValidCon = true
					conInt.LT = &intV
				}
			}
			if v, valOK := param["lte"]; valOK {
				if intV, intErr := strconv.Atoi(v); intErr == nil {
					isValidCon = true
					conInt.LTE = &intV
				}
			}
			if v, valOK := param["eq"]; valOK {
				if intV, intErr := strconv.Atoi(v); intErr == nil {
					isValidCon = true
					conInt.EQ = &intV
				}
			}
			if v, valOK := param["null"]; valOK {
				t := true
				f := false
				switch v {
				case "true":
					isValidCon = true
					conInt.IsNull = &t
				case "false":
					isValidCon = true
					conInt.IsNull = &f
				}
			}
			if v, valOK := param["in"]; valOK {
				if len(v) > 0 {
					list := []int{}
					for _, v := range strings.Split(v, ",") {
						if intV, intErr := strconv.Atoi(v); intErr == nil {
							if !isValidCon {
								isValidCon = true
							}
							list = append(list, intV)
						}
					}
					conInt.IN = list
				}
			}
			if v, valOK := param["contain"]; valOK {
				if len(v) > 0 {
					list := []int{}
					for _, v := range strings.Split(v, ",") {
						if intV, intErr := strconv.Atoi(v); intErr == nil {
							if !isValidCon {
								isValidCon = true
							}
							list = append(list, intV)
						}
					}
					conInt.CONTAIN = list
				}
			}
			if isValidCon {
				valueField.Set(reflect.ValueOf(conInt))
			}
		case *condition.Float32:
			conFloat32 := &condition.Float32{}
			isValidCon := false
			if v, valOK := param["gt"]; valOK {
				if floatV, floatErr := strconv.ParseFloat(v, 32); floatErr == floatErr {
					isValidCon = true
					float32V := float32(floatV)
					conFloat32.GT = &float32V
				}
			}
			if v, valOK := param["gte"]; valOK {
				if floatV, floatErr := strconv.ParseFloat(v, 32); floatErr == floatErr {
					isValidCon = true
					float32V := float32(floatV)
					conFloat32.GTE = &float32V
				}
			}
			if v, valOK := param["lt"]; valOK {
				if floatV, floatErr := strconv.ParseFloat(v, 32); floatErr == floatErr {
					isValidCon = true
					float32V := float32(floatV)
					conFloat32.LT = &float32V
				}
			}
			if v, valOK := param["lte"]; valOK {
				if floatV, floatErr := strconv.ParseFloat(v, 32); floatErr == floatErr {
					isValidCon = true
					float32V := float32(floatV)
					conFloat32.LTE = &float32V
				}
			}
			if v, valOK := param["eq"]; valOK {
				if floatV, floatErr := strconv.ParseFloat(v, 32); floatErr == floatErr {
					isValidCon = true
					float32V := float32(floatV)
					conFloat32.EQ = &float32V
				}
			}
			if v, valOK := param["null"]; valOK {
				t := true
				f := false
				switch v {
				case "true":
					isValidCon = true
					conFloat32.IsNull = &t
				case "false":
					isValidCon = true
					conFloat32.IsNull = &f
				}
			}
			if isValidCon {
				valueField.Set(reflect.ValueOf(conFloat32))
			}
		case *condition.JSON:
			conJSON := &condition.JSON{}
			isValidCon := false
			if v, valOK := param["eq"]; valOK {
				isValidCon = true
				conJSON.EQ = &v
			}
			if v, valOK := param["in"]; valOK {
				if len(v) > 0 {
					isValidCon = true
					conJSON.IN = strings.Split(v, ",")
				}
			}
			if isValidCon {
				valueField.Set(reflect.ValueOf(conJSON))
			}
		case *condition.String:
			conString := &condition.String{}
			isValidCon := false
			if v, valOK := param["eq"]; valOK {
				isValidCon = true
				conString.EQ = &v
			}
			if v, valOK := param["neq"]; valOK {
				isValidCon = true
				conString.NEQ = &v
			}
			if v, valOK := param["like"]; valOK {
				isValidCon = true
				conString.Like = &v
			}
			if v, valOK := param["nlike"]; valOK {
				isValidCon = true
				conString.NotLike = &v
			}
			if v, valOK := param["start_at"]; valOK {
				isValidCon = true
				conString.StartAt = &v
			}
			if v, valOK := param["end_at"]; valOK {
				isValidCon = true
				conString.EndAt = &v
			}
			if v, valOK := param["null"]; valOK {
				t := true
				f := false
				switch v {
				case "true":
					isValidCon = true
					conString.IsNull = &t
				case "false":
					isValidCon = true
					conString.IsNull = &f
				}
			}
			if v, valOK := param["in"]; valOK {
				if len(v) > 0 {
					isValidCon = true
					conString.IN = strings.Split(v, ",")
				}
			}
			if v, valOK := param["ordered_intersection"]; valOK {
				if len(v) > 0 {
					isValidCon = true
					conString.OrderedIntersection = strings.Split(v, ",")
					sort.SliceStable(conString.OrderedIntersection, func(i, j int) bool {
						return conString.OrderedIntersection[i] < conString.OrderedIntersection[j]
					})
				}
			}
			if isValidCon {
				valueField.Set(reflect.ValueOf(conString))
			}
		case *condition.Bool:
			conBool := &condition.Bool{}
			isValidCon := false
			if v, valOK := param["eq"]; valOK {
				t := true
				f := false
				switch v {
				case "true":
					isValidCon = true
					conBool.EQ = &t
				case "false":
					isValidCon = true
					conBool.EQ = &f
				}
			}
			if v, valOK := param["not"]; valOK {
				t := true
				f := false
				switch v {
				case "true":
					isValidCon = true
					conBool.Not = &t
				case "false":
					isValidCon = true
					conBool.Not = &f
				}
			}
			if v, valOK := param["null"]; valOK {
				t := true
				f := false
				switch v {
				case "true":
					isValidCon = true
					conBool.IsNull = &t
				case "false":
					isValidCon = true
					conBool.IsNull = &f
				}
			}
			if isValidCon {
				valueField.Set(reflect.ValueOf(conBool))
			}
		}
	}
}

func ParseOrderByParams(params map[string]bool, holder interface{}) {
	e := reflect.ValueOf(holder).Elem()
	for i := 0; i < e.NumField(); i++ {
		paramKey, _ := e.Type().Field(i).Tag.Lookup("db_col")
		if dbAlias, ok := e.Type().Field(i).Tag.Lookup("db_alias"); ok {
			paramKey = dbAlias + "." + paramKey
		}
		if json, ok := e.Type().Field(i).Tag.Lookup("json"); ok {
			paramKey = json
		}
		if val, valOK := params[paramKey]; valOK {
			valueField := e.Field(i)
			ob := &builder.OrderBy{
				Desc: !val,
			}
			valueField.Set(reflect.ValueOf(ob))
		}
	}
}

func getGivenKeyValues(holder interface{}, useAlias bool) []field.KeyValue {
	var values []field.KeyValue
	if holder == nil {
		return values
	}
	e := reflect.ValueOf(holder).Elem()
	for i := 0; i < e.NumField(); i++ {
		col, _ := e.Type().Field(i).Tag.Lookup("db_col")
		valueField := e.Field(i)
		if !valueField.IsNil() {
			escapedCol := "\"" + col + "\""
			escapedAlias := ""
			if useAlias { // joined query 才需要
				// default alias
				if v, ok := holder.(schema.Tabler); ok {
					alias := v.TableName()
					escapedAlias = "\"" + alias + "\"."
				}
				// override by db_alias
				if alias, aliasOK := e.Type().Field(i).Tag.Lookup("db_alias"); aliasOK {
					escapedAlias = "\"" + alias + "\"."
				}
			} else {
				if v, ok := holder.(schema.Tabler); ok {
					alias := v.TableName()
					escapedAlias = "\"" + alias + "\"."
				}
			}
			key := escapedAlias + escapedCol
			if dbJsonOp, ok := e.Type().Field(i).Tag.Lookup("db_json_op"); ok {
				key = key + dbJsonOp
			}
			values = append(values, field.KeyValue{Key: key, Value: valueField.Interface()})
		}
	}
	return values
}

type Entity interface {
	Scan(*sql.Rows) error
}

type modelEntity interface {
	CreateEntity() Entity
}

type modelJoinedConditions interface {
	GetAliasColumnNames() [][]string
	GetTableJoins() builder.TableJoins
	modelEntity
}

type modelJoinedConditionsWithAggregate interface {
	GetAliasColumnNames() [][]string
	GetTableJoins() builder.TableJoins
	GetAggregateColumn() *string
	modelEntity
}

type modelTable interface {
	GetTableName() string
}

type Conditions interface {
	GetColumnNames() []string
	modelTable
	modelEntity
}

func BuildWhereClause(holder interface{}) clause.Where {
	var expressions []clause.Expression
	var withAlias bool

	switch holder.(type) {
	case modelJoinedConditions, modelJoinedConditionsWithAggregate:
		withAlias = true
	case Conditions:
		withAlias = false
	default:
		withAlias = true
	}

	kvs := getGivenKeyValues(holder, withAlias)
	for _, kv := range kvs {
		expression := condition.BuildExpression(kv.Key, kv.Value)
		expressions = append(expressions, expression...)
	}

	return clause.Where{Exprs: expressions}
}

func BuildOrderByClause(holder interface{}) clause.OrderBy {
	var columns []clause.OrderByColumn

	kvs := getGivenKeyValues(holder, true)

	for _, kv := range kvs {

		column := builder.BuildOrderByExpression(kv.Key, kv.Value)

		columns = append(columns, column)
	}

	return clause.OrderBy{Columns: columns}
}
