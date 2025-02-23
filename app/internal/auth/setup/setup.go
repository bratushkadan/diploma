package setup

const (
	EnvKeyYdbEndpoint   = "YDB_ENDPOINT"
	EnvKeyYdbAuthMethod = "YDB_AUTH_METHOD"

	EnvKeyYdbDocApiEndpoint = "YDB_DOC_API_ENDPOINT"

	EnvKeyAwsAccessKeyId     = "AWS_ACCESS_KEY_ID"
	EnvKeyAwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"

	EnvKeySqsQueueUrlEmailConfirmations = "SQS_QUEUE_URL_EMAIL_CONFIRMATIONS"
	EnvKeySqsQueueUrlAccountCreations   = "SQS_QUEUE_URL_ACCOUNT_CREATIONS"

	EnvKeySenderEmail                  = "SENDER_EMAIL"
	EnvKeySenderPassword               = "SENDER_PASSWORD"
	EnvKeyEmailConfirmationApiEndpoint = "EMAIL_CONFIRMATION_API_ENDPOINT"
	EnvKeyEmailConfirmationOrigin      = "EMAIL_CONFIRMATION_ORIGIN"

	EnvKeyAccountIdHashSalt = "APP_ID_ACCOUNT_HASH_SALT"
	EnvKeyTokenIdHashSalt   = "APP_ID_TOKEN_HASH_SALT"

	EnvKeyPasswordHashSalt = "APP_PASSWORD_HASH_SALT"

	EnvKeyAuthTokenPrivateKey     = "APP_AUTH_TOKEN_PRIVATE_KEY"
	EnvKeyAuthTokenPublicKey      = "APP_AUTH_TOKEN_PUBLIC_KEY"
	EnvKeyAuthTokenPrivateKeyPath = "APP_AUTH_TOKEN_PRIVATE_KEY_PATH"
	EnvKeyAuthTokenPublicKeyPath  = "APP_AUTH_TOKEN_PUBLIC_KEY_PATH"
)

const (
	EnvKeyStorePicturesBucket = "PICTURES_BUCKET"
)

// Yandex Cloud Serverless
const (
	EnvKeyYmqTriggerHttpEndpointsEnabled = "YMQ_TRIGGER_HTTP_ENDPOINTS_ENABLED"
)
