package main

import (
	"log"
	"net/http"
	"web-server/routes"
	"web-server/utils"
	_ "web-server/utils/session/providers/memory"
)

func main() {
	utils.Run()
	r := routes.Init()
	log.Fatal(http.ListenAndServe(":8080", r))
}
