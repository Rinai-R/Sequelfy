package main

import (
	"github.com/Rinai-R/Sequelfy/app/component"
	"github.com/sirupsen/logrus"
)

func main() {
	app := component.NewSequelfyApp()
	err := app.Clear()
	if err != nil {
		logrus.Error("Failed to clear history: ", err)
	}
	logrus.Info("History cleared")
}
