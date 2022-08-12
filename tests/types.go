package tests

import "github.com/spacemeshos/explorer-backend/model"

type layerResp struct {
	Data       []model.Layer `json:"data"`
	Pagination pagination    `json:"pagination"`
}

type epochResp struct {
	Data       []model.Epoch `json:"data"`
	Pagination pagination    `json:"pagination"`
}

type transactionResp struct {
	Data       []model.Transaction `json:"data"`
	Pagination pagination          `json:"pagination"`
}

type smesherResp struct {
	Data       []model.Smesher `json:"data"`
	Pagination pagination      `json:"pagination"`
}

type rewardResp struct {
	Data       []model.Reward `json:"data"`
	Pagination pagination     `json:"pagination"`
}

type accountResp struct {
	Data       []model.Account `json:"data"`
	Pagination pagination      `json:"pagination"`
}

type atxResp struct {
	Data       []model.Activation `json:"data"`
	Pagination pagination         `json:"pagination"`
}
type blockResp struct {
	Data       []model.Block `json:"data"`
	Pagination pagination    `json:"pagination"`
}

type appResp struct {
	Data       []model.App `json:"data"`
	Pagination pagination  `json:"pagination"`
}

type redirect struct {
	Redirect string `json:"redirect"`
}

type pagination struct {
	TotalCount  int  `json:"totalCount"`
	PageCount   int  `json:"pageCount"`
	PerPage     int  `json:"perPage"`
	Next        int  `json:"next"`
	HasNext     bool `json:"hasNext"`
	Current     int  `json:"current"`
	Previous    int  `json:"previous"`
	HasPrevious bool `json:"hasPrevious"`
}
