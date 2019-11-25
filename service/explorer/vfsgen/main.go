package main

import (
	"log"
	"net/http"
	"os"

	"github.com/shurcooL/vfsgen"
)

func main() {
	assets := http.Dir("../webfiles")
	err := vfsgen.Generate(assets, vfsgen.Options{
		PackageName: "explorer",
		// BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatal(err)
	}

	oldLocation := "./assets_vfsdata.go"
	newLocation := "../assets_vfsdata.go"
	err = os.Rename(oldLocation, newLocation)
	if err != nil {
		log.Fatal(err)
	}
}
