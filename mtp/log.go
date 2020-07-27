package mtp

import log_ "github.com/puhitaku/mtplvcap/log"

var log *log_.Children

func SetLogger(l *log_.Children) {
	log = l
}
