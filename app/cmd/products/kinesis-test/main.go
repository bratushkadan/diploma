package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/bratushkadan/floral/internal/auth/setup"
	"github.com/bratushkadan/floral/pkg/cfg"
	"github.com/bratushkadan/floral/pkg/yds"
)

func GeneratePartitionKey(payload []byte) string {
	hash := sha256.New()
	hash.Write(payload)
	hashSum := hash.Sum(nil)

	return hex.EncodeToString(hashSum)
}

func main() {
	m := cfg.AssertEnv(setup.EnvKeyAwsAccessKeyId, setup.EnvKeyAwsSecretAccessKey)

	ctx := context.Background()

	streamName := "test-stream"
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
