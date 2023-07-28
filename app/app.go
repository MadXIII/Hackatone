package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/madxiii/hackatone/configs"
	"github.com/madxiii/hackatone/internal/repository"
	"github.com/madxiii/hackatone/internal/repository/postgres"
	"github.com/madxiii/hackatone/internal/repository/redis"
	"github.com/madxiii/hackatone/internal/services"
	"github.com/madxiii/hackatone/internal/transport/http/handler"
)

type App interface {
	Run() error
}

type app struct {
	cfg     configs.Configs
	handler *handler.Handler
	route   *echo.Echo
}

func New() (App, error) {
	cfg, err := configs.New()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := postgres.InitDB(ctx, cfg)
	if err != nil {
		log.Printf("InitDB err: %v\n", err)
		return nil, err
	}
	defer db.Close()

	rdb, err := redis.InitRDB(ctx, cfg.Store.RDB)
	if err != nil {
		log.Printf("InitRDB err: %v\n", err)
		return nil, err
	}
	defer rdb.Close()

	repo := repository.New(db, rdb)
	service := services.New(cfg, repo)
	handler := handler.New(cfg, service)

	return app{
		cfg:     cfg,
		handler: handler,
		route:   echo.New(),
	}, nil
}

func (a app) Run() error {
	a.Routes()

	go a.start()

	return nil
}

func (a app) start() {
	port := fmt.Sprintf(":%s", a.cfg.Server.Address)
	if err := a.route.Start(port); err != nil {
		log.Fatalf("incorrect server shutdown: %v\n", err)
	}
}
