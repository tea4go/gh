package nustdbclient

import (
	"strings"
	"sync"
	"time"

	"github.com/nutsdb/nutsdb"
)

const (
	ConnTimeout = time.Second * 3
	OperTimeout = time.Second * 5
)

type TNustDBField struct {
	Key   string
	Value string
}
type TNustDBClient struct {
	db     *nutsdb.DB
	bucket string
	head   string
}

var instance *TNustDBClient
var once sync.Once

/**
 * 初始化一个单例,一般用于程序启动时
 */
func InitInstance(bucket_name, db_path string) *TNustDBClient {
	if instance == nil {
		db, err := nutsdb.Open(
			nutsdb.DefaultOptions,
			nutsdb.WithDir(db_path),
		)
		if err != nil {
			panic(err)
		}

		err = db.Update(
			func(tx *nutsdb.Tx) error {
				if !tx.ExistBucket(nutsdb.DataStructureBTree, bucket_name) {
					return tx.NewBucket(nutsdb.DataStructureBTree, bucket_name)
				}
				return nil
			})
		if err != nil {
			panic(err)
		}

		instance = &TNustDBClient{
			db:     db,
			bucket: bucket_name,
		}

	}
	return instance
}

/**
 * 获取一个单例,可以用这个不需要考虑线程安全
 */
func GetInstance(bucket_name, db_path string) *TNustDBClient {
	if instance == nil {
		instance = InitInstance(bucket_name, db_path)
	}
	return instance
}

/**
 * 获取一个线程安全的单例
 */
func GetSafeInstance(bucket_name, db_path string) *TNustDBClient {
	once.Do(func() {
		instance = InitInstance(bucket_name, db_path)
	})
	return instance
}

/**
 * 连接etcd
 */
func (d *TNustDBClient) GetHead() string {
	return d.head
}

func (d *TNustDBClient) SetHead(head string) {
	if head != "" {
		if head[0] == '/' {
			d.head = head
		} else {
			d.head = "/" + head
		}
		if head[len(head)-1] != '/' {
			d.head += "/"
		}
	} else {
		d.head = ""
	}
}

/*
*
  - Set Value
    ttl : NusDB支持TTL(存活时间)的功能，可以对指定的bucket里的key过期时间的设置
*/
func (s *TNustDBClient) Set(keyname string, value string, args ...int) error {
	var ttl uint32
	if len(args) >= 1 {
		ttl = uint32(args[0])
	}
	err := s.db.Update(
		func(tx *nutsdb.Tx) error {
			return tx.Put(s.bucket, []byte(s.head+keyname), []byte(value), ttl)
		})

	return err
}

/**
 * Get Single Key
 */
func (s *TNustDBClient) Get(keyname string) (value string, err error) {
	err = s.db.View(
		func(tx *nutsdb.Tx) error {
			v, err := tx.Get(s.bucket, []byte(s.head+keyname))
			if err != nil {
				return err
			}
			value = string(v)
			return nil
		})

	return
}

func (s *TNustDBClient) GetAll(keyname string) (items []*TNustDBField, err error) {
	err = s.db.View(
		func(tx *nutsdb.Tx) error {
			keys, values, err := tx.GetAll(s.bucket)
			if err != nil {
				return err
			}

			for k, key := range keys {
				if keyname == "" || strings.HasPrefix(string(key), s.head+keyname) {
					tmp := string(key)
					tmp = strings.Replace(tmp, s.head, "", 1)
					items = append(items, &TNustDBField{Key: tmp, Value: string(values[k])})
				}
			}

			return nil
		})

	return items, err
}

/**
 * Delete One
 */
func (s *TNustDBClient) Del(keyname string) error {
	err := s.db.Update(
		func(tx *nutsdb.Tx) error {
			return tx.Delete(s.bucket, []byte(s.head+keyname))
		})

	return err
}

func (s *TNustDBClient) DelAll(keyname string) error {
	err := s.db.Update(
		func(tx *nutsdb.Tx) error {
			keys, err := tx.GetKeys(s.bucket)
			if err != nil {
				return err
			}

			for _, key := range keys {
				if keyname == "" || strings.HasPrefix(string(key), s.head+keyname) {
					tx.Delete(s.bucket, key)
				}
			}

			return nil
		})
	return err
}
