package mtp

import "github.com/sirupsen/logrus"

var log *logrus.Logger

func SetLogger(l *logrus.Logger) {
	log = l
}
