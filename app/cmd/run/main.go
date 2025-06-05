package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Rinai-R/Sequelfy/app/component"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancle := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	logrus.Info("App is starting...")
	app := component.NewSequelfyApp()
	history, err := app.Load()
	if err != nil {
		logrus.Error(err)
		return
	} else {
		for _, m := range history {
			fmt.Println(m)
		}
	}
	go func() {
		<-sigChan
		logrus.Info("App is stopping...")
		cancle()
		os.Exit(0)
	}()
	// 运行程序
	app.Run(ctx, history, sigChan)
}
