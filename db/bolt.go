package db

import (
	"encoding/json"
	"go.etcd.io/bbolt"
)

var db *bbolt.DB

// 初始化数据库
func InitDB(path string) error {
	var err error
	db, err = bbolt.Open(path, 0600, nil)
	if err != nil {
		return err
	}

	// 创建必要的bucket
	return db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("tasks"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("settings"))
		return err
	})
}

// 关闭数据库
func CloseDB() error {
	return db.Close()
}

// 保存数据
func SaveData(bucket string, key string, data interface{}) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return bbolt.ErrBucketNotFound
		}

		value, err := json.Marshal(data)
		if err != nil {
			return err
		}

		return b.Put([]byte(key), value)
	})
}

// 获取数据
func GetData(bucket string, key string, result interface{}) error {
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return bbolt.ErrBucketNotFound
		}

		value := b.Get([]byte(key))
		if value == nil {
			return nil // 没有找到数据
		}

		return json.Unmarshal(value, result)
	})
}

// 获取所有键
func GetAllKeys(bucket string) ([][]byte, error) {
	var keys [][]byte
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return bbolt.ErrBucketNotFound
		}

		return b.ForEach(func(k, v []byte) error {
			keys = append(keys, append([]byte(nil), k...))
			return nil
		})
	})
	return keys, err
}

// 删除数据
func DeleteData(bucket string, key string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		return b.Delete([]byte(key))
	})
}
