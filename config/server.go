package config

import (
	"flag"
)

var FlagRunAddr string
var BasicUrl string

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&BasicUrl, "b", "", "basic shorten URL")
	flag.Parse()
}
