package blog

import (
	"context"
	"errors"
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

	return UpdateCache(key, data)
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

func UpdateCache(key string, data interface{}) error {
	Debug("UpdateCache: key=[" + key + "]")
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
