// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2021: Nordix Foundation
//   %%
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//   ========================LICENSE_END===================================
//

package kafkaclient

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

//go:generate mockery --name KafkaFactory
type KafkaFactory interface {
	NewKafkaClient(topicID string) (KafkaClient, error)
}

type KafkaFactoryImpl struct {
	BootstrapServer string
}

func (kf *KafkaFactoryImpl) NewKafkaClient(topicID string) (KafkaClient, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kf.BootstrapServer,
		"group.id":          "dmaap-mediator-producer",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	consumer.Commit()
	err = consumer.Subscribe(topicID, nil)
	if err != nil {
		return nil, err
	}
	return &KafkaClientImpl{consumer: consumer}, nil
}

//go:generate mockery --name KafkaClient
type KafkaClient interface {
	ReadMessage() ([]byte, error)
}
type KafkaClientImpl struct {
	consumer *kafka.Consumer
}

func (kc *KafkaClientImpl) ReadMessage() ([]byte, error) {
	msg, err := kc.consumer.ReadMessage(time.Second)
	if err != nil {
		return nil, err
	}
	return msg.Value, nil
}
