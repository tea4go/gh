package nustdbclient

import (
	"fmt"
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
					fmt.Println("NewBucket =", bucket_name)
					return tx.NewBucket(nutsdb.DataStructureBTree, bucket_name)
				}
				fmt.Println("Exist Bucket =", bucket_name)
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

/**
 * 连接etcd
 */
func (d *TNustDBClient) GetBucketName() string {
	return d.bucket
}

func (d *TNustDBClient) SetBucketName(bucket_name string, args ...uint16) (err error) {
	if bucket_name != "" {
		var ds uint16
		ds = nutsdb.DataStructureBTree
		if len(args) >= 1 {
			ds = args[0]
		}

		err = d.db.Update(
			func(tx *nutsdb.Tx) error {
				if !tx.ExistBucket(ds, bucket_name) {
					fmt.Println("SetBucketName NewBucket =", bucket_name)
					return tx.NewBucket(ds, bucket_name)
				}
				fmt.Println("SetBucketName Exist Bucket =", bucket_name)
				return nil
			})

		if err == nil {
			d.bucket = bucket_name
		}
	}
	return err
}

func (s *TNustDBClient) LPush(keyname string, value string) error {
	err := s.db.Update(
		func(tx *nutsdb.Tx) error {
			fmt.Println("s.bucket", s.bucket)
			return tx.LPush(s.bucket, []byte(s.head+keyname), []byte(value))
		})

	return err
}

func (s *TNustDBClient) LPushByBucket(bucket_name, keyname string, value string) error {
	if bucket_name == "" {
		bucket_name = s.bucket
	}

	err := s.db.Update(
		func(tx *nutsdb.Tx) error {
			fmt.Println("LPush", bucket_name)
			return tx.LPush(bucket_name, []byte(s.head+keyname), []byte(value))
		})

	return err
}

func (s *TNustDBClient) LRangeByBucket(bucket_name, keyname string) (items []string, err error) {
	if bucket_name == "" {
		bucket_name = s.bucket
	}
	s.db.View(
		func(tx *nutsdb.Tx) (err error) {
			datas, err := tx.LRange(bucket_name, []byte(s.head+keyname), 0, -1)
			for _, v := range datas {
				items = append(items, string(v))
			}
			return err
		})
	return
}

/*
*
  - Set Value
    ttl : NusDB支持TTL(存活时间)的功能，可以对指定的bucket里的key过期时间的设置
*/
func (s *TNustDBClient) SetValue(keyname string, value string, args ...int) error {
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

func (s *TNustDBClient) SetValueByBucket(bucket_name, keyname string, value string, args ...int) error {
	if bucket_name == "" {
		bucket_name = s.bucket
	}

	var ttl uint32
	if len(args) >= 1 {
		ttl = uint32(args[0])
	}

	err := s.db.Update(
		func(tx *nutsdb.Tx) error {
			return tx.Put(bucket_name, []byte(s.head+keyname), []byte(value), ttl)
		})

	return err
}

func (s *TNustDBClient) GetValue(keyname string) (value string, err error) {
	err = s.db.View(
		func(tx *nutsdb.Tx) error {
			fmt.Println("s.bucket", s.bucket)
			v, err := tx.Get(s.bucket, []byte(s.head+keyname))
			if err != nil {
				return err
			}
			value = string(v)
			return nil
		})

	return
}

func (s *TNustDBClient) GetValueByBucket(bucket_name, keyname string) (value string, err error) {
	if bucket_name == "" {
		bucket_name = s.bucket
	}
	err = s.db.View(
		func(tx *nutsdb.Tx) error {
			fmt.Println("s.bucket", s.bucket)
			v, err := tx.Get(bucket_name, []byte(s.head+keyname))
			if err != nil {
				return err
			}
			value = string(v)
			return nil
		})

	return
}

func (s *TNustDBClient) GetAllValue(keyname string) (items []*TNustDBField, err error) {
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
func (s *TNustDBClient) DelValue(keyname string) error {
	err := s.db.Update(
		func(tx *nutsdb.Tx) error {
			return tx.Delete(s.bucket, []byte(s.head+keyname))
		})

	return err
}

func (s *TNustDBClient) DelAllValue(keyname string) error {
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
