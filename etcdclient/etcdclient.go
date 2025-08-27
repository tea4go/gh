package etcdclient

import (
	"context"
	"strings"
	"sync"
	"time"

	"go.etcd.io/etcd/clientv3"
)

const (
	ConnTimeout = time.Second * 3
	OperTimeout = time.Second * 5
)

type TEtcdClient struct {
	client *clientv3.Client
	head   string
}

var instance *TEtcdClient
var once sync.Once

/**
 * 初始化一个单例,一般用于程序启动时
 */
func InitInstance() *TEtcdClient {
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
func GetInstance() *TEtcdClient {
	if instance == nil {
		instance = InitInstance()
	}
	return instance
}

/**
 * 获取一个线程安全的单例
 */
func GetSafeInstance() *TEtcdClient {
	once.Do(func() {
		instance = InitInstance()
	})
	return instance
}

/**
 * 连接etcd
 */
func (etcd *TEtcdClient) GetHead() string {
	return etcd.head
}

func (etcd *TEtcdClient) SetHead(head string) {
	if head != "" {
		if head[0] == '/' {
			etcd.head = head
		} else {
			etcd.head = "/" + head
		}
		if head[len(head)-1] != '/' {
			etcd.head += "/"
		}
	} else {
		etcd.head = ""
	}
}

/**
 * Set Value
 */
func (etcd *TEtcdClient) Set(key string, value string, args ...int) error {
	ctx, _ := context.WithTimeout(context.Background(), OperTimeout)
	if len(args) > 0 {
		expires := args[0]
		lease, err := etcd.client.Grant(ctx, int64(expires))
		if err != nil {
			return err
		}
		opp := clientv3.WithLease(lease.ID)
		_, err = etcd.client.Put(ctx, etcd.head+key, value, opp)
		if err == nil {
			return nil
		} else {
			return err
		}
	} else {
		_, err := etcd.client.Put(ctx, etcd.head+key, value)
		if err == nil {
			return nil
		} else {
			return err
		}
	}
}

/**
 * hash get,获取一个map键值对结构,对于排序的结构从ectd查出来是有序的,但map不保证有序性，所以放入map后是无序的
 */
func (etcd *TEtcdClient) HGet(getResp *clientv3.GetResponse, err error) (map[string]string, int64, error) {
	result := make(map[string]string)
	if err != nil {
		return result, 0, err
	}

	var key string
	for _, v := range getResp.Kvs {
		key = strings.Replace(string(v.Key), etcd.head, "", 1)
		result[key] = string(v.Value)
	}
	return result, getResp.Count, nil
}

/**
 * Get Single Key
 */
func (etcd *TEtcdClient) Get(key string) (string, error) {
	ctx, _ := context.WithTimeout(context.Background(), OperTimeout)
	getResponse, error := etcd.client.Get(ctx, etcd.head+key)
	result, _, err := etcd.HGet(getResponse, error)
	return result[key], err
}

/**
 * Get By prefix Mutiple Key
 */
func (etcd *TEtcdClient) GetAll(prefix string) (map[string]string, int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), OperTimeout)
	withPrefix := clientv3.WithPrefix()
	return etcd.HGet(etcd.client.Get(ctx, etcd.head+prefix, withPrefix))
}

/**
 * 获取最大键,用于获取最大ID,比如Key_001 ... Key_102 最大为Key_102
 */
func (etcd *TEtcdClient) GetMaxKey(prefix string) (string, error) {
	ctx, _ := context.WithTimeout(context.Background(), OperTimeout)
	withPrefix := clientv3.WithPrefix()
	withSort := clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend)
	withLimit := clientv3.WithLimit(1)
	resp, err := etcd.client.Get(ctx, etcd.head+prefix, withPrefix, withSort, withLimit)
	if err != nil {
		return "", err
	}

	for _, v := range resp.Kvs {
		return string(v.Key), nil
	}
	//没有数据
	return "", nil
}

/**
 * Count By prefix data
 */
func (etcd *TEtcdClient) Count(prefix string) (int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), OperTimeout)
	withCount := clientv3.WithCountOnly()
	withPrefix := clientv3.WithPrefix()
	ret, err := etcd.client.Get(ctx, etcd.head+prefix, withPrefix, withCount)
	if err != nil {
		return 0, err
	} else {
		return ret.Count, err
	}
}

/**
 * Get By prefix Limit N
 */
func (etcd *TEtcdClient) GetLimit(prefix string, limit int) (map[string]string, int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), OperTimeout)
	withPrefix := clientv3.WithPrefix()
	withLimit := clientv3.WithLimit(int64(limit))
	return etcd.HGet(etcd.client.Get(ctx, etcd.head+prefix, withPrefix, withLimit))
}

/**
 * Get By Range,Not Contains endKey,[startKey,endKey)
 */
func (etcd *TEtcdClient) GetRange(startKey string, endKey string) (map[string]string, int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), OperTimeout)
	withRange := clientv3.WithRange(etcd.head + "/" + endKey)
	return etcd.HGet(etcd.client.Get(ctx, etcd.head+startKey, withRange))
}

/**
 * Get By Range,Contains StartKey[startKey,N-1]
 */
func (etcd *TEtcdClient) GetRangeLimit(startKey string, limit int) (map[string]string, int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), OperTimeout)
	withLimit := clientv3.WithLimit(int64(limit))
	withFrom := clientv3.WithFromKey()
	return etcd.HGet(etcd.client.Get(ctx, etcd.head+startKey, withFrom, withLimit))
}

/**
 * Delete One
 */
func (etcd *TEtcdClient) Del(key string) (int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), OperTimeout)
	ret, err := etcd.client.Delete(ctx, etcd.head+key)
	if err != nil {
		return 0, err
	}
	return ret.Deleted, nil
}

/**
 * Delete All By Prefix
 */
func (etcd *TEtcdClient) DelAll(prefix string) (int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), OperTimeout)
	withPrefix := clientv3.WithPrefix()
	ret, err := etcd.client.Delete(ctx, etcd.head+prefix, withPrefix)
	if err != nil {
		return 0, err
	}
	return ret.Deleted, nil
}

/**
 * 返回原生接口
 */
func (etcd *TEtcdClient) GetClient() *clientv3.Client {
	return etcd.client
}

/**
 * 返回原生接口
 */
func (etcd *TEtcdClient) GetKV() clientv3.KV {
	return etcd.client.KV
}

/**
 * 连接etcd
 */
func Connect(etcdAddrs []string) (*TEtcdClient, error) {
	etcd := &TEtcdClient{head: "tea4go"}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdAddrs,
		DialTimeout: ConnTimeout,
	})
	if err != nil {
		return nil, err
	}

	etcd.client = cli
	return etcd, nil
}
