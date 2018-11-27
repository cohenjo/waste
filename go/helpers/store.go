package helpers

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/hashicorp/consul/api"
	"github.com/outbrain/golib/log"
)

// KVStoreType represents the possible KV store types.
type KVStoreType string

// The possible KV stores.
const (
	KVStoreTypeConsul KVStoreType = "CONSUL"
	KVStoreTypeRedis  KVStoreType = "REDIS"
	KVStoreTypeNone   KVStoreType = "NONE"
)

func (t *KVStoreType) String() string {
	return fmt.Sprintf("%v", *t)
}

func (t *KVStoreType) Set(value string) error {
	switch value {
	case string(KVStoreTypeConsul):
		*t = KVStoreTypeConsul
	case string(KVStoreTypeRedis):
		*t = KVStoreTypeRedis
	default:
		*t = KVStoreTypeNone
	}
	return nil
}

type KVPair struct {
	Hash        string
	Name        string
	LastRun     string
	NextRun     string
	Duration    string
	IsRecurring string
	Params      string
}

type KVStore interface {
	Set(string, []byte) error
	Get(string) (string, error)
	// Fetch() ([]string, error)
	// Remove(string) error
}

type RedisKVStore struct {
	client *redis.Client
}

func GetRedisStore() RedisKVStore {
	rkvs := RedisKVStore{
		client: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
	}

	pong, err := rkvs.client.Ping().Result()
	log.Infof(pong, err)
	return rkvs
}

func (rclient RedisKVStore) Set(key string, value []byte) error {
	return rclient.client.Set(key, value, 0).Err()

}
func (rclient RedisKVStore) Get(key string) (string, error) {
	return rclient.client.Get(key).Result()
}

type ConsulKVStore struct {
	client *api.Client
	config *api.Config
}

func GetConsulStore() ConsulKVStore {
	config := api.DefaultConfig()
	config.Address = Config.KVStoreAddress
	client, err := api.NewClient(config)
	if err != nil {
		panic(err)
	}
	ckvs := ConsulKVStore{
		client: client,
		config: config,
	}
	return ckvs
}

func (client ConsulKVStore) Set(key string, value []byte) error {
	log.Infof("storing value to KV, Key: %s\n", key)
	kv := client.client.KV()
	p := &api.KVPair{Key: key, Value: value}
	_, err := kv.Put(p, nil)
	return err

}
func (client ConsulKVStore) Get(key string) (string, error) {
	kv := client.client.KV()
	// Lookup the pair
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		return "", err
	}
	return string(pair.Value), nil
}
