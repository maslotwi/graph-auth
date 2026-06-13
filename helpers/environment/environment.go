package environment

import "os"

var (
	Port          string
	BaseUrl       string
	Neo4jURI      string
	Neo4jUser     string
	Neo4jPassword string
	RedisAddr     string
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

	RedisAddr = os.Getenv("REDIS_ADDR")
	if RedisAddr == "" {
		RedisAddr = "localhost:6379"
	}
}
