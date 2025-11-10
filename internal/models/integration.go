package models

type Integration struct {
	IntegrationId   int64
	Name            string
	IntegrationType int
	HashedApiKey    *string
}

var IntegrationTypes = map[int]string{
	0: "openai",
	1: "anthropic",
	2: "google",
	3: "ollama",
}
