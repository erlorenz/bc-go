package integration

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {

	godotenv.Load("../.env")

	os.Exit(m.Run())
}
