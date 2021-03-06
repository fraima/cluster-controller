package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/fraima/cluster-controller/internal/config"
	"github.com/fraima/cluster-controller/internal/controller"
	"github.com/fraima/cluster-controller/internal/kubernetes"
)

var (
	Version = "undefined"
)

func main() {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level.SetLevel(zap.DebugLevel)
	logger, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)

	var (
		configPath, kuberconfigPath string
	)
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&kuberconfigPath, "kuberconfig", "", "path to kuberconfig file")
	flag.Parse()

	if configPath == "" {
		zap.L().Fatal("not found config param")
	}

	if kuberconfigPath == "" {
		zap.L().Fatal("not found kuberconfig param")
	}

	cfg, err := config.Read(configPath)
	if err != nil {
		zap.L().Fatal("read configuration", zap.Error(err))
	}

	zap.L().Debug("configuration", zap.Any("config", cfg), zap.String("kuberconfig", kuberconfigPath), zap.String("version", Version))

	kubeClient, err := kubernetes.NewClient(kuberconfigPath)
	for err != nil {
		zap.L().Fatal("init kubernetes", zap.Error(err))
	}
	defer kubeClient.Close()

	cntrl, err := controller.New(
		kubeClient,
		cfg.Controller,
	)
	if err != nil {
		zap.L().Fatal("init controller", zap.Error(err))
	}

	_ = cntrl

	zap.L().Info("started")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
