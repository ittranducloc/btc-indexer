package main

import (
	"context"
	btc_indexer "github.com/darkknightbk52/btc-indexer"
	"github.com/darkknightbk52/btc-indexer/client/blockchain"
	"github.com/darkknightbk52/btc-indexer/common/log"
	"github.com/darkknightbk52/btc-indexer/service/indexer"
	"github.com/darkknightbk52/btc-indexer/store"
	"github.com/darkknightbk52/btc-indexer/subscriber"
	"github.com/micro/cli"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/config"
	"github.com/micro/go-micro/config/source/env"
	"github.com/micro/go-micro/registry/consul"
	"go.uber.org/zap"
	"sync"
	"time"
)

const prodFlag = "prod"

func main() {
	prod := false
	opts := []micro.Option{
		micro.RegisterTTL(time.Second * 30),
		micro.RegisterInterval(time.Second * 15),
		micro.Registry(consul.NewRegistry()),
		micro.Flags(
			cli.BoolFlag{
				Name:  prodFlag,
				Usage: "Enable production mode",
			},
		),
		micro.Name("go.micro.srv.btc.indexer"),
		micro.Action(func(ctx *cli.Context) {
			prod = ctx.Bool(prodFlag)
		}),
	}
	microSrv := micro.NewService(opts...)
	microSrv.Init()
	log.Init(prod)

	cfgScanner := config.NewConfig()
	err := cfgScanner.Load(
		//consul.NewSource(),
		env.NewSource(env.WithStrippedPrefix("IDX")),
	)
	if err != nil {
		log.L().Fatal("Failed to load configs:", zap.Error(err))
	}
	cfg := btc_indexer.Config{}
	err = cfgScanner.Scan(&cfg)
	if err != nil {
		log.L().Fatal("Failed to load configs", zap.Error(err))
	}

	err = cfg.Validate()
	if err != nil {
		log.L().Fatal("Invalid config values", zap.Error(err))
	}

	subOpts := []subscriber.Option{
		subscriber.Url(cfg.BlockchainSubscriber.FullNodeUrl),
	}
	if cfg.BlockchainSubscriber.TimeoutInSecond > 0 {
		subOpts = append(subOpts, subscriber.TimeoutDuration(time.Second*time.Duration(cfg.BlockchainSubscriber.TimeoutInSecond)))
	}
	if cfg.BlockchainSubscriber.RetryTimeInSecond > 0 {
		subOpts = append(subOpts, subscriber.RetryDuration(time.Second*time.Duration(cfg.BlockchainSubscriber.RetryTimeInSecond)))
	}
	sub := subscriber.NewSubscriber(subOpts...)

	manager, err := store.NewPostgresManager(cfg.DB.DSN())
	if err != nil {
		log.L().Fatal("Failed to Create Store Manager", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	client, err := blockchain.NewBlockchainClient(ctx, &wg, cfg.BlockchainClient, cfg.Indexer.ChainParams())
	if err != nil {
		log.L().Fatal("Failed to Create Blockchain Client", zap.Error(err))
	}

	indexerSrv := indexer.NewIndexer(cfg.Indexer, sub, manager, client)

	microSrv.Init(
		micro.AfterStart(func() error {
			err = indexerSrv.Listen(ctx, cfg.Indexer.FromBlockHeight)
			if err != context.Canceled {
				return err
			}
			return nil
		}),
		micro.BeforeStop(func() error {
			cancel()
			wg.Wait()
			return nil
		}),
	)

	err = microSrv.Run()
	if err != nil {
		log.L().Fatal("Have errors while BTC Indexer service running", zap.Error(err))
	}
}
