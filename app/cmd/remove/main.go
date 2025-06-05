package main

import (
	"github.com/Rinai-R/Sequelfy/app/component"
	"github.com/sirupsen/logrus"
)

func main() {
	app := component.NewSequelfyApp()
	err := app.RemoveAll()
	if err != nil {
		logrus.Error("Failed to remove history file: ", err)
	}
	logrus.Println("History file removed successfully")
}
