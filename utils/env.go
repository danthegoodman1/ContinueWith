package utils

import "os"

var (
	Env = os.Getenv("ENV")

	PGDSN = os.Getenv("PG_DSN")

	ProviderAPIUserExchange = MustEnv("PROVIDER_USER_EXCHANGE_URL")
	ProviderAuthHeader      = MustEnv("PROVIDER_AUTH_HEADER")

	// CRDB by default, which means serializable isolation by default
	IsPostgres = os.Getenv("IS_POSTGRES") == "1"

	// Default 12 hours
	RefreshTokenExpireSeconds = GetEnvOrDefaultInt("REFRESH_TOKEN_EXPIRE_SECONDS", 12*3600)
	// Default 1 hour
	AccessTokenExpireSeconds = GetEnvOrDefaultInt("ACCESS_TOKEN_EXPIRE_SECONDS", 3600)

	AdminKey = MustEnv("ADMIN_KEY")
)
