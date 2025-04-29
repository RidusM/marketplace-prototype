package main

import (
	"flag"
	"userService/internal/app"
)

var configPath = flag.String("config", "/configs/prod.yaml", "specifying config path")

//var configPath = flag.String("config", "myconfig/dev.yaml", "specifying config path")

func main() {
	flag.Parse()
	app.Run(*configPath)
}
