package ymq

import "time"

// Yandex Cloud Message Queue Trigger HTTP Request Structure

type EventMetadata struct {
	EventId   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	CreatedAt time.Time `json:"created_at"`
	CloudId   string    `json:"cloud_id"`
	FolderId  string    `json:"folder_id"`
}

type MessageAttributeValue struct {
	DataType    string `json:"data_type"`
	BinaryValue []byte `json:"binary_value"`
	StringValue string `json:"string_value"`
}

type YMQMessageDetails struct {
	QueueId string `json:"queue_id"`
	Message struct {
		MessageId              string
		Md5OfBody              string
		Body                   string
		Attributes             map[string]string
		MessageAttributes      map[string]*MessageAttributeValue
		Md5OfMessageAttributes string
	} `json:"message"`
}

type YMQMessage struct {
	EventMetadata EventMetadata     `json:"event_metadata"`
	Details       YMQMessageDetails `json:"details"`
}

type YMQRequest struct {
	Messages []YMQMessage `json:"messages"`
}

type YMQResponse struct {
	StatusCode int
}
