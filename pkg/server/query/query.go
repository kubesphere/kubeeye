package query

import (
	"github.com/gin-gonic/gin"
	"github.com/kubesphere/kubeeye/pkg/utils"
	"k8s.io/utils/strings/slices"
	"net/url"
	"sort"
	"strconv"
)

const (
	OrderBy       = "sortBy"
	Ascending     = "ascending"
	Descending    = "descending"
	Limit         = "limit"
	Page          = "page"
	LabelSelector = "labelSelector"
)
const (
	CreateTime        = "createTime"
	Name              = "name"
	InspectPolicy     = "inspectPolicy"
	Phase             = "phase"
	Duration          = "duration"
	Suspend           = "suspend"
	InspectType       = "inspectType"
	LastTaskStatus    = "lastTaskStatus"
	LastTaskStartTime = "lastTaskStartTime"
)

type SortBy string
type Filter map[string]string
type Query struct {
	// pagination
	Pagination *Pagination
	// sort by
	SortBy SortBy
	// sort result in ascending or descending order, default to descending
	Ascending bool
	// filters
	Filters *Filter
	// label selector
	LabelSelector string
}

type Result struct {
	// total number of items
	TotalItems int `json:"totalItems,omitempty"`
	// items
	Items interface{} `json:"items"`
}

type Pagination struct {
	// items per page
	Limit int
	// offset
	Offset int
}

func ParseQuery(g *gin.Context) *Query {
	q := NewQuery()
	q.Pagination = ParsePagination(g.Request.URL.Query())
	q.Filters = q.ParseFilter(g.Request.URL.Query())
	q.SortBy = SortBy(g.Request.URL.Query().Get(OrderBy))
	q.Ascending = q.ParseAscending(g.Request.URL.Query().Get(Ascending))
	q.LabelSelector = g.Request.URL.Query().Get(LabelSelector)
	return q
}

func NewQuery() *Query {
	return &Query{
		Pagination: &Pagination{
			Limit:  10,
			Offset: 0,
		},
		SortBy:        "",
		Ascending:     false,
		Filters:       nil,
		LabelSelector: "",
	}
}

func (q *Query) ParseAscending(b string) bool {
	if b == "true" {
		return true
	}
	return false
}

func (q *Query) ParseFilter(values url.Values) *Filter {
	var filters *Filter
	continues := []string{Limit, Page, OrderBy, Ascending, LabelSelector}
	for key, value := range values {
		if !slices.Contains(continues, key) {
			if filters == nil {
				filters = &Filter{}
			}
			(*filters)[key] = value[0]
		}

	}
	return filters
}

func NewPagination() *Pagination {
	return &Pagination{
		Limit:  10,
		Offset: 0,
	}
}

func ParsePagination(values url.Values) *Pagination {
	limit := values.Get(Limit)
	page := values.Get(Page)
	pagination := NewPagination()
	if page != "" {
		atoi, err := strconv.Atoi(page)
		if err != nil || atoi < 1 {
			atoi = 1
		}
		pagination.Offset = (atoi - 1) * pagination.Limit
	}
	if limit != "" {
		atoi, err := strconv.Atoi(limit)
		if err != nil || atoi < 1 {
			atoi = 10
		}
		pagination.Limit = atoi
	}
	return pagination
}

func (q *Query) GetPageData(data interface{}, c compare, f filterC) Result {
	toMap, err := utils.StructToMap(data)
	if err != nil {
		return Result{
			TotalItems: 0,
			Items:      nil,
		}
	}

	if q.Filters != nil && f != nil {

		toMap = q.Filters.F(toMap, f)
	}

	if q.SortBy != "" && c != nil {
		if q.Ascending {
			q.SortBy.Asc(toMap, c)
		} else {
			q.SortBy.Desc(toMap, c)
		}
	}

	start, end := q.Pagination.computeIndex(len(toMap))

	return Result{
		TotalItems: len(toMap),
		Items:      toMap[start:end],
	}
}

func (p *Pagination) computeIndex(total int) (int, int) {
	var startIndex, endIndex = 0, 0

	if p.Limit < 0 || p.Offset < 0 || p.Offset > total {
		return 0, 0
	}

	startIndex = p.Offset
	endIndex = startIndex + p.Limit

	if endIndex > total {
		endIndex = total
	}
	if startIndex > total {
		startIndex = 0
	}

	return startIndex, endIndex
}

type compare func(i, j map[string]interface{}, orderBy string) bool

func (s *SortBy) Desc(data []map[string]interface{}, c compare) {
	sort.Slice(data, func(i, j int) bool {
		return !c(data[i], data[j], string(*s))
	})
}

func (s *SortBy) Asc(data []map[string]interface{}, c compare) {
	sort.Slice(data, func(i, j int) bool {
		return c(data[i], data[j], string(*s))
	})
}

type filterC func(data map[string]interface{}, f *Filter) bool

func (f *Filter) F(d []map[string]interface{}, fc filterC) []map[string]interface{} {
	var data []map[string]interface{}
	for _, m := range d {
		if fc(m, f) {
			data = append(data, m)
		}
	}
	return data
}

func (f *Filter) Get(key string) string {
	if key == "" || f == nil {
		return ""
	}

	s, ok := (*f)[key]
	if !ok {
		return ""
	}
	return s
}

func (f *Filter) Keys() (keys []string) {
	if f == nil {
		return keys
	}
	for k := range *f {
		keys = append(keys, k)
	}
	return keys
}
