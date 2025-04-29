package main

import (
	"flag"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/app"
)

//TODO: bring linter
// TODO: ports expose dockerfile

var configPath = flag.String("config", "configs/local.yaml", "specifying config path")

func main() {
	app.Run(*configPath)
}
