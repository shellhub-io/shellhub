package order

const (
	Asc  = "asc"
	Desc = "desc"
)

type Query struct {
	OrderBy string `query:"order_by" validate:"omitempty,oneof=asc desc"`
}

func (q *Query) Normalize() {
	if q.OrderBy == "" || (q.OrderBy != Asc && q.OrderBy != Desc) {
		q.OrderBy = Desc
	}
}
