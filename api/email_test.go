package api

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/maslotwi/graph-auth/helpers/environment"
)

func TestSendMagicLinkEmail(t *testing.T) {
	_ = godotenv.Load("../.env")
	environment.LoadEnv()

	to := os.Getenv("TEST_EMAIL")
	if to == "" {
		t.Skip("set TEST_EMAIL to run this test, e.g. TEST_EMAIL=you@example.com go test ./api/ -run TestSendMagicLinkEmail -v")
	}

	err := SendMagicLinkEmail(to, "test-token-abc123")
	if err != nil {
		t.Fatalf("failed to send email: %v", err)
	}
	t.Log("email sent successfully")
}
