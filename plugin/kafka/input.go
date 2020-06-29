package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/discretemind/glink/stream"
)

type kafkaPlugin struct {
	cfg     *sarama.Config
	brokers []string
}

func Config(cfg *sarama.Config, brokers []string) *kafkaPlugin {
	return &kafkaPlugin{
		cfg:     cfg,
		brokers: brokers,
	}
}

func (k *kafkaPlugin) InputGroup(ctx context.Context, groupId string, topics ...string) (result *stream.DataStream) {
	//result = stream.Input()

	cg, err := sarama.NewConsumerGroup(k.brokers, groupId, k.cfg)
	if err != nil {
		//result.Error(err)
		return
	}

	if err := cg.Consume(ctx, topics, &inputHandler{}); err != nil {
		//result.Error(err)
	}
	return
}

type inputHandler struct {
	handler stream.FilterHandler
}

func (h *inputHandler) Setup(session sarama.ConsumerGroupSession) error {

	//session.
	//session.
}

func (h *inputHandler) Cleanup(session sarama.ConsumerGroupSession) error {

}

func (h *inputHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {
		h.handler(stream.Event{})
	}
}
