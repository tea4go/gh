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

func (rp *TRedisClient) GetClient() *rediss.ClusterClient {
	return rp.poollist
}

//订阅主题
func (rp *TRedisClient) Sub(topic string) *rediss.PubSub {
	var ctx = context.Background()
	return rp.poollist.PSubscribe(ctx, rp.head+topic)
}

//发送主题
func (rp *TRedisClient) Pub(topic, msg string) (int64, error) {
	var ctx = context.Background()
	return rp.poollist.Publish(ctx, topic, msg).Result()
}

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

func (Self *TRedisClient) GetHead() string {
	return Self.head
}

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
func (Self *TRedisClient) Set(key string, value string, args ...interface{}) (bool, error) {
	ctx, _ := context.WithTimeout(context.Background(), OperTimeout)
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
func (Self *TRedisClient) Exists(key string) bool {
	ctx, _ := context.WithTimeout(context.Background(), OperTimeout)
	n, _ := Self.poollist.Exists(ctx, Self.head+key).Result()
	return n == 1
}

/**
 * Get Single Key
 */
func (Self *TRedisClient) Get(key string) (string, error) {
	ctx, _ := context.WithTimeout(context.Background(), OperTimeout)
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
func (Self *TRedisClient) GetExpire(key string) (time.Duration, error) {
	ctx, _ := context.WithTimeout(context.Background(), OperTimeout)
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
func (Self *TRedisClient) Del(key string) (bool, error) {
	ctx, _ := context.WithTimeout(context.Background(), OperTimeout)
	n, err := Self.poollist.Del(ctx, Self.head+key).Result()
	if err == rediss.Nil {
		return n > 0, fmt.Errorf("键值%s不存在！", key)
	} else {
		return n > 0, err
	}
}
