package database

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/kvtools/valkeyrie/store"
	"gopkg.in/telebot.v3"
	"io/ioutil"
	"strings"

	app "github.com/zt-sv/grafana-annotations-bot/internal/app/grafana-annotations-bot"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/kvtools/boltdb"
	etcd "github.com/kvtools/etcdv2"
	"github.com/kvtools/valkeyrie"
)

const (
	bucket = "grafana-annotations-bot"
)

// DbClient : client for store
type DbClient struct {
	logger         log.Logger
	store          store.Store
	storeKeyPrefix string
}

// NewDB : Create new store client
func NewDB(config app.StorageConfig, logger log.Logger) (*DbClient, error) {
	var err error
	var kvStore store.Store
	ctx := context.Background()

	switch strings.ToLower(config.StoreType) {
	case app.StoreTypeBolt:
		kvStore, err = valkeyrie.NewStore(ctx, boltdb.StoreName, []string{config.BoltdbStoreConfig.Path}, &boltdb.Config{Bucket: bucket})
		if err != nil {
			level.Error(logger).Log("msg", "failed to create bolt store backend", "err", err)
			return nil, err
		}

	case app.StoreTypeEtcd:
		tlsConfig := &tls.Config{}

		if config.EtcdStoreConfig.TLSCert != "" {
			cert, err := tls.LoadX509KeyPair(config.EtcdStoreConfig.TLSCert, config.EtcdStoreConfig.TLSKey)
			if err != nil {
				level.Error(logger).Log("msg", "failed to create etcd store backend, could not load certificates", "err", err)
				return nil, err
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		if config.EtcdStoreConfig.TLSCA != "" {
			caCert, err := ioutil.ReadFile(config.EtcdStoreConfig.TLSCA)
			if err != nil {
				level.Error(logger).Log("msg", "failed to create etcd store backend, could not load ca certificate", "err", err)
				return nil, err
			}

			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}

		tlsConfig.InsecureSkipVerify = config.EtcdStoreConfig.TLSInsecureSkipVerify

		if !config.EtcdStoreConfig.TLSInsecure {
			kvStore, err = valkeyrie.NewStore(ctx, etcd.StoreName, []string{config.EtcdStoreConfig.URL.String()}, &etcd.Config{TLS: tlsConfig})
		} else {
			kvStore, err = valkeyrie.NewStore(ctx, etcd.StoreName, []string{config.EtcdStoreConfig.URL.String()}, nil)
		}

		if err != nil {
			level.Error(logger).Log("msg", "failed to create etcd store backend", "err", err)
			return nil, err
		}

		level.Info(logger).Log("msg", "Create ETCD store")
	default:
		level.Error(logger).Log("msg", "please provide one of the following supported store backends: bolt, etcd")
		return nil, err
	}

	client := &DbClient{
		logger:         logger,
		store:          kvStore,
		storeKeyPrefix: config.StoreKeyPrefix,
	}

	defer kvStore.Close()

	return client, nil
}

// StoreValue : telebot chat and tags list
type StoreValue struct {
	Tags     []string
	ThreadID int
	Chat     *telebot.Chat
}

func (client *DbClient) createStoreKey(chat *telebot.Chat, thread int) string {
	if thread != 0 {
		return fmt.Sprintf("%s/%d-%d", client.storeKeyPrefix, chat.ID, thread)
	}
	return fmt.Sprintf("%s/%d", client.storeKeyPrefix, chat.ID)
}

// AddChatTags : Add telebot chat and subscribed tags to bolt store
func (client *DbClient) AddChatTags(chat *telebot.Chat, thread int, tags []string) error {
	ctx := context.Background()
	storeValue, err := json.Marshal(&StoreValue{
		Tags:     tags,
		ThreadID: thread,
		Chat:     chat,
	})

	if err != nil {
		return err
	}

	err = client.store.Put(
		ctx,
		client.createStoreKey(chat, thread),
		storeValue,
		nil,
	)

	return err
}

// GetChatTags : Get subscribed tags for the chat from bolt store
func (client *DbClient) GetChatTags(chat *telebot.Chat, thread int) ([]string, error) {
	ctx := context.Background()
	pair, err := client.store.Get(ctx, client.createStoreKey(chat, thread), nil)

	if err != nil {
		level.Error(client.logger).Log("msg", "failed to get chat tags", "err", err)

		return nil, err
	}

	value := StoreValue{ThreadID: 0}

	json.Unmarshal(pair.Value, &value)

	return value.Tags, err
}

// ExistChat : Check chat exist into bolt store
func (client *DbClient) ExistChat(chat *telebot.Chat, thread int) (bool, error) {
	ctx := context.Background()
	exist, err := client.store.Exists(ctx, client.createStoreKey(chat, thread), nil)

	if err == store.ErrKeyNotFound {
		return false, nil
	}

	if err != nil {
		level.Error(client.logger).Log("msg", fmt.Sprintf("Check key %s error", client.createStoreKey(chat, thread)), "err", err)
	}

	return exist, err
}

// Remove : Remove chat from bolt store
func (client *DbClient) Remove(chat *telebot.Chat, thread int) error {
	ctx := context.Background()
	return client.store.Delete(ctx, client.createStoreKey(chat, thread))
}

// List : Get chat and tags list from bolt store
func (client *DbClient) List() ([]StoreValue, error) {
	ctx := context.Background()
	pairs, err := client.store.List(ctx, client.storeKeyPrefix, nil)

	if err != nil {
		level.Error(client.logger).Log("msg", fmt.Sprintf("Could not list %s keys", client.storeKeyPrefix), "err", err)
		return nil, err
	}

	var values []StoreValue
	for _, kv := range pairs {
		v := StoreValue{ThreadID: 0}

		if err := json.Unmarshal(kv.Value, &v); err != nil {
			level.Error(client.logger).Log("msg", fmt.Sprintf("Could not unmarshal json value %s", kv.Value), "err", err)
			return nil, err
		}
		values = append(values, v)
	}

	return values, nil
}
