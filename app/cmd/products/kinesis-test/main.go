package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/bratushkadan/floral/internal/auth/setup"
	"github.com/bratushkadan/floral/pkg/cfg"
	ydbpkg "github.com/bratushkadan/floral/pkg/ydb"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicreader"
)

func consumeYdbClient(ctx context.Context, reader *topicreader.Reader) {
	for {
		log.Print("reading message batch")
		batch, err := reader.ReadMessagesBatch(ctx)
		if err != nil {
			log.Fatalf("failed to read message batch: %v", err)
		}
		log.Print("read message batch")

		for _, msg := range batch.Messages {
			content, _ := io.ReadAll(msg)
			log.Print(string(content))
			_ = reader.Commit(msg.Context(), msg)
		}
	}
}

func newYdbTopicReader(ctx context.Context) *topicreader.Reader {
	endpoint := cfg.MustEnv(setup.EnvKeyYdbEndpoint)

	log.Print("setup ydb")
	db, err := ydb.Open(ctx, endpoint, ydbpkg.GetYdbAuthOpts(ydbpkg.YdbAuthMethodEnviron)...)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("set up ydb")

	consumerAndGroupId := "test-topic-consumer-1"

	reader, err := db.Topic().StartReader(consumerAndGroupId, topicoptions.ReadTopic("test-topic"))
	if err != nil {
		log.Fatalf("failed to read topic: %v", err)
	}

	return reader
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// topicreader := newYdbTopicReader(context.Background())
	// consumeYdbClient(context.Background(), topicreader)

	// m := plain.Mechanism{
	// 	Username: "",
	// 	// key id aje8r19oomdd4j3rbu1q,
	// 	// api key is below,
	// 	Password: "",
	// }

	// d := kafka.Dialer{
	// 	SASLMechanism: m,
	// 	TLS: &tls.Config{
	// 		InsecureSkipVerify: true,
	// 	},
	// }

	// conn, err := d.Dial("tcp", "ydb-03.serverless.yandexcloud.net:9093")
	// if err != nil {
	// 	log.Fatalf("Failed to ping Kafka broker: %v", err)
	// }
	// defer conn.Close()

	// reader := kafka.NewReader(kafka.ReaderConfig{
	// 	Brokers:           []string{"ydb-03.serverless.yandexcloud.net:9093"},
	// 	Dialer:            &d,
	// 	Topic:             "test-topic",
	// 	SessionTimeout:    10 * time.Second,
	// 	HeartbeatInterval: 3 * time.Second,
	// 	CommitInterval:    0,
	// 	GroupID:           "test-topic-consumer-1",
	// })
	// defer reader.Close()

	// for {
	// 	log.Print("read message")
	// 	msg, err := reader.ReadMessage(context.Background())
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	if err := reader.CommitMessages(context.Background(), msg); err != nil {
	// 		log.Fatalf("failed to commit messages: %v", err)
	// 	}
	// 	log.Printf("value: %s", string(msg.Value))
	// }

	auth := plain.Auth{
		User: "@/ru-central1/b1ge5gt845ec0q6sflsv/etnqf4dnapbeccm06epl",
		// key id aje8r19oomdd4j3rbu1q,
		// api key is below,
		Pass: "",
	}

	// One client can both produce and consume!
	// Consuming can either be direct (no consumer group), or through a group. Below, we use a group.
	cl, err := kgo.NewClient(
		kgo.SeedBrokers("ydb-03.serverless.yandexcloud.net:9093"),
		kgo.ConsumeTopics("test-topic"),
		kgo.ConsumerGroup("test-topic-consumer-1"),
		kgo.DialTLS(),
		kgo.SASL(auth.AsMechanism()),
		kgo.DisableAutoCommit(),
		// For YDB Topics Kafka API
		kgo.Balancers(kgo.RoundRobinBalancer()),
		kgo.SessionTimeout(5*time.Second),
		kgo.HeartbeatInterval(1500*time.Millisecond),
	)
	if err != nil {
		log.Fatalf("failed to setup kgo client: %v", err)
	}
	defer cl.Close()

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := cl.Ping(ctx); err != nil {
		log.Fatal(fmt.Errorf("failed to ping brokers: %w", err))
	}
	ctx = context.Background()

	// produce START

	// record := &kgo.Record{Topic: "test-topic", Value: []byte(fmt.Sprintf(`{"name": "Dan", "time": "%s"}`, time.Now().Format(time.RFC3339Nano)))}
	// // This is **Asynchronous** produce! For synchronous produce use cl.ProduceSync.
	// log.Print("trying to produce...")
	// innerCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	// cl.Produce(innerCtx, record, func(_ *kgo.Record, err error) {
	// 	defer cancel()
	// 	if err != nil {
	// 		log.Printf("record had a produce error: %v", err)
	// 	} else {
	// 		log.Printf("produced a Product record: %v", record)
	// 	}
	// })

	// <-innerCtx.Done()

	// produce END

	for {
		log.Print("fetching")

		fetches := cl.PollFetches(ctx)
		if fetches.IsClientClosed() {
			break
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			for _, err := range errs {
				log.Printf("err: %+v, %T", err.Err, err.Err)
			}
			log.Fatal(fmt.Sprint(errs))
		}

		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			for _, record := range p.Records {
				log.Printf("topic %s partition %d offset %d \n", p.Topic, p.Partition, record.Offset)
				// if err := json.NewDecoder(bytes.NewReader(record.Value)).Decode(&product); err != nil {
				// 	fmt.Printf("failed to unmarshal message from topic to a product: %v\n", err)
				// 	continue
				// }

				log.Printf("Read product message from a Kafka topic: %+v", record.Value)
				// fmt.Printf("Read product message from a Kafka topic: %+v\n", product)
				// products = append(products, product)
			}
		})

		log.Print("commit topic offsets")
		err = cl.CommitUncommittedOffsets(ctx)
		if err != nil {
			log.Printf("failed to commit offsets: %v", err)
		}
	}
}
