// Package pagination implements the universal list-endpoint contract:
// every list takes page, page_size, q, sort_by and sort_order.
package pagination

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
	SortAsc         = "asc"
	SortDesc        = "desc"
)

// Params is the parsed, validated, safe-to-use form of the query string.
type Params struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	Search     string `form:"q"`
	SortColumn string
	SortOrder  string
}

func (p Params) Offset() int        { return (p.Page - 1) * p.PageSize }
func (p Params) Limit() int         { return p.PageSize }
func (p Params) OrderBy() string    { return p.SortColumn + " " + p.SortOrder }
func (p Params) SearchPattern() string { return "%" + p.Search + "%" }
func (p Params) HasSearch() bool    { return p.Search != "" }

// Sortable is a resource's whitelist: API field name -> database column.
type Sortable map[string]string

func (s Sortable) Fields() []string {
	out := make([]string, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Raw is the unvalidated query input.
type Raw struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	Q         string `form:"q"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}

// Resolve validates raw query input against a resource's whitelist.
func Resolve(raw Raw, sortable Sortable, defaultSort string) (Params, error) {
	col, ok := sortable[defaultSort]
	if !ok {
		panic(fmt.Sprintf("pagination: default sort %q is not in the sortable whitelist", defaultSort))
	}

	p := Params{
		Page:       raw.Page,
		PageSize:   raw.PageSize,
		Search:     strings.TrimSpace(raw.Q),
		SortColumn: col,
		SortOrder:  SortDesc,
	}

	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.PageSize < 1 {
		p.PageSize = DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	}

	if raw.SortBy != "" {
		col, ok := sortable[raw.SortBy]
		if !ok {
			return Params{}, apperr.Validation(apperr.Fields{
				"sort_by": {fmt.Sprintf("must be one of: %s", strings.Join(sortable.Fields(), ", "))},
			})
		}
		p.SortColumn = col
	}

	switch strings.ToLower(raw.SortOrder) {
	case "":
	case SortAsc:
		p.SortOrder = SortAsc
	case SortDesc:
		p.SortOrder = SortDesc
	default:
		return Params{}, apperr.Validation(apperr.Fields{
			"sort_order": {"must be one of: asc, desc"},
		})
	}

	return p, nil
}

// Meta is the pagination block returned in the response envelope.
type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

func NewMeta(p Params, total int64) Meta {
	totalPages := 0
	if p.PageSize > 0 {
		totalPages = int((total + int64(p.PageSize) - 1) / int64(p.PageSize))
	}
	return Meta{
		Page:       p.Page,
		PageSize:   p.PageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    p.Page < totalPages,
		HasPrev:    p.Page > 1 && total > 0,
	}
}

// Page is a page of results plus its meta.
type Page[T any] struct {
	Items []T
	Meta  Meta
}

func NewPage[T any](items []T, p Params, total int64) Page[T] {
	if items == nil {
		items = []T{}
	}
	return Page[T]{Items: items, Meta: NewMeta(p, total)}
}
