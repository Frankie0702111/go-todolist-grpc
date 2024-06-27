package condition

import "time"

type Int struct {
	EQ      *int
	GT      *int
	GTE     *int
	LT      *int
	LTE     *int
	IN      []int
	CONTAIN []int
	IsNull  *bool
}

type Float32 struct {
	EQ     *float32
	GT     *float32
	GTE    *float32
	LT     *float32
	LTE    *float32
	IsNull *bool
}

type Time struct {
	EQ     *time.Time
	GT     *time.Time
	GTE    *time.Time
	LT     *time.Time
	LTE    *time.Time
	IsNull *bool
}

type String struct {
	EQ                  *string
	NEQ                 *string
	Like                *string
	NotLike             *string
	StartAt             *string
	EndAt               *string
	IN                  []string
	OrderedIntersection []string
	IsNull              *bool
}

type Bool struct {
	EQ     *bool
	IsNull *bool
	Not    *bool
}

type JSON struct {
	EQ *string
	IN []string
}
