package main

import (
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var log = &logrus.Logger{
	Out: os.Stdout,
	Formatter: &prefixed.TextFormatter{
		ForceFormatting: true,
		TimestampFormat: "2006-01-02 15:04:05",
	},
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			log.WithField("prefix", "http").Info(r.Method, r.URL.Path, r.RemoteAddr)
		}()
		next.ServeHTTP(w, r)
	})
}
