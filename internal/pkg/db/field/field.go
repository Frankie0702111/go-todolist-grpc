package field

import (
	"context"
	"database/sql/driver"
	"errors"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// KeyValue
type KeyValue struct {
	Key   string
	Value interface{}
}

// String
type String struct {
	Val             string
	Given           bool
	CaseInsensitive bool
}

func (s String) Value() (driver.Value, error) {
	return s.Val, nil
}

func (s String) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return gorm.Expr("?", s.Val)
}

func (s String) GormDataType() string {
	return "string"
}

// NullString
type NullString struct {
	Val             string
	IsNull          bool
	IsNotNull       bool
	Given           bool
	CaseInsensitive bool
}

func (s NullString) Value() (driver.Value, error) {
	if !s.Given || s.IsNull {
		return nil, nil
	}

	return s.Val, nil
}

func (s NullString) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	if !s.Given || s.IsNull {
		return gorm.Expr("?", nil)
	}
	return gorm.Expr("?", s.Val)
}

func (s NullString) GormDataType() string {
	return "string"
}

// StringArray
type StringArray struct {
	Val   pq.StringArray
	Given bool
}

func (s StringArray) Value() (driver.Value, error) {
	return s.Val.Value()
}

func (s StringArray) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return gorm.Expr("?", s.Val)
}

func (s StringArray) GormDataType() string {
	return "text[]"
}

// Int
type Int struct {
	Val   int
	Given bool
}

func (i Int) Value() (driver.Value, error) {
	if !i.Given {
		return nil, nil
	}

	return i.Val, nil
}

func (i Int) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return gorm.Expr("?", i.Val)
}

func (i Int) GormDataType() string {
	return "int"
}

func (i *Int) Scan(src any) error {
	i.Given = true
	switch val := src.(type) {
	case int:
		i.Val = val
	case int8:
		i.Val = int(val)
	case int32:
		i.Val = int(val)
	case int64:
		i.Val = int(val)
	default:
		return errors.New("non-integer type")
	}

	return nil
}

// Int64
type Int64 struct {
	Val   int64
	Given bool
}

func (i Int64) Value() (driver.Value, error) {
	return i.Val, nil
}

func (i Int64) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return gorm.Expr("?", i.Val)
}

func (i Int64) GormDataType() string {
	return "int"
}

// Int32Array
type Int32Array struct {
	Val   pq.Int32Array
	Given bool
}

func (i Int32Array) Value() (driver.Value, error) {
	return i.Val.Value()
}

func (i Int32Array) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return gorm.Expr("?", i.Val)
}

func (i Int32Array) GormDataType() string {
	return "integer[]"
}

// Int64Array
type Int64Array struct {
	Val   pq.Int64Array
	Given bool
}

func (i Int64Array) Value() (driver.Value, error) {
	return i.Val.Value()
}

func (i Int64Array) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return gorm.Expr("?", i.Val)
}

func (i Int64Array) GormDataType() string {
	return "integer[]"
}

// NullInt
type NullInt struct {
	Val       int
	IsNull    bool
	IsNotNull bool
	Given     bool
}

func (i NullInt) Value() (driver.Value, error) {
	if !i.Given || i.IsNull {
		return nil, nil
	}

	return i.Val, nil
}

func (i NullInt) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	if !i.Given || i.IsNull {
		return gorm.Expr("?", nil)
	}

	return gorm.Expr("?", i.Val)
}

func (i NullInt) GormDataType() string {
	return "int"
}

// Float32
type Float32 struct {
	Val   float32
	Given bool
}

func (f Float32) Value() (driver.Value, error) {
	return f.Val, nil
}

func (f Float32) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return gorm.Expr("?", f.Val)
}

func (f Float32) GormDataType() string {
	return "float"
}

// Float64
type Float64 struct {
	Val   float64
	Given bool
}

func (f Float64) Value() (driver.Value, error) {
	return f.Val, nil
}

func (f Float64) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return gorm.Expr("?", f.Val)
}

func (f Float64) GormDataType() string {
	return "float"
}

// NullFloat32
type NullFloat32 struct {
	Val       float32
	IsNull    bool
	IsNotNull bool
	Given     bool
}

func (f NullFloat32) Value() (driver.Value, error) {
	if !f.Given || f.IsNull {
		return nil, nil
	}

	return f.Val, nil
}

func (f NullFloat32) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	if !f.Given || f.IsNull {
		return gorm.Expr("?", nil)
	}

	return gorm.Expr("?", f.Val)
}

func (f NullFloat32) GormDataType() string {
	return "float"
}

// NullFloat64
type NullFloat64 struct {
	Val       float64
	Given     bool
	IsNull    bool
	IsNotNull bool
}

func (f NullFloat64) Value() (driver.Value, error) {
	if !f.Given || f.IsNull {
		return nil, nil
	}

	return f.Val, nil
}

func (f NullFloat64) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	if !f.Given || f.IsNull {
		return gorm.Expr("?", nil)
	}

	return gorm.Expr("?", f.Val)
}

func (f NullFloat64) GormDataType() string {
	return "float"
}

// Time
type Time struct {
	Val   time.Time
	Given bool
}

func (t Time) Value() (driver.Value, error) {
	return t.Val, nil
}

func (t Time) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return gorm.Expr("?", t.Val)
}

func (t Time) GormDataType() string {
	return "time"
}

// NullTime
type NullTime struct {
	Val    time.Time
	IsNull bool
	Given  bool
}

func (t NullTime) Value() (driver.Value, error) {
	if !t.Given || t.IsNull {
		return nil, nil
	}

	return t.Val, nil
}

func (t NullTime) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	if !t.Given || t.IsNull {
		return gorm.Expr("?", nil)
	}

	return gorm.Expr("?", t.Val)
}

func (t NullTime) GormDataType() string {
	return "time"
}

// Bool
type Bool struct {
	Val   bool
	Given bool
}

func (b Bool) Value() (driver.Value, error) {
	return b.Val, nil
}

func (b Bool) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return gorm.Expr("?", b.Val)
}

func (b Bool) GormDataType() string {
	return "bool"
}

// NullBool
type NullBool struct {
	Val       bool
	IsNull    bool
	IsNotNull bool
	Given     bool
}

func (b NullBool) Value() (driver.Value, error) {
	if !b.Given || b.IsNull {
		return nil, nil
	}

	return b.Val, nil
}

func (b NullBool) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	if !b.Given || b.IsNull {
		return gorm.Expr("?", nil)
	}

	return gorm.Expr("?", b.Val)
}

func (b NullBool) GormDataType() string {
	return "boolean"
}

// Cus
type Cus struct {
	Val   interface{}
	Given bool
}

func (c Cus) Value() (driver.Value, error) {
	return c.Val, nil
}

func (c Cus) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return gorm.Expr("?", c.Val)
}

func (c Cus) GormDataType() string {
	switch c.Val.(type) {
	case string:
		return "string"
	case int, int8, int16, int32, int64:
		return "int"
	case float32, float64:
		return "float"
	case bool:
		return "bool"
	case time.Time:
		return "time"
	case byte:
		return "bytes"
	}

	return "string"
}
