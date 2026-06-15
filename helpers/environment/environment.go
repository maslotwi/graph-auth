package environment

import "os"

var (
	Port          string
	BaseUrl       string
	FrontendURL   string
	IssuerURL     string
	Neo4jURI      string
	Neo4jUser     string
	Neo4jPassword string
	RedisAddr     string

	SMTPHost          string
	SMTPPort          string
	SMTPServerName    string
	SMTPUser          string
	SMTPPass          string
	SMTPFrom          string
	RSAPrivateKeyPath string
)

func LoadEnv() {
	BaseUrl = os.Getenv("BASE_URL")
	if BaseUrl == "" {
		BaseUrl = "127.0.0.1:8080"
	}

	Port = os.Getenv("PORT")
	if Port == "" {
		Port = "8080"
	}

	IssuerURL = os.Getenv("ISSUER_URL")
	if IssuerURL == "" {
		IssuerURL = "http://localhost:8080"
	}

	Neo4jURI = os.Getenv("NEO4J_URI")
	if Neo4jURI == "" {
		Neo4jURI = "bolt://localhost:7687"
	}

	Neo4jUser = os.Getenv("NEO4J_USER")
	if Neo4jUser == "" {
		Neo4jUser = "neo4j"
	}

	Neo4jPassword = os.Getenv("NEO4J_PASSWORD")
	if Neo4jPassword == "" {
		Neo4jPassword = "graphauth"
	}

	RedisAddr = os.Getenv("REDIS_URL")
	if RedisAddr == "" {
		RedisAddr = "redis://localhost:6379/0"
	}

	FrontendURL = os.Getenv("FRONTEND_URL")
	if FrontendURL == "" {
		FrontendURL = "http://localhost:5173"
	}

	SMTPHost = os.Getenv("SMTP_HOST")
	if SMTPHost == "" {
		SMTPHost = "localhost"
	}

	SMTPPort = os.Getenv("SMTP_PORT")
	if SMTPPort == "" {
		SMTPPort = "1025"
	}

	SMTPServerName = os.Getenv("SMTP_SERVER_NAME")
	if SMTPServerName == "" {
		SMTPServerName = SMTPHost
	}

	SMTPUser = os.Getenv("SMTP_USER")
	SMTPPass = os.Getenv("SMTP_PASS")

	SMTPFrom = os.Getenv("SMTP_FROM")
	if SMTPFrom == "" {
		SMTPFrom = "noreply@graph-auth.local"
	}

	RSAPrivateKeyPath = os.Getenv("RSA_PRIVATE_KEY_PATH")
	if RSAPrivateKeyPath == "" {
		RSAPrivateKeyPath = "./jwt_rsa_key.pem"
	}
}
