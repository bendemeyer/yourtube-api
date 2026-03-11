package sqldb

import (
	"context"
	"database/sql"
	_ "embed"
	"yourtube/models"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

//go:embed videos.sql
var alterVideosSql string

var db *bun.DB = nil

func InitDb(dsn string) {
	if db != nil {
		return
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	bundb := bun.NewDB(sqldb, pgdialect.New())

	ctx := context.Background()

	tableModels := []interface{}{
		(*models.Channel)(nil),
		(*models.Video)(nil),
		(*models.Playlist)(nil),
		(*models.PlaylistVideo)(nil),
		(*models.Category)(nil),
		(*models.User)(nil),
		(*models.AllowedChannel)(nil),
		(*models.AllowedVideo)(nil),
		(*models.BlockedVideo)(nil),
		(*models.VideoView)(nil),
	}

	bundb.RegisterModel(
		(*models.PlaylistVideo)(nil),
		(*models.AllowedChannel)(nil),
		(*models.AllowedVideo)(nil),
		(*models.BlockedVideo)(nil),
	)

	for _, model := range tableModels {
		_, err := bundb.NewCreateTable().Model(model).IfNotExists().Exec(ctx)
		if err != nil {
			panic(err)
		}
	}

	bundb.ExecContext(ctx, alterVideosSql)

	db = bundb
}

func GetDb() *bun.DB {
	return db
}
