package tests

import (
	"context"
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestPostgres struct {
	ctx       context.Context
	Container *postgres.PostgresContainer
}

func NewPostgres() *TestPostgres {
	log.Println("Starting Postgres container")
	ctx := context.Background()
	pgContainer, _ := postgres.Run(ctx,
		"postgres:17.3-alpine",
		testcontainers.
			WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	log.Println("Postgres container started")
	return &TestPostgres{Container: pgContainer, ctx: ctx}
}

func (tp *TestPostgres) Cleanup() {
	log.Println("Terminating Postgres container")
	tp.Container.Terminate(tp.ctx)
	log.Println("Postgres container terminated")
}

func (tp *TestPostgres) MustConnStr() string {
	return tp.Container.MustConnectionString(tp.ctx)
}
