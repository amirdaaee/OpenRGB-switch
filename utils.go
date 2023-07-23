package main

import (
	"github.com/sirupsen/logrus"
)

func ErrorLog(err error, msg string) {
	logrus.WithError(err).Error(msg)
}
