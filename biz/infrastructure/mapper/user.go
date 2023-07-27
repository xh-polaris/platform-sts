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
	UserCollectionName = "user"
	prefixUserCacheKey = "cache:user:"
)

var _ UserMapper = (*userMapper)(nil)

type (
	// UserMapper is an interface to be customized, add more methods here,
	// and implement the added methods in customUserMapper.
	UserMapper interface {
		FindOneByAuth(ctx context.Context, auth db.Auth) (*db.User, error)
		Insert(ctx context.Context, data *db.User) error
		FindOne(ctx context.Context, id string) (*db.User, error)
		Update(ctx context.Context, data *db.User) error
		Delete(ctx context.Context, id string) error
	}

	userMapper struct {
		conn *monc.Model
	}
)

// NewUserMapper returns a mapper for the mongo.
func NewUserMapper(config *config.Config) UserMapper {
	conn := monc.MustNewModel(config.Mongo.URL, config.Mongo.DB, UserCollectionName, config.CacheConf)
	return &userMapper{
		conn: conn,
	}
}

func (m *userMapper) FindOneByAuth(ctx context.Context, auth db.Auth) (*db.User, error) {
	var data db.User
	err := m.conn.FindOneNoCache(ctx, &data, bson.M{"auth": auth})
	switch err {
	case nil:
		return &data, nil
	case monc.ErrNotFound:
		return nil, consts.ErrNotFound
	default:
		return nil, err
	}
}

func (m *userMapper) Insert(ctx context.Context, data *db.User) error {
	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
		data.CreateAt = time.Now()
		data.UpdateAt = time.Now()
	}

	key := prefixUserCacheKey + data.ID.Hex()
	_, err := m.conn.InsertOne(ctx, key, data)
	return err
}

func (m *userMapper) FindOne(ctx context.Context, id string) (*db.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, consts.ErrInvalidObjectId
	}

	var data db.User
	key := prefixUserCacheKey + id
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

func (m *userMapper) Update(ctx context.Context, data *db.User) error {
	data.UpdateAt = time.Now()
	key := prefixUserCacheKey + data.ID.Hex()
	_, err := m.conn.UpdateOne(ctx, key, bson.M{"_id": data.ID}, bson.M{"$set": data})
	return err
}

func (m *userMapper) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return consts.ErrInvalidObjectId
	}
	key := prefixUserCacheKey + id
	_, err = m.conn.DeleteOne(ctx, key, bson.M{"_id": oid})
	return err
}
