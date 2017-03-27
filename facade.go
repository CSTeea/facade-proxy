package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"stathat.com/c/jconfig"
	"io"
	"os"
	"net/http"
	"io/ioutil"
	"strings"
	"time"
)

var f os.File
var apiAddress string
var startTime,requestTime time.Time
var conf *jconfig.Config

func logInit() {
	log.SetFormatter(&log.JSONFormatter{})
	f, err := os.Create("debug.log")
	if err != nil {
		log.Error("creat log failed", err)
		return
	}
	
	log.SetOutput(f)
	if(conf.GetString("env") == "dev" ) {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	log.Info("Program start")

}

func main() {
	conf = jconfig.LoadConfig("config.json")
	logInit()
	fmt.Printf("Facade proxy %s listening in port %s\n", conf.GetString("version"), conf.GetString("port"))
	apiAddress = conf.GetString("remote")
	http.HandleFunc("/healthcheck", healthCheck)
	http.HandleFunc("/", redirect)
	
    err := http.ListenAndServe(":"+conf.GetString("port"), nil)

    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

func redirect(w http.ResponseWriter, r *http.Request) {
	startTime = time.Now()
	requesturi := r.RequestURI
	if string([]rune(requesturi)[0]) == "/" {
		requesturi = requesturi[1:len(requesturi)]
	}

	fmt.Print( r.Method + ": " + requesturi + "\n" )
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Error("Erronous body in request", err)
	} else {
		log.WithFields(log.Fields{
			"type": r.Method,
			"uri": requesturi,
			"payload": string(body),
			}).Info("request")

		reader := strings.NewReader(string(body))
		req, err := http.NewRequest(r.Method, apiAddress+requesturi, reader)

		if err != nil {
			log.Error("Error creating request", err)
		} else {
			for k, v := range r.Header {
				fmt.Print("Req header: "+k+": "+strings.Join(v, ",")+"\n")
				req.Header.Set(k, strings.Join(v, ","))
			}

			client := &http.Client{}
			requestTime = time.Now()
			resp, err := client.Do(req)
			if err != nil {
				log.Error("Error performing request", err)
			} else {
				defer resp.Body.Close()
    			resp_body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Error("Error reading response", err)
				} else {
					for k, v := range resp.Header {
						for _, vv := range v {
							w.Header().Set(k,vv)
						}
					}

					w.Header().Set("X-Time-Full", time.Since(startTime).String())
					w.Header().Set("X-Time-Request", time.Since(requestTime).String())
					w.Header().Set("X-Time-Overhead", fmt.Sprintf("%F", float64(requestTime.Sub(startTime).Nanoseconds()) / 1000000) + "ms")
					
    				io.WriteString(w, string(resp_body) )
    			}
			}

		}
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "All good here!")
}


