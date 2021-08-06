package main

import (
	_ "encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	//use full path for import statement, otherwise import statements in subdirectories fail for some reason
	"github.com/SJSU-Drone-Cloud/mission/routes"
	//"github.com/QianMason/drone-backend/utils"
	_ "github.com/unrolled/render"
	_ "github.com/urfave/negroni"
)

func main() {
	var port string
	args := os.Args
	if len(args) > 2 {
		log.Fatal("provided", len(args)-1, "number of arguments, expected 1")
	} else if len(args) == 2 {
		port = args[1]
	} else {
		fmt.Println("standard port")
		port = "8080"
	}
	fmt.Println("handling routes")
	r := routes.NewRouter()
	http.Handle("/", r)
	fmt.Println("listening on port:", port)
	http.ListenAndServe(":"+port, nil)
}
