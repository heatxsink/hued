package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/heatxsink/go-hue-web/light_state_factory"
	"github.com/heatxsink/go-hue/groups"
	"github.com/heatxsink/go-hue/portal"
	"net/http"
	"os"
)

var (
	hueUsername string = ""
	hueHostname string = ""
	hostname    string = "127.0.0.1"
	port        int    = 9000
)

type ApiResponse struct {
	Result     bool   `json:"result"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func group_name_presets(name string) int {
	returnValue := -1
	if name == "all" {
		returnValue = 0
	} else if name == "bedroom" {
		returnValue = 1
	} else if name == "living-room" {
		returnValue = 2
	} else if name == "hallway" {
		returnValue = 3
	}
	return returnValue
}

func groupV1(w http.ResponseWriter, req *http.Request) {
	apiResponse := ApiResponse{Result: true, Message: "", StatusCode: http.StatusOK}
	req.ParseForm()
	queryParams := req.Form
	if req.Method == "GET" {
		name, nameExists := queryParams["name"]
		state, stateExists := queryParams["state"]
		if nameExists {
			groupID := group_name_presets(name[0])
			if stateExists {
				gg := groups.New(hueHostname, hueUsername)
				groupState := light_state_factory.GetLightState(state[0])
				gg.SetGroupState(groupID, groupState)
			} else {
				apiResponse.Result = false
				apiResponse.Message = "Invalid state."
				apiResponse.StatusCode = http.StatusUnauthorized
			}
		} else {
			apiResponse.Result = false
			apiResponse.Message = "Invalid id or name."
			apiResponse.StatusCode = http.StatusUnauthorized
		}
	} else {
		apiResponse.Result = false
		apiResponse.Message = "Not an HTTP GET."
		apiResponse.StatusCode = http.StatusForbidden
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(apiResponse.StatusCode)
	jsonData, err := json.Marshal(&apiResponse)
	if err != nil {
		glog.Errorf("Error: %s\n", err.Error())
	}
	w.Write([]byte(jsonData))
}

func statusV1(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusTeapot)
	w.Write([]byte(http.StatusText(http.StatusTeapot)))
}

func phoneV1(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "./www/phone.html")
}

func tabletV1(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "./www/tablet.html")
}

func staticAssets(w http.ResponseWriter, req *http.Request) {
	if len(req.URL.Path) == 1 {
		root(w, req)
	} else if string(req.URL.Path[1:7]) == "static" {
		path := fmt.Sprintf("./www/%s", req.URL.Path[1:])
		http.ServeFile(w, req, path)
	} else {
		root(w, req)
	}
}

func root(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusTeapot)
	w.Write([]byte(http.StatusText(http.StatusTeapot)))
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: hued -key=[string]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.StringVar(&hueUsername, "key", os.Getenv("HUE_USERNAME"), "Philips HUE Hub api key.")
	flag.Parse()
	flag.Usage = usage
}

func main() {
	if hueUsername != "" {
		pp, err := portal.GetPortal()
		if err != nil {
			glog.Errorf("Error: %s\n", err.Error())
		}
		hueHostname = pp[0].InternalIPAddress
		mux := http.NewServeMux()
		mux.HandleFunc("/api/1/group", groupV1)
		mux.HandleFunc("/api/1/status", statusV1)
		mux.HandleFunc("/tablet", tabletV1)
		mux.HandleFunc("/phone", phoneV1)
		mux.HandleFunc("/", staticAssets)
		fullHostname := fmt.Sprintf("%s:%d", hostname, port)
		startMessage := fmt.Sprintf("Starting local hued-web on %s\n", fullHostname)
		fmt.Println(startMessage)
		glog.Infof(startMessage)
		err = http.ListenAndServe(fullHostname, mux)
		if err != nil {
			glog.Errorf("Error: %s\n", err.Error())
		}
	} else {
		usage()
	}
}
