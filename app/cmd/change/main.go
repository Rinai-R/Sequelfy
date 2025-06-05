package main

import (
	"fmt"

	"github.com/Rinai-R/Sequelfy/app/component"
	"github.com/sirupsen/logrus"
)

func main() {
	app := component.NewSequelfyApp()
	var input string
	logrus.Println("Please enter the name of the history file you want to change to:")
	fmt.Scanln(&input)
	err := app.Change(input)
	if err != nil {
		logrus.Error(err)
	} else {
		msg, err := app.Load()
		if err != nil {
			logrus.Error(err)
		} else {
			for _, m := range msg {
				fmt.Println(m)
			}
		}
		fmt.Println()
		logrus.Println("History file changed successfully!")
	}
}
