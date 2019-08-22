package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var createURL = flag.Bool("create", false, "Create a new url, using key=value parameters or json on stdin")
var parseURL = flag.String("parse", "", "Parse a redirector URL")
var runServer = flag.Bool("serve", false, "Run a redirecting webserver")
var rotateSecret = flag.Bool("rotate", false, "Rotate the secret keys")
var configFile = flag.String("config", "clicktrack-conf.json", "Use this configuration file")
var initConfig = flag.Bool("init", false, "Create a new configuration file")

func main() {
	flag.Parse()

	var c Config
	var err error

	if *initConfig {
		err = c.Init()
		if err != nil {
			log.Fatalf("Failed to initialize configuration: %v\n", err)
		}
		err = c.Save(*configFile)
		if err != nil {
			log.Fatalf("Failed to save configuration: %v\n", err)
		}
		return
	}

	err = c.LoadOrInit(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration file: %v\n", err)
	}

	if *rotateSecret {
		err = c.Rotate()
		if err != nil {
			log.Fatalf("Failed to rotate secret: %v\n", err)
		}
		err = c.Save(*configFile)
		if err != nil {
			log.Fatalf("Failed to save configuration: %v\n", err)
		}
		return
	}

	if *createURL {
		v := map[string]interface{}{}
		if len(flag.Args()) == 0 {
			dec := json.NewDecoder(os.Stdin)
			err = dec.Decode(&v)
			if err != nil {
				log.Fatalf("Failed to parse json: %v\n", err)
			}
		} else {
			for _, arg := range flag.Args() {
				kv := strings.SplitN(arg, "=", 2)
				switch len(kv) {
				case 1:
					v[arg] = true
				case 2:
					v[kv[0]] = kv[1]
				}
			}
		}

		url, err := MakeURL(c, v)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(url)
	}

	if *parseURL != "" {
		u, err := url.Parse(*parseURL)
		if err != nil {
			log.Fatalf("Failed to parse URL: %v\n", err)
		}
		v, err := DecryptURL(c, u)
		if err != nil {
			log.Fatalf("Failed to decrypt URL: %v\n", err)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(v)
	}

	if *runServer {
		s := &http.Server{
			Addr:           c.Listen,
			Handler:        NewServer(c),
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 18,
		}
		fmt.Printf("Listening on %s\n", c.Listen)
		log.Fatalln(s.ListenAndServe())
	}
}
