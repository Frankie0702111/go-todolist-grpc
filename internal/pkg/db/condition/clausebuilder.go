package condition

import (
	"fmt"
	"go-todolist-grpc/internal/pkg/log"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func BuildExpression(name string, value any) []clause.Expression {
	var exps = make([]clause.Expression, 0)

	switch v := value.(type) {
	case *Int:
		exps = parseIntClause(name, *v)
	case *Float32:
		exps = parseFloat32Clause(name, *v)
	case *Time:
		exps = parseTimeClause(name, *v)
	case *String:
		exps = parseStringClause(name, *v)
	case *Bool:
		exps = parseBoolClause(name, *v)
	case *JSON:
		exps = parseJSONClause(name, *v)
	default:
		log.Warning.Printf("unknown where condition type %T", v)
	}

	return exps
}

func parseIntClause(name string, value Int) []clause.Expression {
	var exps = make([]clause.Expression, 0)

	if value.EQ != nil {
		exps = append(exps, clause.Eq{Column: name, Value: *value.EQ})
	}

	if value.GT != nil {
		exps = append(exps, clause.Gt{Column: name, Value: *value.GT})
	}

	if value.LT != nil {
		exps = append(exps, clause.Lt{Column: name, Value: *value.LT})
	}

	if value.GTE != nil {
		exps = append(exps, clause.Gte{Column: name, Value: *value.GTE})
	}

	if value.LTE != nil {
		exps = append(exps, clause.Lte{Column: name, Value: *value.LTE})
	}

	if value.IN != nil && len(value.IN) > 0 {
		var values = make([]interface{}, 0)
		for _, val := range value.IN {
			values = append(values, val)
		}
		exps = append(exps, clause.IN{Column: name, Values: values})
	}

	if value.CONTAIN != nil && len(value.CONTAIN) > 0 {
		var keys = make([]string, 0)
		var values = make([]interface{}, 0)
		for _, val := range value.CONTAIN {
			keys = append(keys, "?")
			values = append(values, val)
		}

		exps = append(exps, clause.Expr{SQL: name + ` @> ARRAY [` + strings.Join(keys, ", ") + `]::INT[]`, Vars: values})
	}

	if value.IsNull != nil {
		var exp clause.Expression
		exp = clause.Eq{Column: name, Value: nil}

		if !*value.IsNull {
			exp = clause.Not(exp)
		}

		exps = append(exps, exp)
	}

	return exps
}

func parseFloat32Clause(name string, value Float32) []clause.Expression {
	var exps = make([]clause.Expression, 0)

	if value.EQ != nil {
		exps = append(exps, clause.Eq{Column: name, Value: *value.EQ})
	}

	if value.GT != nil {
		exps = append(exps, clause.Gt{Column: name, Value: *value.GT})
	}

	if value.LT != nil {
		exps = append(exps, clause.Lt{Column: name, Value: *value.LT})
	}

	if value.GTE != nil {
		exps = append(exps, clause.Gte{Column: name, Value: *value.GTE})
	}

	if value.LTE != nil {
		exps = append(exps, clause.Lte{Column: name, Value: *value.LTE})
	}

	if value.IsNull != nil {
		var exp clause.Expression

		exp = clause.Eq{Column: name, Value: nil}

		if !*value.IsNull {
			exp = clause.Not(exp)
		}

		exps = append(exps, exp)
	}

	return exps
}

func parseTimeClause(name string, value Time) []clause.Expression {
	var exps = make([]clause.Expression, 0)

	if value.EQ != nil {
		exps = append(exps, clause.Eq{Column: name, Value: *value.EQ})
	}

	if value.GT != nil {
		exps = append(exps, clause.Gt{Column: name, Value: *value.GT})
	}

	if value.LT != nil {
		exps = append(exps, clause.Lt{Column: name, Value: *value.LT})
	}

	if value.GTE != nil {
		exps = append(exps, clause.Gte{Column: name, Value: *value.GTE})
	}

	if value.LTE != nil {
		exps = append(exps, clause.Lte{Column: name, Value: *value.LTE})
	}

	if value.IsNull != nil {
		var exp clause.Expression = clause.Eq{Column: name, Value: nil}

		if !*value.IsNull {
			exp = clause.Not(exp)
		}

		exps = append(exps, exp)
	}

	return exps
}

func parseStringClause(name string, value String) []clause.Expression {
	var exps = make([]clause.Expression, 0)

	if value.EQ != nil {
		exps = append(exps, clause.Eq{Column: name, Value: *value.EQ})
	}

	if value.NEQ != nil {
		exps = append(exps, clause.Not(clause.Eq{Column: name, Value: *value.NEQ}))
	}

	if value.Like != nil {
		exps = append(exps, clause.Like{Column: name, Value: "%" + escapeLikeCharacters(*value.Like) + "%"})
	}

	if value.NotLike != nil {
		exps = append(exps, clause.Not(clause.Like{Column: name, Value: "%" + escapeLikeCharacters(*value.NotLike) + "%"}))
	}

	if value.StartAt != nil {
		exps = append(exps, clause.Like{Column: name, Value: escapeLikeCharacters(*value.StartAt) + "%"})
	}

	if value.EndAt != nil {
		exps = append(exps, clause.Like{Column: name, Value: "%" + escapeLikeCharacters(*value.EndAt)})
	}

	if value.IN != nil && len(value.IN) > 0 {
		var values []interface{}
		for _, val := range value.IN {
			values = append(values, val)
		}
		exps = append(exps, clause.IN{Column: name, Values: values})
	}

	if len(value.OrderedIntersection) > 0 {
		exps = append(exps, clause.Like{Column: name, Value: "%" + strings.Join(value.OrderedIntersection, "%") + "%"})
	}

	if value.IsNull != nil {
		var exp clause.Expression = clause.Eq{Column: name, Value: nil}

		if *value.IsNull == false {
			exp = clause.Not(exp)
		}

		exps = append(exps, exp)
	}

	return exps
}

func escapeLikeCharacters(word string) string {
	var n int
	for i := range word {
		if c := word[i]; c == '%' || c == '_' || c == '\\' {
			n++
		}
	}
	// No characters to escape.
	if n == 0 {
		return word
	}
	var b strings.Builder
	b.Grow(len(word) + n)
	for _, c := range word {
		if c == '%' || c == '_' || c == '\\' {
			b.WriteByte('\\')
		}
		b.WriteRune(c)
	}
	return b.String()
}

func parseBoolClause(name string, value Bool) []clause.Expression {
	var exps = make([]clause.Expression, 0)

	if value.EQ != nil {
		exps = append(exps, gorm.Expr(fmt.Sprintf("%s IS %v", name, *value.EQ)))
	}

	if value.IsNull != nil {
		var exp clause.Expression

		exp = clause.Eq{Column: name, Value: nil}

		if !*value.IsNull {
			exp = clause.Not(exp)
		}

		exps = append(exps, exp)
	}

	if value.Not != nil {
		exps = append(exps, gorm.Expr(fmt.Sprintf("%s IS NOT %v", name, *value.Not)))
	}

	return exps
}

func parseJSONClause(name string, value JSON) []clause.Expression {
	var exps = make([]clause.Expression, 0)

	if value.EQ != nil {
		exps = append(exps, clause.Expr{SQL: name + " = ?", Vars: []interface{}{*value.EQ}})
	}

	if value.IN != nil && len(value.IN) > 0 {
		params := []string{}
		var values []interface{}
		for _, val := range value.IN {
			values = append(values, val)
			params = append(params, "?")
		}

		prepares := "(" + strings.Join(params, ",") + ")"

		exps = append(exps, clause.Expr{SQL: name + " in " + prepares, Vars: values})
	}

	return exps
}
