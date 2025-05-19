package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Message struct {
	ID        string
	Topic     string
	Payload   []byte
	Headers   map[string]string
	Timestamp int64
}

type Subscription struct {
	ID      string
	Topic   string
	Group   string
	Message chan *Message
}

type Storage struct {
	mu            sync.RWMutex
	dataDir       string
	topics        map[string][]*Message
	subscriptions map[string]*Subscription
	walFile       *os.File
	encoder       *json.Encoder
}

func NewStorage(dataDir string) (*Storage, error) {
	if err := os.Mkdir(dataDir, 0755); err != nil {
		return nil, err
	}

	walPath := filepath.Join(dataDir, "message.wal")
	walFile, err := os.OpenFile(walPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	s := &Storage{
		dataDir:       dataDir,
		topics:        make(map[string][]*Message),
		subscriptions: make(map[string]*Subscription),
		walFile:       walFile,
		encoder:       json.NewEncoder(walFile),
	}

	return s, nil
}

func (s *Storage) loadFromWAL() error {
	walPath := filepath.Join(s.dataDir, "message.wal")
	file, err := os.Open(walPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var message Message
		if err := decoder.Decode(&message); err != nil {
			return err
		}

		s.mu.Lock()
		s.topics[message.Topic] = append(s.topics[message.Topic], &message)
		s.mu.Unlock()
	}
	return nil
}


