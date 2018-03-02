package pagination

import (
	"math"
	"strconv"

	"github.com/eyecuelab/kit/web/meta"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

const resultsPerPage = 20

// Data pagination data for jsonapi
var Data meta.Pagination

// Apply apply pagination to a provided scope
func Apply(c echo.Context, scope *gorm.DB, model interface{}, list interface{}, perPage int) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page == 0 {
		page = 1
	}

	var count int
	if err := scope.Offset(0).Model(model).Count(&count).Error; err != nil {
		return err
	}
	setData(count, resultsPerPage, page)

	strPerPage := c.QueryParam("per_page")
	if strPerPage != "" {
		perPage, _ = strconv.Atoi(strPerPage)
	} else if perPage == 0 {
		perPage = resultsPerPage
	}

	if err := scope.Offset(offset(page, perPage)).Limit(perPage).
		Find(list).Error; err != nil {
		return err
	}

	return nil
}

func setData(count int, perPage int, page int) {
	Data.Count = count
	pages := float64(count) / float64(perPage)
	Data.Max = int(math.Ceil(pages))
	Data.Page = page
}

func offset(pageNumber int, resultsPerPage int) int {
	return (pageNumber - 1) * resultsPerPage
}