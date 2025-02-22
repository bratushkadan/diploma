package ydb_adapter

// TODO: read from config
const (
	tableAccounts      = "`auth/accounts`"
	tableRefreshTokens = "`auth/refresh_tokens`"

	tableAccountsIndexEmailUnique    = "idx_email_uniq"
	tableRefreshTokensIndexAccountId = "idx_account_id"
)
