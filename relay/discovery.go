package relay

import (
	"context"
	"fmt"
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

type PubSubManager struct {
	ps     *pubsub.PubSub
	topics map[string]*pubsub.Topic
	subs   map[string]*pubsub.Subscription
	mu     sync.Mutex
}

func NewPubSubManager(ctx context.Context, h host.Host) (*PubSubManager, error) {
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, err
	}

	return &PubSubManager{
		ps:     ps,
		topics: make(map[string]*pubsub.Topic),
		subs:   make(map[string]*pubsub.Subscription),
	}, nil
}

func (pm *PubSubManager) JoinTopic(topicName string) (*pubsub.Topic, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if topic, exists := pm.topics[topicName]; exists {
		return topic, nil
	}

	topic, err := pm.ps.Join(topicName)
	if err != nil {
		return nil, err
	}
	pm.topics[topicName] = topic

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}
	pm.subs[topicName] = sub

	fmt.Printf("[PubSub] Subscribed to topic: %s\n", topicName)
	return topic, nil
}

func (pm *PubSubManager) GetPubSub() *pubsub.PubSub {
	return pm.ps
}
