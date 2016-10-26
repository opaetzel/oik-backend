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
	"github.com/gorilla/handlers"
	"github.com/natefinch/lumberjack"
)

type SMTPConfig struct {
	UserName string
	Password string
	Host     string
	Port     int
	From     string
}

type Config struct {
	UseTLS       bool
	HTTPPort     int
	HTTPSPort    int
	PemFile      string
	KeyFile      string
	DBName       string
	DBUser       string
	DBPassword   string
	ImageStorage string
	StaticFolder string
	AppUrl       string
	LogFile      string
	MailConfig   SMTPConfig
}

var conf Config

func main() {
	tomlFile := flag.String("config", "dev_config.toml", "configuration file")
	flag.Parse()

	if _, err := toml.DecodeFile(*tomlFile, &conf); err != nil {
		fmt.Println("Error trying to read configuration in", *tomlFile)
		fmt.Println(err)
		os.Exit(-1)
	}
	lJack := lumberjack.Logger{
		Filename:   conf.LogFile,
		MaxSize:    10, // megabytes
		MaxBackups: 10,
		MaxAge:     28, //days
	}
	log.SetOutput(&lJack)
	initDB(conf.DBName, conf.DBUser, conf.DBPassword)

	router := NewRouter()
	http.Handle("/", router)
	if conf.UseTLS {
		go func() {
			if err := http.ListenAndServeTLS(":"+strconv.Itoa(conf.HTTPSPort), conf.PemFile, conf.KeyFile, handlers.LoggingHandler(os.Stdout, router)); err != nil {
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
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(conf.HTTPPort), handlers.LoggingHandler(&lJack, router)))
	}
}
