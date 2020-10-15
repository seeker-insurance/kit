package pagination

import (
	"math"
	"strconv"

	"net/url"
)

type (
	Pagination struct {
		Count   int `json:"item_count"`
		Max     int `json:"max"`
		Page    int `json:"page"`
		PerPage int
		Offset  int
		Url     url.URL
	}
)

func (p *Pagination) Links() map[string]string {
	pageValues := map[string]int{"self": p.Page}

	if p.Max != 1 {
		pageValues["first"] = 1
		pageValues["last"] = p.Max
	}

	if p.Max != p.Page {
		pageValues["next"] = p.Page + 1
	}

	if p.Page > 1 {
		pageValues["prev"] = p.Page - 1
	}

	return p.linkify(pageValues)
}

func (p Pagination) linkify(pageValues map[string]int) map[string]string {
	links := make(map[string]string)
	for k, v := range pageValues {
		links[k] = p.withPageNumber(v)
	}
	return links
}

func (p Pagination) withPageNumber(num int) string {
	q := p.Url.Query()
	q.Set("page[number]", strconv.Itoa(num))
	p.Url.RawQuery = q.Encode()

	return p.Url.RequestURI()
}

// DefaultPerPage ...
var DefaultPerPage = 20

// MaxPerPage ...
var MaxPerPage = 100

// Data pagination data for jsonapi
var Data Pagination

// SetData sets data to show pagination
func SetData(count int, perPage int, page int) {
	Data.Count = count
	pages := float64(count) / float64(perPage)
	Data.Max = int(math.Ceil(pages))
	Data.Page = page
}

func offset(pageNumber int, perPage int) int {
	return (pageNumber - 1) * perPage
}
