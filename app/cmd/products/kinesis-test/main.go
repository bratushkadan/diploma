package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/bratushkadan/floral/internal/auth/setup"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/yds"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
)

func GeneratePartitionKey(payload []byte) string {
	hash := sha256.New()
	hash.Write(payload)
	hashSum := hash.Sum(nil)

	return hex.EncodeToString(hashSum)
}

func main() {

	// for range 5 {
	// 	amazonKinesisApiPublish()
	// }

	// return

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
		kgo.ConsumerGroup("test-topic-consumer-1"),
		kgo.ConsumeTopics("/ru-central1/b1ge5gt845ec0q6sflsv/etnqf4dnapbeccm06epl/test-topic"),
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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

func amazonKinesisApiPublish() {
	m := cfg.AssertEnv(setup.EnvKeyAwsAccessKeyId, setup.EnvKeyAwsSecretAccessKey)

	ctx := context.Background()

	streamName := "test-topic"
	endpoint := "https://yds.serverless.yandexcloud.net"
	endpointPath := "/ru-central1/b1ge5gt845ec0q6sflsv/etnqf4dnapbeccm06epl/" + streamName
	kinesisClient, err := yds.New(ctx, m[setup.EnvKeyAwsAccessKeyId], m[setup.EnvKeyAwsSecretAccessKey], endpoint)
	if err != nil {
		log.Fatal(err)
	}

	data := []byte(`{"message": "you're cool!"}`)

	_, err = kinesisClient.PutRecords(ctx, &kinesis.PutRecordsInput{
		Records: []types.PutRecordsRequestEntry{
			{
				Data:         []byte(data),
				PartitionKey: aws.String(GeneratePartitionKey(data)),
			},
		},
		StreamName: aws.String(endpointPath),
	})
	if err != nil {
		log.Fatalf("failed to put kinesis records: %v", err)
	}
	log.Print("Succesfully put kinesis records to a stream!")
}

func amazonKinesisApiPublishAndReadFromShards() {
	m := cfg.AssertEnv(setup.EnvKeyAwsAccessKeyId, setup.EnvKeyAwsSecretAccessKey)

	ctx := context.Background()

	streamName := "test-stream"
	endpoint := "https://yds.serverless.yandexcloud.net"
	endpointPath := "/ru-central1/b1ge5gt845ec0q6sflsv/etnqf4dnapbeccm06epl/" + streamName
	kinesisClient, err := yds.New(ctx, m[setup.EnvKeyAwsAccessKeyId], m[setup.EnvKeyAwsSecretAccessKey], endpoint)
	if err != nil {
		log.Fatal(err)
	}

	amazonKinesisApiPublish()

	describeStreamOut, err := kinesisClient.DescribeStream(ctx, &kinesis.DescribeStreamInput{
		StreamName: aws.String(endpointPath),
	})
	if err != nil {
		log.Fatalf("failed to describe stream: %v", err)
	}

	var shards []*string
	for _, v := range describeStreamOut.StreamDescription.Shards {
		shards = append(shards, v.ShardId)
	}

	// Yandex Data Streams не поддерживает (https://yandex.cloud/ru/docs/data-streams/kinesisapi/api-ref)
	// методы API RegisterStreamConsumer и ListStreamConsumers, поэтому для того, чтобы имитировать поведение Consumer,
	// необходимо разработать механизм назначения отдельных потребителей сегментам (Partitions/Kinesis API Shard) и
	// необходимо самостоятельно вести наблюдение за отступом (sequence number)
	// Это большое ограничение и неудобство.

	for _, shardId := range shards {
		log.Printf(`iterate over shardId "%s"`, *shardId)
		shardIteratorOut, err := kinesisClient.GetShardIterator(ctx, &kinesis.GetShardIteratorInput{
			StreamName: aws.String(endpointPath),
			ShardId:    shardId,
			// Option 1. ShardIteratorTypeAtSequenceNumber if you want to read all available records starting at a certain position.
			// ShardIteratorType:      types.ShardIteratorTypeAtSequenceNumber,
			// StartingSequenceNumber: aws.String("0"),
			// Option 2. ShardIteratorTypeAfterSequenceNumber  if you want to read all available records starting after a certain position.
			// ShardIteratorType:      types.ShardIteratorTypeAfterSequenceNumber,
			// StartingSequenceNumber: aws.String("0"),
			// Option 3. ShardIteratorTypeLatest  if you want to read all available records starting after a certain position.
			// ShardIteratorType: types.ShardIteratorTypeLatest,
			// Option 4. ShardIteratorTypeTrimHorizon if you want to read all available records starting from the oldest record.
			ShardIteratorType: types.ShardIteratorTypeTrimHorizon,
		})
		if err != nil {
			log.Fatalf("failed to prepare kinesis shard iterator: %v", err)
		}

		log.Printf(`shard iterator "%s"`, *shardIteratorOut.ShardIterator)

		out, err := kinesisClient.GetRecords(ctx, &kinesis.GetRecordsInput{
			ShardIterator: shardIteratorOut.ShardIterator,
		})
		if err != nil {
			log.Fatalf("failed to get records from kinesis shard iterator: %v", err)
		}

		for _, r := range out.Records {
			log.Printf(`record: "%s"`, r.Data)
		}

		// stream is closed if it's null !
		if out.NextShardIterator != nil {
			log.Printf("next shard iterator %s", *out.NextShardIterator)
		}
	}

	log.Print("Successfully read records from the kinesis stream")

}
