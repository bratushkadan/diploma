package ydb_adapter

// TODO: read from config
const (
	tableAccounts      = "accounts"
	tableRefreshTokens = "refresh_tokens"

	tableAccountsIndexEmailUnique    = "idx_email_uniq"
	tableRefreshTokensIndexAccountId = "idx_account_id"
)
