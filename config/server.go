package config

import (
	"flag"
)

var FlagRunAddr string
var BasicURL string

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&BasicURL, "b", "", "basic shorten URL")
	flag.Parse()
}
