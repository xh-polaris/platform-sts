package model

import (
	"context"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/monc"
	"go.mongodb.org/mongo-driver/bson"
)

const UrlCollectionName = "url"

var _ UrlModel = (*customUrlModel)(nil)

type (
	// UrlModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUrlModel.
	UrlModel interface {
		urlModel
		FindOneByPath(ctx context.Context, path string) (*Url, error)
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

func (m *defaultUrlModel) FindOneByPath(ctx context.Context, path string) (*Url, error) {
	url := Url{}
	err := m.conn.FindOneNoCache(ctx, &url, bson.M{"url": bson.M{"$regex": ".*" + path}})
	switch err {
	case nil:
		return &url, nil
	case monc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
