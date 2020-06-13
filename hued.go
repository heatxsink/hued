package main

//go:generate rice embed-go
import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/golang/glog"
	"github.com/gorilla/mux"

	"github.com/heatxsink/go-hue/groups"
	"github.com/heatxsink/go-hue/portal"
	"github.com/heatxsink/hued/factory"
)

var (
	hueUsername  string
	hueHostname  string
	hostname     string
	port         int
	healthy      int32
	templates    map[string]*template.Template
	router       *mux.Router
	buttonState  int
	buttonStates []string
)

type APIResponse struct {
	Result     bool   `json:"result"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: hued -key=[string]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	buttonState = 0
	buttonStates = []string{"deep-sea", "blue", "reading"}
	flag.StringVar(&hueUsername, "key", os.Getenv("HUE_USERNAME"), "Philips HUE Hub api key.")
	flag.StringVar(&hostname, "h", "0.0.0.0", "Hostname of server.")
	flag.IntVar(&port, "p", 9000, "Port number of server.")
	flag.Parse()
	flag.Usage = usage
}

func getHueHubHostname(username string) string {
	pp, err := portal.GetPortal()
	if err != nil {
		glog.Errorf("Error: %s\n", err.Error())
	}
	hn := pp[0].InternalIPAddress
	glog.V(1).Infoln("Hostname: ", hn)
	return hn
}

func loadRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/flic", flicV1).Name("flic").Methods("GET", "POST")
	r.HandleFunc("/phone", phoneV1).Name("phone").Methods("GET")
	r.HandleFunc("/home", homeV1).Name("home").Methods("GET")
	r.HandleFunc("/status", statusV1).Name("status").Methods("GET")
	r.HandleFunc("/api/1/group", groupV1).Name("api_group").Methods("GET")
	r.HandleFunc("/api/1/status", statusV1).Name("api_status").Methods("GET")
	r.HandleFunc("/", rootV1).Name("root").Methods("GET")
	fs := http.FileServer(rice.MustFindBox("www/").HTTPBox())
	sh := http.StripPrefix("/", blackholeHandler(fs))
	r.PathPrefix("/").Handler(sh)
	return r
}

func blackholeHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loadTemplates() (map[string]*template.Template, error) {
	box, err := rice.FindBox("www/")
	if err != nil {
		return nil, err
	}
	templates := make(map[string]*template.Template)
	templateString, err := box.String("templates/home.html")
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New("home.html").Parse(templateString)
	if err != nil {
		return nil, err
	}
	templates["home.html"] = tmpl
	templateString1, err := box.String("templates/phone.html")
	if err != nil {
		return nil, err
	}
	tmpl1, err := template.New("phone.html").Parse(templateString1)
	if err != nil {
		return nil, err
	}
	templates["phone.html"] = tmpl1
	return templates, nil
}

func getContext(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	ctx := make(map[string]interface{})
	return ctx, nil
}

func rootV1(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/home", 302)
}

func phoneV1(w http.ResponseWriter, r *http.Request) {
	tmpl := templates["phone.html"]
	ctx, err := getContext(w, r)
	if err != nil {
		glog.Errorln(err)
	}
	ctx["title"] = "hued"
	err = tmpl.Execute(w, ctx)
	if err != nil {
		glog.Errorf("tmpl.Execute(): %s", err)
	}
}

func homeV1(w http.ResponseWriter, r *http.Request) {
	tmpl := templates["home.html"]
	ctx, err := getContext(w, r)
	if err != nil {
		glog.Errorln(err)
	}
	ctx["title"] = "hued - home"
	err = tmpl.Execute(w, ctx)
	if err != nil {
		glog.Errorf("tmpl.Execute(): %s", err)
	}
}

func groupV1(w http.ResponseWriter, r *http.Request) {
	apiResponse := APIResponse{Result: true, Message: "", StatusCode: http.StatusOK}
	r.ParseForm()
	queryParams := r.Form
	if r.Method == "GET" {
		name, nameExists := queryParams["name"]
		state, stateExists := queryParams["state"]
		if nameExists {
			groupID := factory.GroupNamePresets(name[0])
			if stateExists {
				gg := groups.New(hueHostname, hueUsername)
				groupState := factory.GetLightState(state[0])
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

func flicV1(w http.ResponseWriter, r *http.Request) {
	apiResponse := APIResponse{Result: true, Message: "", StatusCode: http.StatusOK}
	r.ParseForm()
	queryParams := r.Form
	currentLightState := buttonStates[buttonState]
	if r.Method == "GET" {
		apiResponse.Message = currentLightState
	} else if r.Method == "POST" {
		name, nameExists := queryParams["name"]
		if nameExists {
			groupID := factory.GroupNamePresets(name[0])
			gg := groups.New(hueHostname, hueUsername)
			groupState := factory.GetLightState(currentLightState)
			gg.SetGroupState(groupID, groupState)
			if buttonState == (len(buttonStates) - 1) {
				buttonState = 0
			} else {
				buttonState = buttonState + 1
			}
			apiResponse.Message = currentLightState
		} else {
			apiResponse.Result = false
			apiResponse.Message = "Invalid id or name."
			apiResponse.StatusCode = http.StatusUnauthorized
		}
	} else {
		apiResponse.Result = false
		apiResponse.Message = "Not an HTTP POST."
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

func statusV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusTeapot)
	w.Write([]byte(http.StatusText(http.StatusTeapot)))
}

func main() {
	if hueUsername != "" {
		var err error
		hueHostname = getHueHubHostname(hueUsername)
		router := loadRouter()
		templates, err = loadTemplates()
		if err != nil {
			glog.Fatalln(err)
		}
		httpAddress := fmt.Sprintf("%s:%d", hostname, port)
		server := &http.Server{
			Addr:         httpAddress,
			Handler:      router,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second,
		}
		done := make(chan bool)
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		go func() {
			<-quit
			glog.Infoln("Server is shutting down...")
			atomic.StoreInt32(&healthy, 0)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			server.SetKeepAlivesEnabled(false)
			if err := server.Shutdown(ctx); err != nil {
				glog.Fatalln("Could not gracefully shutdown the server: %v\n", err)
			}
			close(done)
		}()
		glog.Infoln("Server is ready to handle requests at", httpAddress)
		atomic.StoreInt32(&healthy, 1)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			glog.Fatalf("Could not listen on %s: %v\n", httpAddress, err)
		}
		<-done
		glog.Infoln("Server stopped")
	} else {
		usage()
	}
}
