package main

import (
	"flag"

	"github.com/sirupsen/logrus"
)

func init() {
	verbose := flag.Bool("d", false, "enable debug mode")
	flag.Parse()
	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
func main() {
	ctrl, err := NewController()
	if err != nil {
		logrus.Panic(err)
	}
	nxtI, nxtColor := ctrl.GetNext()
	if err = ctrl.SetColor(nxtColor); err != nil {
		logrus.Panic(err)
	}
	if err = ctrl.SetState(nxtI); err != nil {
		logrus.Panic(err)
	}
}
