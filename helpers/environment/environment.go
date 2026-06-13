package environment

import "os"

var (
	Port    string
	BaseUrl string
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
}
