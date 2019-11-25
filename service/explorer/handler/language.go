package handler

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Language struct {
	asset http.FileSystem
	Store map[string]map[string]string
}

var Lang *Language

func LanguageInit(asset http.FileSystem) {
	Lang = &Language{
		asset: asset,
		Store: map[string]map[string]string{},
	}
	l := Lang
	v, err := l.asset.Open("/language/")
	if err != nil {
		log.Fatal(err)
	}

	var fi []os.FileInfo
	fi, err = v.Readdir(1)
	for err == nil {
		data := loadAsset(l.asset, "/language/"+fi[0].Name())
		var lang map[string]string
		if err := json.Unmarshal(data, &lang); err != nil {
			log.Fatal(err)
		}

		var name string
		ss := strings.Split(fi[0].Name(), ".")
		if  len(ss) > 1 {
			name = ss[0]
		} else {
			name = fi[0].Name()
		}
		

		l.Store[name] = lang

		fi, err = v.Readdir(1)
	}

}

func loadAsset(asset http.FileSystem, path string) []byte {
	f, err := asset.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	bs, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	return bs
}
