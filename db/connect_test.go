package db

import (
	"context"
	"libraData/config"
	"testing"

	"github.com/caarlos0/env/v11"
)

func Test_Connect(t *testing.T) {
	config.GetEnvConfig()

	t.Run("test", func(t *testing.T) {
		var cfg config.EnvConfig
		err := env.Parse(&cfg)
		if err != nil {
			t.Fatal(err)
		}

		ctx := context.Background()
		conn := ConnectPGPool(cfg.DATABASE_URL, &ctx)

		defer conn.Close()

	})

}
