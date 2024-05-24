package database

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/kvtools/etcdv2"
	"github.com/kvtools/etcdv3"
	"github.com/kvtools/valkeyrie/store"
	"gopkg.in/telebot.v3"
	"io/ioutil"
	"strings"
	"time"

	app "github.com/zt-sv/grafana-annotations-bot/internal/app/grafana-annotations-bot"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/kvtools/boltdb"
	"github.com/kvtools/valkeyrie"
)

const (
	bucket     = "grafana-annotations-bot"
	opsTimeout = 30 * time.Second
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
	var newStoreConfig valkeyrie.Config
	var newStoreDriver string
	var newStoreEndpoints []string
	ctx, cancel := context.WithTimeout(context.Background(), opsTimeout)

	defer cancel()

	switch strings.ToLower(config.StoreType) {
	case app.StoreTypeBolt:
		newStoreConfig = &boltdb.Config{Bucket: bucket}
		newStoreDriver = boltdb.StoreName
		newStoreEndpoints = []string{config.BoltdbStoreConfig.Path}

	case app.StoreTypeEtcdV2, app.StoreTypeEtcdV3:
		newStoreEndpoints = []string{}

		for _, e := range config.EtcdStoreConfig.Endpoints {
			newStoreEndpoints = append(newStoreEndpoints, e.String())
		}

		tlsConfig := &tls.Config{}
		switch strings.ToLower(config.StoreType) {
		case app.StoreTypeEtcdV2:
			newStoreDriver = etcdv2.StoreName
			newStoreConfig = &etcdv2.Config{TLS: tlsConfig}
		case app.StoreTypeEtcdV3:
			newStoreDriver = etcdv3.StoreName
			newStoreConfig = &etcdv3.Config{TLS: tlsConfig}
		}

		if !config.EtcdStoreConfig.TLSInsecure {
			if config.EtcdStoreConfig.TLSCert != "" {
				cert, err := tls.LoadX509KeyPair(config.EtcdStoreConfig.TLSCert, config.EtcdStoreConfig.TLSKey)
				if err != nil {
					level.Error(logger).Log("msg", "failed to create store backend, could not load certificates", "err", err)
					return nil, err
				}
				tlsConfig.Certificates = []tls.Certificate{cert}
			}

			if config.EtcdStoreConfig.TLSCA != "" {
				caCert, err := ioutil.ReadFile(config.EtcdStoreConfig.TLSCA)
				if err != nil {
					level.Error(logger).Log("msg", "failed to create store backend, could not load ca certificate", "err", err)
					return nil, err
				}

				caCertPool := x509.NewCertPool()
				caCertPool.AppendCertsFromPEM(caCert)
				tlsConfig.RootCAs = caCertPool
			}

			tlsConfig.InsecureSkipVerify = config.EtcdStoreConfig.TLSInsecureSkipVerify
		}

	default:
		level.Error(logger).Log("msg", fmt.Sprintf("Please provide one of the following supported store backends %s, %s. %s", app.StoreTypeBolt, app.StoreTypeEtcdV2, app.StoreTypeEtcdV3))
		return nil, err
	}

	level.Info(logger).Log("msg", fmt.Sprintf("Create %s store backend", newStoreDriver), "endpoints", newStoreEndpoints)

	kvStore, err = valkeyrie.NewStore(ctx, newStoreDriver, newStoreEndpoints, newStoreConfig)

	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprintf("Failed to create %s store backend", newStoreDriver), "err", err)
		return nil, err
	}

	client := &DbClient{
		logger:         logger,
		store:          kvStore,
		storeKeyPrefix: config.StoreKeyPrefix,
	}

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
	ctx, cancel := context.WithTimeout(context.Background(), opsTimeout)

	defer cancel()

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
	ctx, cancel := context.WithTimeout(context.Background(), opsTimeout)
	pair, err := client.store.Get(ctx, client.createStoreKey(chat, thread), nil)

	defer cancel()

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
	ctx, cancel := context.WithTimeout(context.Background(), opsTimeout)
	exist, err := client.store.Exists(ctx, client.createStoreKey(chat, thread), nil)

	defer cancel()

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
	ctx, cancel := context.WithTimeout(context.Background(), opsTimeout)
	defer cancel()
	return client.store.Delete(ctx, client.createStoreKey(chat, thread))
}

// List : Get chat and tags list from bolt store
func (client *DbClient) List() ([]StoreValue, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opsTimeout)
	pairs, err := client.store.List(ctx, client.storeKeyPrefix, nil)

	defer cancel()

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
