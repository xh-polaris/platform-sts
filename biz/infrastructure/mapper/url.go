package mapper

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/monc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/xh-polaris/platform-sts/biz/infrastructure/config"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/consts"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/data/db"
)

const (
	UrlCollectionName = "url"
	prefixUrlCacheKey = "cache:url:"
)

var _ UrlMapper = (*urlMapper)(nil)

type (
	// UrlMapper is an interface to be customized, add more methods here,
	// and implement the added methods in customUrlModel.
	UrlMapper interface {
		FindOneByPath(ctx context.Context, path string) (*db.Url, error)
		Insert(ctx context.Context, data *db.Url) error
		FindOne(ctx context.Context, id string) (*db.Url, error)
		Update(ctx context.Context, data *db.Url) error
		Delete(ctx context.Context, id primitive.ObjectID) error
	}

	urlMapper struct {
		conn *monc.Model
	}
)

// NewUrlMapper returns a model for the mongo.
func NewUrlMapper(config *config.Config) UrlMapper {
	conn := monc.MustNewModel(config.Mongo.URL, config.Mongo.DB, UrlCollectionName, config.CacheConf)
	return &urlMapper{
		conn: conn,
	}
}

func (m *urlMapper) FindOneByPath(ctx context.Context, path string) (*db.Url, error) {
	url := db.Url{}
	err := m.conn.FindOneNoCache(ctx, &url, bson.M{"url": bson.M{"$regex": ".*" + path}})
	switch err {
	case nil:
		return &url, nil
	case monc.ErrNotFound:
		return nil, consts.ErrNotFound
	default:
		return nil, err
	}
}

func (m *urlMapper) Insert(ctx context.Context, data *db.Url) error {
	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
		data.CreateAt = time.Now()
		data.UpdateAt = time.Now()
	}

	key := prefixUrlCacheKey + data.ID.Hex()
	_, err := m.conn.InsertOne(ctx, key, data)
	return err
}

func (m *urlMapper) FindOne(ctx context.Context, id string) (*db.Url, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, consts.ErrInvalidObjectId
	}

	var data db.Url
	key := prefixUrlCacheKey + id
	err = m.conn.FindOne(ctx, key, &data, bson.M{"_id": oid})
	switch err {
	case nil:
		return &data, nil
	case monc.ErrNotFound:
		return nil, consts.ErrNotFound
	default:
		return nil, err
	}
}

func (m *urlMapper) Update(ctx context.Context, data *db.Url) error {
	data.UpdateAt = time.Now()
	key := prefixUrlCacheKey + data.ID.Hex()
	_, err := m.conn.ReplaceOne(ctx, key, bson.M{"_id": data.ID}, data)
	return err
}

func (m *urlMapper) Delete(ctx context.Context, id primitive.ObjectID) error {
	key := prefixUrlCacheKey + id.String()
	_, err := m.conn.DeleteOne(ctx, key, bson.M{"_id": id})
	return err
}
