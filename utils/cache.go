package utils

import (
	"bytes"
	"encoding/gob"
	"errors"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	_ "github.com/astaxie/beego/cache/memcache"
	_ "github.com/astaxie/beego/cache/redis"
)

var cc cache.Cache

// InitCache 根据conf初始化cache
func InitCache() {
	cacheConfig := beego.AppConfig.String("cache::cache")
	cc = nil

	if "redis" == cacheConfig {
		initRedis()
	} else {
		initMemcache()
	}

}

func initMemcache() {
	var err error
	cc, err = cache.NewCache("memcache", StringsJoin(`{"conn":"`, beego.AppConfig.String("cache::memcache_host"), `"}`))

	if err != nil {
		beego.Info(err)
	}

}

func initRedis() {
	LogOut("info", "缓存采用的是redis")
	// cc = &cache.Cache{}
	var err error

	defer func() {
		if r := recover(); r != nil {
			cc = nil
		}
	}()
	host := beego.AppConfig.String("cache::redis_host")
	LogOut("info", StringsJoin("连接参数:", host))
	cc, err = cache.NewCache("redis", StringsJoin(`{"conn":"`, host, `"}`))

	if err != nil {
		LogOut("error", err)
	}
}

// SetCache 设置缓存
func SetCache(key string, value interface{}, timeout int) error {
	data, err := Encode(value)
	if err != nil {
		return err
	}
	if cc == nil {
		return errors.New("cc is nil")
	}

	defer func() {
		if r := recover(); r != nil {
			LogOut("error", r)
			cc = nil
		}
	}()
	timeouts := time.Duration(timeout) * time.Second
	err = cc.Put(key, data, timeouts)
	if err != nil {
		LogOut("error", err)
		LogOut("error", StringsJoin("SetCache失败，key:", key))

		return err
	}

	return nil

}

// GetCache 获得缓存信息
func GetCache(key string, to interface{}) error {
	if cc == nil {
		return errors.New("cc is nil")
	}

	defer func() {
		if r := recover(); r != nil {
			LogOut("error", r)
			cc = nil
		}
	}()

	data := cc.Get(key)
	if data == nil {
		return errors.New("Cache不存在")
	}

	err := Decode(data.([]byte), to)
	if err != nil {
		LogOut("error", err)
		LogOut("error", StringsJoin("GetCache失败，key:", key))
	}

	return err
}

// DelCache 删除缓存信息
func DelCache(key string) error {
	if cc == nil {
		return errors.New("cc is nil")
	}

	defer func() {
		if r := recover(); r != nil {
			cc = nil
		}
	}()

	err := cc.Delete(key)
	if err != nil {
		return errors.New("Cache删除失败")
	}
	return nil

}

// Encode 用gob进行数据编码
func Encode(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode 用gob进行数据解码
func Decode(data []byte, to interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(to)
}
