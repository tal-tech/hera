package kafkaconsumer

import (
	"runtime/debug"
	"strings"
	"sync"
	"time"

	logger "github.com/tal-tech/loggerX"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/spf13/cast"
)

const MIN_CONSUMER_COUNT = 3
const RETRY_CONSUMER_NUM = 3
const RETRY_PRODUCER_NUM = 3

var produce sarama.SyncProducer

var mu sync.Mutex

type IKafka struct {
	quiter   chan struct{}
	callback IKafkaCallback
}

type IBroker struct {
	Topic     string
	FailTopic string
	KafkaHost []string
}
type IConsumer struct {
	Cnt     int
	GroupId string
}

var panicConsume = make(chan int)

type IKafkaCallback func(broker IBroker, partition int32, offset int64, Key string, Value []byte, ext interface{}) bool

func NewIKafka(callback IKafkaCallback) (this *IKafka) {

	this = new(IKafka)
	this.callback = callback
	this.quiter = make(chan struct{}, 0)

	return
}
func (this *IKafka) Run(brokers []IBroker, consumer *IConsumer) {
	tag := "kafkaconsumer.IKafka.Run"
	if brokers == nil || len(brokers) == 0 {
		logger.E(tag, "No brokers found")
	}
	if consumer.Cnt == 0 {

		consumer.Cnt = MIN_CONSUMER_COUNT
		logger.W("NewIKafka", "No consumer count found,use default %d", consumer.Cnt)
	}

	for _, broker := range brokers {
		for i := 0; i < consumer.Cnt; i++ {
			go this.consume(broker, consumer.GroupId)
		}
		go this.watchConsumer(broker, consumer.GroupId)
	}

}

func (this *IKafka) Close() {
	logger.W("IKafaDao", "Close consumer")
	close(this.quiter)
	time.Sleep(time.Second)
}

func (this *IKafka) watchConsumer(broker IBroker, groupId string) {
	for {
		select {
		case <-panicConsume:
			go this.consume(broker, groupId)
		}
		time.Sleep(time.Second)
	}
}
func initSarama(broker IBroker, cursor string) (consumer *cluster.Consumer) {
	cfg := cluster.NewConfig()
	cfg.Config.ClientID = "kafkaWoker"
	cfg.Config.Consumer.MaxWaitTime = 500 * time.Millisecond
	cfg.Config.Consumer.MaxProcessingTime = 300 * time.Millisecond
	cfg.Config.Consumer.Offsets.CommitInterval = 350 * time.Millisecond
	cfg.Config.Consumer.Offsets.Initial = sarama.OffsetNewest
	cfg.Config.Consumer.Offsets.Retention = time.Hour * 24 * 15
	cfg.Config.Consumer.Return.Errors = true
	cfg.Group.Return.Notifications = true
	//cfg.Version = sarama.V0_10_2_0
	consumer, err := cluster.NewConsumer(broker.KafkaHost, cursor, strings.Split(broker.Topic, ","), cfg)
	if err != nil {
		logger.F("Consume", err)
	}

	if broker.FailTopic != "" {

		if _, err := broker.newProducer(); err != nil {
			logger.F("Producer", err)
		}
	}

	return
}

func (this *IKafka) consume(broker IBroker, cursor string) {
	defer recovery()
	logger.I("Consume", "Start consume from broker %v", broker.KafkaHost)
	//初始化消费队列
	var consumer *cluster.Consumer
	consumer = initSarama(broker, cursor)
	if consumer != nil {
		defer consumer.Close()
	}

	go func(c *cluster.Consumer) {
		for notification := range c.Notifications() {
			logger.D("ConsumeNotifaction", "Rebanlance %+v", notification)
		}
	}(consumer)
	go func(c *cluster.Consumer) {
		for err := range c.Errors() {
			logger.E("Consume", err)
		}
	}(consumer)
	message := consumer.Messages()
	for {
		select {
		case <-this.quiter:
			logger.W("Consume", "IKAFKA_RECV_QUIT")
			return
		case event, ok := <-message:
			if !ok {
				continue
			}
			count := 0
			for {
				ret := this.callback(broker, event.Partition, event.Offset, string(event.Key), event.Value, nil)
				if ret == true {
					logger.D("Consume", "return_true:KEY:%s,OFFSET:%d,PARTITION:%d", string(event.Key), event.Offset, event.Partition)
					consumer.MarkOffset(event, "")
					break
				} else {
					logger.E("Consume", "return_false:KEY:%s,VAL:%s,OFFSET:%d,PARTITION:%d", string(event.Key), string(event.Value), event.Offset, event.Partition)

					if count >= RETRY_CONSUMER_NUM {
						if broker.FailTopic != "" {
							for i := 0; i < RETRY_PRODUCER_NUM; i++ {
								if err := broker.sendByHashPartition(broker.FailTopic, event.Value, event.Value); err != nil {
									logger.E("Consume", "SendToFilerTopic:%s %s %v", broker.FailTopic, string(event.Value), err)
									continue
								}
								break
							}
							consumer.MarkOffset(event, "")
							break
						}
					}
					count++
				}

			}
		}
	}
}
func (this *IBroker) newProducer() (sarama.SyncProducer, error) {

	var err error
	if produce == nil {
		mu.Lock()
		defer mu.Unlock()
		cfg := sarama.NewConfig()
		cfg.Producer.Partitioner = sarama.NewHashPartitioner
		cfg.Producer.RequiredAcks = sarama.WaitForAll
		cfg.Producer.Return.Successes = true

		produce, err = sarama.NewSyncProducer(this.KafkaHost, cfg)
	}
	return produce, err

}
func (this *IBroker) sendByHashPartition(topic string, data []byte, key []byte) error {

	p, err := this.newProducer()
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{Topic: topic, Key: sarama.ByteEncoder(key), Value: sarama.ByteEncoder(data)}
	if partition, offset, err := p.SendMessage(msg); err != nil {
		return err
	} else {
		logger.D("SendByHashPartition", "[KAFKA_OUT]partition:%d,offset:%d,topic:%s,data:%s", partition, offset, topic, string(data))
	}
	return nil
}

func recovery() {
	if rec := recover(); rec != nil {
		if err, ok := rec.(error); ok {
			logger.E("IKafkaPanicRecover", "Unhandled error: %v\n stack:%v", err.Error(), cast.ToString(debug.Stack()))
		} else {
			logger.E("IKafkaPanicRecover", "Panic: %v\n stack:%v", rec, cast.ToString(debug.Stack()))
		}
		panicConsume <- 1
	}
}
