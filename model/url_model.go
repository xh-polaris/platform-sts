package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/monc"
)

const UrlCollectionName = "url"

var _ UrlModel = (*customUrlModel)(nil)

type (
	// UrlModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUrlModel.
	UrlModel interface {
		urlModel
	}

	customUrlModel struct {
		*defaultUrlModel
	}
)

// NewUrlModel returns a model for the mongo.
func NewUrlModel(url, db string, c cache.CacheConf) UrlModel {
	conn := monc.MustNewModel(url, db, UrlCollectionName, c)
	return &customUrlModel{
		defaultUrlModel: newDefaultUrlModel(conn),
	}
}
