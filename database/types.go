package database

type Pagination[T any] struct {
	Data []*T  `json:"data"`
	Meta *Meta `json:"meta"`
}

type Meta struct {
	ItemsPerPage uint64 `json:"items_per_page"`
	TotalItems   uint64 `json:"total_items"`
	CurrentPage  uint64 `json:"current_page"`
	TotalPages   uint64 `json:"total_pages"`
}

type Paging struct {
	Page  uint64 `json:"page"`
	Limit uint64 `json:"limit"`
}

type Sorting struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

type Condition struct {
	Field string
	Value interface{}
	Op    string // "eq", "ne", "lt", "gt", "lte", "gte", "in", "like", ...
}

type CommonCondition struct {
	Conditions []Condition
	Sorting    []Sorting
	Paging     *Paging
}

func NewCommonCondition() *CommonCondition {
	return &CommonCondition{
		Conditions: []Condition{},
		Sorting:    []Sorting{},
		Paging:     &Paging{},
	}
}

func (cc *CommonCondition) AddCondition(field string, value interface{}, op string) {
	cc.Conditions = append(cc.Conditions, Condition{
		Field: field,
		Value: value,
		Op:    op,
	})
}

func (cc *CommonCondition) SetPaging(limit, page uint64) {
	cc.Paging.Limit = limit
	cc.Paging.Page = page
}

func (cc *CommonCondition) AddSorting(field, order string) {
	cc.Sorting = append(cc.Sorting, Sorting{
		Field: field,
		Order: order,
	})
}

func (cc *CommonCondition) WithPaging(limit, page uint64) *CommonCondition {
	cc.Paging = &Paging{
		Limit: limit,
		Page:  page,
	}
	return cc
}

func (cc *CommonCondition) WithCondition(field string, value interface{}, op string) *CommonCondition {
	condition := Condition{
		Field: field,
		Value: value,
		Op:    op,
	}
	cc.Conditions = append(cc.Conditions, condition)
	return cc
}

func (cc *CommonCondition) WithSorting(field string, order string) *CommonCondition {
	sorting := Sorting{
		Field: field,
		Order: order,
	}
	cc.Sorting = append(cc.Sorting, sorting)
	return cc
}
