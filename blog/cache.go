package blog

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
)

const Key_SQL_GetPostInfo = `SELECT post.*, ` +
	`IFNULL(poststatistics.star1,0), ` +
	`IFNULL(poststatistics.star2,0), ` +
	`IFNULL(poststatistics.star3,0), ` +
	`IFNULL(poststatistics.star4,0), ` +
	`IFNULL(poststatistics.star5,0)  ` +
	`FROM post ` +
	`LEFT JOIN poststatistics ` +
	`ON post.id = poststatistics.postid ` +
	`WHERE post.id = `

const Key_SQL_loadPost = `select * from post where id = `

var rdsUpdatingCount int64

const rdsUpdatingLimit = 500

func DBGetCache(key string, data interface{}) error {
	if !dbcache {
		return errors.New("not use Database Cache")
	}

	return GetCache(key, data)
}

func DBUpdateCache(key string, data interface{}) error {
	if !dbcache {
		return errors.New("not use Database Cache")
	}

	return UpdateCacheWithLimit(key, data)
}

func DBRemoveCache(key string) error {
	if !dbcache {
		return errors.New("not use Database Cache")
	}

	return RemoveCache(key)
}

func GetCache(key string, data interface{}) error {
	v, err := checkKey(key)
	if err != nil {
		Debug("GetCache fail:" + err.Error())
		return err
	}

	if err := decodeJson([]byte(v), data); err != nil {
		Debug("GetCache fail:" + err.Error())
		return err
	}

	return nil
}

func accessLimit(counter *int64, limit int64, name string, fn func() error) error {
	count := atomic.AddInt64(counter, 1)
	defer func() {
		atomic.AddInt64(counter, -1)
	}()

	if count > limit {
		err := fmt.Errorf("%s() reach limit %d", name, limit)
		Debug(err.Error())
		return err
	}

	return fn()
}

func UpdateCacheWithLimit(key string, data interface{}) error {
	return accessLimit(
		&rdsUpdatingCount, rdsUpdatingLimit,
		"UpdateCache",
		func() error {
			return UpdateCache(key, data)
		})
}

func UpdateCache(key string, data interface{}) error {
	//	Debug("UpdateCache: key=[" + key + "]")
	err := rdb.Set(context.Background(),
		key,
		encodeJson(data),
		sessionTimeout,
	).Err()
	if err != nil {
		Debug("UpdateCache fail:" + err.Error())
	}

	return err
}

func RemoveCache(key string) error {
	Debug("RemoveCacheCache: key=[" + key + "]")
	err := removeKey(key)
	if err != nil {
		Debug("RemoveCache fail:" + err.Error())
	}
	return err
}
