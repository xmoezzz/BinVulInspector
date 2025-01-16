package models

const (
	SortTypeAsc  = "asc"
	SortTypeDesc = "desc"

	Asc  int8 = 1
	Desc int8 = -1
)

func SortTypeValue(sort string) int8 {
	if sort == SortTypeAsc {
		return Asc
	}
	return Desc
}

func SortTypes() []string {
	return []string{SortTypeAsc, SortTypeDesc}
}

const (
	Disabled = 0
	Enabled  = 1
)

func Statuses() []int {
	return []int{Disabled, Enabled}
}
