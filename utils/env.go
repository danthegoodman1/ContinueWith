package utils

import "os"

var (
	Env = os.Getenv("ENV")

	PGDSN = os.Getenv("PG_DSN")

	ProviderAPIUserExchange = MustEnv("PROVIDER_USER_EXCHANGE_URL")
	ProviderAuthHeader      = MustEnv("PROVIDER_AUTH_HEADER")
)
