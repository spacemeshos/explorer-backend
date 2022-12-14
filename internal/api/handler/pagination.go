package handler

import (
	"github.com/labstack/echo/v4"
	"strconv"
)

type PaginationMetadata struct {
	TotalCount  int64 `json:"totalCount"`
	PageCount   int64 `json:"pageCount"`
	PerPage     int64 `json:"perPage"`
	Next        int64 `json:"next"`
	HasNext     bool  `json:"hasNext"`
	HasPrevious bool  `json:"hasPrevious"`
	Current     int64 `json:"current"`
	Previous    int64 `json:"previous"`
}

func GetPagination(c echo.Context) (pageNumber, pageSize int64) {
	pageNumber = 1
	pageSize = 20
	if page := c.QueryParam("page"); page != "" {
		pageNumber, _ = strconv.ParseInt(page, 10, 32)
		if pageNumber <= 0 {
			pageNumber = 1
		}
	}
	if size := c.QueryParam("pagesize"); size != "" {
		pageSize, _ = strconv.ParseInt(size, 10, 32)
		if pageSize <= 0 {
			pageSize = 20
		}
	}
	return pageNumber, pageSize
}

func GetPaginationMetadata(total int64, pageNumber int64, pageSize int64) PaginationMetadata {
	pageCount := (total + pageSize - 1) / pageSize
	result := PaginationMetadata{
		TotalCount: total,
		PageCount:  pageNumber,
		PerPage:    pageSize,
		Next:       pageCount,
		Current:    pageNumber,
		Previous:   1,
	}
	if pageNumber < pageCount {
		result.Next = pageNumber + 1
		result.HasNext = true
	}
	if pageNumber > 1 {
		result.Previous = pageNumber - 1
		result.HasPrevious = true
	}
	return result
}
