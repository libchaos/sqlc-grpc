package main

import (
	"database/sql"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	// database driver
	_ "github.com/jackc/pgx/v4/stdlib"

	"booktest/internal/server"
)

const serviceName = "booktest"

//go:embed proto/apidocs.swagger.json
var openAPISpec []byte

func main() {
	cfg := server.Config{
		ServiceName: serviceName,
	}
	var dbURL string
	var dev bool
	flag.StringVar(&dbURL, "db", "", "The Database connection URL")
	flag.IntVar(&cfg.Port, "port", 5000, "The server port")
	flag.IntVar(&cfg.PrometheusPort, "prometheusPort", 0, "The metrics server port")
	flag.StringVar(&cfg.JaegerAgent, "jaegerAgent", "", "The Jaeger Tracing agent URL")
	flag.StringVar(&cfg.Cert, "cert", "", "The path to the server certificate file in PEM format")
	flag.StringVar(&cfg.Key, "key", "", "The path to the server private key in PEM format")
	flag.BoolVar(&cfg.EnableCors, "cors", false, "Enable CORS middleware")
	flag.BoolVar(&cfg.EnableGrpcUI, "grpcui", false, "Serve gRPC Web UI")
	flag.BoolVar(&dev, "dev", false, "Set logger to development mode")
	flag.Parse()

	log := logger(dev)
	defer log.Sync()

	if _, err := maxprocs.Set(); err != nil {
		log.Error("startup", zap.Error(err))
		os.Exit(1)
	}
	log.Info("startup", zap.Int("GOMAXPROCS", runtime.GOMAXPROCS(0)))

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal("failed to open DB", zap.Error(err))
	}

	srv := server.New(cfg, log, registerServer(log, db), registerHandlers(), openAPISpec)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-done
		log.Warn("signal detected...", zap.Any("signal", sig))
		srv.Shutdown()
	}()
	if err := srv.ListenAndServe(); err != nil {
		if err.Error() != "mux: server closed" {
			log.Fatal(err.Error())
		}
	}
}

func logger(dev bool) *zap.Logger {
	var config zap.Config
	if dev {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]interface{}{
		"service": serviceName,
	}

	log, err := config.Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return log
}
