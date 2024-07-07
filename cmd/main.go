package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/isd-sgcu/rpkm67-backend/config"
	"github.com/isd-sgcu/rpkm67-backend/constant"
	"github.com/isd-sgcu/rpkm67-backend/database"
	"github.com/isd-sgcu/rpkm67-backend/internal/cache"
	"github.com/isd-sgcu/rpkm67-backend/internal/group"
	"github.com/isd-sgcu/rpkm67-backend/internal/pin"
	"github.com/isd-sgcu/rpkm67-backend/internal/selection"
	"github.com/isd-sgcu/rpkm67-backend/internal/stamp"
	"github.com/isd-sgcu/rpkm67-backend/internal/user"
	"github.com/isd-sgcu/rpkm67-backend/logger"
	groupProto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/group/v1"
	pinProto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/pin/v1"
	selectionProto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/selection/v1"
	stampProto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/stamp/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	logger := logger.New(conf)

	db, err := database.InitDatabase(&conf.Db, conf.App.IsDevelopment())
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	redis, err := database.InitRedis(&conf.Redis)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to redis: %v", err))
	}

	cacheRepo := cache.NewRepository(redis)

	pinRepo := pin.NewRepository(redis)
	pinUtils := pin.NewUtils()
	pinSvc := pin.NewService(&conf.Pin, pinUtils, pinRepo, logger.Named("pinSvc"))

	stampRepo := stamp.NewRepository(db)
	stampSvc := stamp.NewService(stampRepo, constant.ActivityIdToIdx, logger.Named("stampSvc"))

	userRepo := user.NewRepository(db)
	groupRepo := group.NewRepository(db)
	groupSvc := group.NewService(groupRepo, userRepo, cacheRepo, &conf.Group, logger.Named("groupSvc"))

	selectionRepo := selection.NewRepository(db)
	selectionSvc := selection.NewService(selectionRepo, cacheRepo, logger.Named("selectionSvc"))

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", conf.App.Port))
	if err != nil {
		panic(fmt.Sprintf("Failed to listen: %v", err))
	}

	grpcServer := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())
	pinProto.RegisterPinServiceServer(grpcServer, pinSvc)
	stampProto.RegisterStampServiceServer(grpcServer, stampSvc)
	groupProto.RegisterGroupServiceServer(grpcServer, groupSvc)
	selectionProto.RegisterSelectionServiceServer(grpcServer, selectionSvc)

	reflection.Register(grpcServer)
	go func() {
		logger.Sugar().Infof("RPKM67 Backend starting at port %v", conf.App.Port)

		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatal("Failed to start RPKM67 Backend service", zap.Error(err))
		}
	}()

	wait := gracefulShutdown(context.Background(), 2*time.Second, logger, map[string]operation{
		"server": func(ctx context.Context) error {
			grpcServer.GracefulStop()
			return nil
		},
		"database": func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return nil
			}
			return sqlDB.Close()
		},
		// "cache": func(ctx context.Context) error {
		// 	return cacheDb.Close()
		// },
	})

	<-wait

	grpcServer.GracefulStop()
	logger.Info("Closing the listener")
	listener.Close()
	logger.Info("RPKM67 Backend service has been shutdown gracefully")
}

type operation func(ctx context.Context) error

func gracefulShutdown(ctx context.Context, timeout time.Duration, log *zap.Logger, ops map[string]operation) <-chan struct{} {
	wait := make(chan struct{})
	go func() {
		s := make(chan os.Signal, 1)

		signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		sig := <-s

		log.Named("graceful shutdown").Sugar().
			Infof("got signal \"%v\" shutting down service", sig)

		timeoutFunc := time.AfterFunc(timeout, func() {
			log.Named("graceful shutdown").Sugar().
				Errorf("timeout %v ms has been elapsed, force exit", timeout.Milliseconds())
			os.Exit(0)
		})

		defer timeoutFunc.Stop()

		var wg sync.WaitGroup

		for key, op := range ops {
			wg.Add(1)
			innerOp := op
			innerKey := key
			go func() {
				defer wg.Done()

				log.Named("graceful shutdown").Sugar().
					Infof("cleaning up: %v", innerKey)
				if err := innerOp(ctx); err != nil {
					log.Named("graceful shutdown").Sugar().
						Errorf("%v: clean up failed: %v", innerKey, err.Error())
					return
				}

				log.Named("graceful shutdown").Sugar().
					Infof("%v was shutdown gracefully", innerKey)
			}()
		}

		wg.Wait()
		close(wait)
	}()

	return wait
}
