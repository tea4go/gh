package nustdbclient

import (
	"sync"
	"time"

	"github.com/nutsdb/nutsdb"
)

const (
	ConnTimeout = time.Second * 3
	OperTimeout = time.Second * 5
)

type TNustDBClient struct {
	db     *nutsdb.DB
	bucket string
}

var instance *TNustDBClient
var once sync.Once

/**
 * 初始化一个单例,一般用于程序启动时
 */
func InitInstance() *TNustDBClient {
	if instance == nil {
		client, err := Connect([]string{"127.0.0.1:2379"})
		if err == nil {
			instance = client
		}
	}
	return instance
}

/**
 * 获取一个单例,可以用这个不需要考虑线程安全
 */
func GetInstance() *TNustDBClient {
	if instance == nil {
		instance = InitInstance()
	}
	return instance
}

/**
 * 获取一个线程安全的单例
 */
func GetSafeInstance() *TNustDBClient {
	once.Do(func() {
		instance = InitInstance()
	})
	return instance
}
