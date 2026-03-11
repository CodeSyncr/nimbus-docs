/*
|--------------------------------------------------------------------------
| Queue Configuration
|--------------------------------------------------------------------------
|
| Driver and backend-specific options for the job queue.
| This is a thin layer over environment variables used by
| the core Nimbus queue package (queue.Boot).
|
| Supported drivers:
|   - sync     (in-process, no persistence)
|   - redis    (Redis lists + delayed jobs)
|   - database (SQL-backed queue_jobs table)
|   - sqs      (AWS SQS)
|   - kafka    (Kafka topic / group)
|
*/

package config

// Queue holds queue configuration for the app.
var Queue QueueConfig

type QueueConfig struct {
	Driver       string
	RedisURL     string
	SQSQueueURL  string
	KafkaBrokers string
	KafkaTopic   string
	KafkaGroupID string
}

func loadQueue() {
	Queue = QueueConfig{
		Driver:       env("QUEUE_DRIVER", "sync"),
		RedisURL:     env("REDIS_URL", ""),
		SQSQueueURL:  env("SQS_QUEUE_URL", ""),
		KafkaBrokers: env("KAFKA_BROKERS", ""),
		KafkaTopic:   env("KAFKA_TOPIC", "nimbus-queue"),
		KafkaGroupID: env("KAFKA_GROUP_ID", "nimbus-queue"),
	}
}

