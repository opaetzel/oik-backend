package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	UseTLS    bool
	HTTPPort  int
	HTTPSPort int
	PemFile   string
	KeyFile   string
}

func main() {
	tomlFile := flag.String("config", "config.toml", "configuration file")
	flag.Parse()

	var conf Config
	if _, err := toml.DecodeFile(*tomlFile, &conf); err != nil {
		fmt.Println("Error trying to read configuration in", *tomlFile)
		fmt.Println(err)
		os.Exit(-1)
	}

	router := NewRouter()
	http.Handle("/", router)
	if conf.UseTLS {
		go func() {
			if err := http.ListenAndServeTLS(":"+strconv.Itoa(conf.HTTPSPort), conf.PemFile, conf.KeyFile, nil); err != nil {
				log.Fatalf("ListenAndServeTLS error: %v", err)
			}
		}()
		redirect := func(w http.ResponseWriter, req *http.Request) {
			log.Println("redirecting... " + req.Host + req.URL.String())
			index := strings.LastIndex(req.Host, ":")
			http.Redirect(w, req,
				"https://"+req.Host[0:index]+req.URL.String(),
				http.StatusMovedPermanently)
		}
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(conf.HTTPPort), http.HandlerFunc(redirect)))
	} else {
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(conf.HTTPPort), nil))
	}
}
