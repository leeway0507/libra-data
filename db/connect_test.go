package db

import (
	"context"
	"libraData"
	"testing"

	"github.com/caarlos0/env/v11"
)

func Test_Connect(t *testing.T) {
	libraData.GetEnvConfig()

	t.Run("test", func(t *testing.T) {
		var cfg libraData.EnvConfig
		err := env.Parse(&cfg)
		if err != nil {
			t.Fatal(err)
		}

		ctx := context.Background()
		conn := connectPGPool(cfg.DATABASE_URL, &ctx)

		defer conn.Close()

	})

}
