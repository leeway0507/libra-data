package db

import (
	"context"
	"libraData"
	"testing"

	"github.com/caarlos0/env/v11"
)

func Test_create_table(t *testing.T) {
	libraData.SetCustomEnv()

	var cfg libraData.EnvConfig
	err := env.Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	conn := connectPG(cfg.DATABASE_URL, &ctx)

	defer conn.Close(ctx)
	t.Run("create book table", func(t *testing.T) {
		err := Create_Book_Table(conn, &ctx)
		if err != nil {
			t.Fatal(err)
		}
	})

}
