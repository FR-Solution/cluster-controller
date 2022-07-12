package main

import (
	"flag"

	"go.uber.org/zap"

	"github.com/fraima/cluster-controller/internal/config"
	"github.com/fraima/cluster-controller/internal/controller"
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

	cntrl, err := controller.New(
		cfg.Controller,
	)
	if err != nil {
		zap.L().Fatal("init controller", zap.Error(err))
	}

	_ = cntrl
	// kubeClient, err := kubernetes.NewClient(kuberconfigPath)
	// for err != nil {
	// 	zap.L().Fatal("connect to Kubernetes", zap.Error(err))
	// 	kubeClient, err = kubernetes.NewClient(kuberconfigPath)
	// }
	// zap.L().Info("connected to kubernetes")
}
