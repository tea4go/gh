package redisclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	rediss "github.com/go-redis/redis/v8"
)

const (
	ConnTimeout = time.Second * 3
	OperTimeout = time.Second * 5
)

type TRedisClient struct {
	SavePath              string `json:"save_path"`
	Poolsize              int    `json:"poolsize"`
	Password              string `json:"password"`
	DbNum                 int    `json:"db_num"`
	IdleTimeoutStr        string `json:"idle_timeout"`
	IdleCheckFrequencyStr string `json:"idle_check_frequency"`
	MaxRetries            int    `json:"max_retries"`
	maxlifetime           int64
	idleTimeout           time.Duration
	idleCheckFrequency    time.Duration
	poollist              *rediss.ClusterClient
	head                  string
}

var instance *TRedisClient
var once sync.Once

// GetClient 获取 Redis 客户端
func (rp *TRedisClient) GetClient() *rediss.ClusterClient {
	return rp.poollist
}

//订阅主题
// Sub 订阅主题
func (rp *TRedisClient) Sub(topic string) *rediss.PubSub {
	var ctx = context.Background()
	return rp.poollist.PSubscribe(ctx, rp.head+topic)
}

//发送主题
// Pub 发布消息
func (rp *TRedisClient) Pub(topic, msg string) (int64, error) {
	var ctx = context.Background()
	return rp.poollist.Publish(ctx, topic, msg).Result()
}

// Init 初始化
func (rp *TRedisClient) Init(cfgStr string) error {
	var ctx = context.Background()
	err := json.Unmarshal([]byte(cfgStr), rp)
	if err != nil {
		fmt.Println("SessionInit Redis", err)
		return err
	}
	rp.idleTimeout, err = time.ParseDuration(rp.IdleTimeoutStr)
	if err != nil {
		return err
	}

	rp.idleCheckFrequency, err = time.ParseDuration(rp.IdleCheckFrequencyStr)
	if err != nil {
		return err
	}

	rp.poollist = rediss.NewClusterClient(&rediss.ClusterOptions{
		Addrs:              strings.Split(rp.SavePath, ";"),
		Password:           rp.Password,
		PoolSize:           rp.Poolsize,
		IdleTimeout:        rp.idleTimeout,
		IdleCheckFrequency: rp.idleCheckFrequency,
		MaxRetries:         rp.MaxRetries,
	})
	return rp.poollist.Ping(ctx).Err()
}

// GetHead 获取头部
func (Self *TRedisClient) GetHead() string {
	return Self.head
}

// SetHead 设置头部
func (Self *TRedisClient) SetHead(head string) {
	if head != "" {
		if head[0] == '/' {
			Self.head = head
		} else {
			Self.head = "/" + head
		}
		if head[len(head)-1] != '/' {
			Self.head += "/"
		}
	} else {
		Self.head = ""
	}
}

/**
 * Set Value
 */
// Set 设置值
func (Self *TRedisClient) Set(key string, value string, args ...interface{}) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), OperTimeout)
	defer cancel()
	if len(args) > 0 {
		expires := args[0].(time.Duration)
		_, err := Self.poollist.Set(ctx, Self.head+key, value, expires).Result()
		if err != nil {
			return false, err
		}
	} else {
		_, err := Self.poollist.Set(ctx, Self.head+key, value, 0).Result()
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

/**
 * Exists Single Key
 */
// Exists 检查 Key 是否存在
func (Self *TRedisClient) Exists(key string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), OperTimeout)
	defer cancel()
	n, _ := Self.poollist.Exists(ctx, Self.head+key).Result()
	return n == 1
}

/**
 * Get Single Key
 */
// Get 获取值
func (Self *TRedisClient) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), OperTimeout)
	defer cancel()
	value, err := Self.poollist.Get(ctx, Self.head+key).Result()
	if err == rediss.Nil {
		return value, fmt.Errorf("键值%s不存在！", key)
	} else {
		return value, err
	}
}

/**
 * GetExpire Single Key
 */
// GetExpire 获取过期时间
func (Self *TRedisClient) GetExpire(key string) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), OperTimeout)
	defer cancel()
	value, err := Self.poollist.TTL(ctx, Self.head+key).Result()
	if err == rediss.Nil {
		return value, fmt.Errorf("键值%s不存在！", key)
	} else {
		return value, err
	}
}

/**
 * Del Single Key
 */
// Del 删除 Key
func (Self *TRedisClient) Del(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), OperTimeout)
	defer cancel()
	n, err := Self.poollist.Del(ctx, Self.head+key).Result()
	if err == rediss.Nil {
		return n > 0, fmt.Errorf("键值%s不存在！", key)
	} else {
		return n > 0, err
	}
}
