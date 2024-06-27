package builder

import "gorm.io/gorm/clause"

type TableJoins struct {
	FromTb    TableAlias
	LeftJoins []LeftJoin
}

type TableAlias struct {
	SubQuery *string
	Schema   *string
	Tb       string
	Alias    string
}

type LeftJoin struct {
	JoinTb           TableAlias
	OnFromTableAlias *string
	OnFromCol        string
	OnToCol          string
	ExtraStatements  []string
}

type OrderBy struct {
	Desc bool
}

func BuildOrderByExpression(name string, val any) clause.OrderByColumn {
	return clause.OrderByColumn{Column: clause.Column{Name: name}, Desc: val.(*OrderBy).Desc}
}
