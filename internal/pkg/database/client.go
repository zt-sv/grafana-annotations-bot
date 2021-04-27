package database

import (
	"encoding/json"
	"strconv"

	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/tucnak/telebot"
)

const (
	bucket        = "grafana-annotations-bot"
	chatKeyPrefix = "tgchat_"
)

// DbClient : client for bolt store
type DbClient struct {
	path   string
	logger log.Logger
}

// ClientConfig : Bolt store client configuration
type ClientConfig struct {
	Path   string
	Logger log.Logger
}

// NewDB : Create new bolt store client
func NewDB(config ClientConfig) (*DbClient, error) {
	boltStore, err := boltdb.New([]string{config.Path}, &store.Config{Bucket: bucket})

	if err != nil {
		level.Error(config.Logger).Log("msg", "failed to create bolt store", "err", err)

		return nil, err
	}

	client := &DbClient{
		path:   config.Path,
		logger: config.Logger,
	}

	defer boltStore.Close()

	return client, nil
}

func (client *DbClient) open() (store.Store, error) {
	boltStore, err := boltdb.New([]string{client.path}, &store.Config{Bucket: bucket})

	if err != nil {
		level.Error(client.logger).Log("msg", "failed to open bolt store", "err", err)

		return nil, err
	}

	return boltStore, nil
}

// StoreValue : telebot chat and tags list
type StoreValue struct {
	Tags []string
	Chat *telebot.Chat
}

func (client *DbClient) createStoreKey(chat *telebot.Chat) string {
	return chatKeyPrefix + strconv.FormatInt(chat.ID, 10)
}

// AddChatTags : Add telebot chat and subscribed tags to bolt store
func (client *DbClient) AddChatTags(chat *telebot.Chat, tags []string) error {
	boltStore, err := client.open()

	if err != nil {
		return err
	}

	defer boltStore.Close()

	storeValue, err := json.Marshal(&StoreValue{
		Tags: tags,
		Chat: chat,
	})

	if err != nil {
		return err
	}

	err = boltStore.Put(
		client.createStoreKey(chat),
		storeValue,
		nil,
	)

	return err
}

// GetChatTags : Get subscribed tags for the chat from bolt store
func (client *DbClient) GetChatTags(chat *telebot.Chat) ([]string, error) {
	boltStore, err := client.open()

	if err != nil {
		return nil, err
	}

	pair, err := boltStore.Get(client.createStoreKey(chat))

	if err != nil {
		level.Error(client.logger).Log("msg", "failed to get chat tags", "err", err)

		return nil, err
	}

	var value StoreValue

	json.Unmarshal(pair.Value, &value)

	defer boltStore.Close()

	return value.Tags, err
}

// ExistChat : Check chat exist into bolt store
func (client *DbClient) ExistChat(chat *telebot.Chat) (bool, error) {
	boltStore, err := client.open()

	if err != nil {
		return false, err
	}

	defer boltStore.Close()

	exist, err := boltStore.Exists(client.createStoreKey(chat))

	if err == store.ErrKeyNotFound {
		return false, nil
	}

	return exist, err
}

// Remove : Remove chat from bolt store
func (client *DbClient) Remove(chat *telebot.Chat) error {
	boltStore, err := client.open()

	if err != nil {
		return err
	}

	defer boltStore.Close()

	return boltStore.Delete(client.createStoreKey(chat))
}

// List : Get chat and tags list from bolt store
func (client *DbClient) List() ([]StoreValue, error) {
	boltStore, err := client.open()

	if err != nil {
		return nil, err
	}

	defer boltStore.Close()

	pairs, err := boltStore.List(chatKeyPrefix)

	if err != nil {
		return nil, err
	}

	var values []StoreValue
	for _, kv := range pairs {
		var v StoreValue

		if err := json.Unmarshal(kv.Value, &v); err != nil {
			return nil, err
		}
		values = append(values, v)
	}

	return values, nil
}
