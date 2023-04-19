package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/heatxsink/go-hue/groups"
	"github.com/heatxsink/go-hue/portal"
	"github.com/heatxsink/hued/presets"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

//go:embed static
var static embed.FS

//go:embed templates
var pages embed.FS

var (
	Version               string
	BuildDate             string
	Hash                  string
	name                  = "hued"
	debugOption           bool
	logToStdErrOption     bool
	loggingFilenameOption string
	hostnameOption        string
	portOption            int
	hueUsernameOption     string

	hueHostname string
	healthy     int32
	templates   map[string]*template.Template
	buttonState int
)

func initLoggerToStdErr() *zap.SugaredLogger {
	stderrSyncer := zapcore.Lock(os.Stderr)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	core := zapcore.NewCore(encoder, stderrSyncer, zapcore.DebugLevel)
	logger := zap.New(core, zap.AddCaller())
	return logger.Sugar()
}

func initLoggerToFile(filename string) *zap.SugaredLogger {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    50, //mb
		MaxBackups: 10,
		MaxAge:     30, //days
		Compress:   false,
	}
	writerSyncer := zapcore.AddSync(lumberJackLogger)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	core := zapcore.NewCore(encoder, writerSyncer, zapcore.DebugLevel)
	logger := zap.New(core, zap.AddCaller())
	return logger.Sugar()
}

func getLogger() *zap.SugaredLogger {
	if logToStdErrOption {
		return initLoggerToStdErr()
	}
	return initLoggerToFile(loggingFilenameOption)
}

type apiResponse struct {
	Result     bool   `json:"result"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func getHueHubHostname() (string, error) {
	pp, err := portal.GetPortal()
	if err != nil {
		return "", err
	}
	return pp[0].InternalIPAddress, nil
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
	fs := http.FileServer(http.FS(static))
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
	ts := make(map[string]*template.Template)
	homeTemplate, err := template.ParseFS(pages, "templates/home.html")
	if err != nil {
		return nil, err
	}
	ts["home.html"] = homeTemplate

	phoneTemplate, err := template.ParseFS(pages, "templates/phone.html")
	if err != nil {
		return nil, err
	}
	ts["phone.html"] = phoneTemplate
	return ts, nil
}

func rootV1(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
}

func phoneV1(w http.ResponseWriter, r *http.Request) {
	slogger := getLogger()
	tmpl := templates["phone.html"]
	ctx := make(map[string]interface{})
	ctx["title"] = "hued"
	err := tmpl.Execute(w, ctx)
	if err != nil {
		slogger.Errorf("tmpl.Execute(): %s", err)
	}
}

func homeV1(w http.ResponseWriter, r *http.Request) {
	slogger := getLogger()
	tmpl := templates["home.html"]
	ctx := make(map[string]interface{})
	ctx["title"] = "hued - home"
	err := tmpl.Execute(w, ctx)
	if err != nil {
		slogger.Errorf("tmpl.Execute(): %s", err)
	}
}

func groupV1(w http.ResponseWriter, r *http.Request) {
	slogger := getLogger()
	ar := apiResponse{Result: true, Message: "", StatusCode: http.StatusOK}
	r.ParseForm()
	queryParams := r.Form
	if r.Method == "GET" {
		name, nameExists := queryParams["name"]
		state, stateExists := queryParams["state"]
		if nameExists {
			groupID := presets.GroupName(name[0])
			if stateExists {
				gg := groups.New(hueHostname, hueUsernameOption)
				groupState := presets.GetLightState(state[0])
				gg.SetGroupState(groupID, groupState)
			} else {
				ar.Result = false
				ar.Message = "Invalid state."
				ar.StatusCode = http.StatusUnauthorized
			}
		} else {
			ar.Result = false
			ar.Message = "Invalid id or name."
			ar.StatusCode = http.StatusUnauthorized
		}
	} else {
		ar.Result = false
		ar.Message = "Not an HTTP GET."
		ar.StatusCode = http.StatusForbidden
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(ar.StatusCode)
	jsonData, err := json.Marshal(&ar)
	if err != nil {
		slogger.Error(err)
	}
	w.Write([]byte(jsonData))
}

func flicV1(w http.ResponseWriter, r *http.Request) {
	slogger := getLogger()
	ar := apiResponse{Result: true, Message: "", StatusCode: http.StatusOK}
	r.ParseForm()
	queryParams := r.Form
	buttonStates := presets.GetButtonStates()
	currentLightState := buttonStates[buttonState]
	if r.Method == "GET" {
		ar.Message = currentLightState
	} else if r.Method == "POST" {
		name, nameExists := queryParams["name"]
		if nameExists {
			groupID := presets.GroupName(name[0])
			gg := groups.New(hueHostname, hueUsernameOption)
			groupState := presets.GetLightState(currentLightState)
			gg.SetGroupState(groupID, groupState)
			if buttonState == (len(buttonStates) - 1) {
				buttonState = 0
			} else {
				buttonState = buttonState + 1
			}
			ar.Message = currentLightState
		} else {
			ar.Result = false
			ar.Message = "Invalid id or name."
			ar.StatusCode = http.StatusUnauthorized
		}
	} else {
		ar.Result = false
		ar.Message = "Not an HTTP POST."
		ar.StatusCode = http.StatusForbidden
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(ar.StatusCode)
	jsonData, err := json.Marshal(&ar)
	if err != nil {
		slogger.Error(err)
	}
	w.Write([]byte(jsonData))
}

func statusV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusTeapot)
	w.Write([]byte(http.StatusText(http.StatusTeapot)))
}

func main() {
	rootCmd := &cobra.Command{
		Use:   name,
		Short: "",
	}
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version number.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version, BuildDate, Hash)
		},
	}
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "server command.",
		RunE: func(cmd *cobra.Command, args []string) error {
			buttonState = 0
			slogger := getLogger()
			var err error
			hueHostname, err = getHueHubHostname()
			if err != nil {
				slogger.Fatal(err)
			}
			router := loadRouter()
			templates, err = loadTemplates()
			if err != nil {
				slogger.Fatal(err)
			}
			httpAddress := fmt.Sprintf("%s:%d", hostnameOption, portOption)
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
				slogger.Info("Server is shutting down...")
				atomic.StoreInt32(&healthy, 0)
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				server.SetKeepAlivesEnabled(false)
				if err := server.Shutdown(ctx); err != nil {
					slogger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
				}
				close(done)
			}()
			slogger.Infof("Server is ready to handle requests at %s", httpAddress)
			atomic.StoreInt32(&healthy, 1)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slogger.Fatalf("Could not listen on %s: %v\n", httpAddress, err)
			}
			<-done
			slogger.Info("Server stopped")

			return nil
		},
	}
	rootCmd.AddCommand(versionCmd, serverCmd)
	rootCmd.PersistentFlags().BoolVarP(&debugOption, "debug", "d", false, "debug mode")
	rootCmd.PersistentFlags().BoolVarP(&logToStdErrOption, "logtostderr", "e", false, "redirect logging to stderr")
	rootCmd.PersistentFlags().StringVarP(&loggingFilenameOption, "log", "l", fmt.Sprintf("./log/%s.log", name), "log filename")
	rootCmd.PersistentFlags().StringVarP(&hueUsernameOption, "key", "k", "", "hue username / key")
	rootCmd.PersistentFlags().StringVarP(&hostnameOption, "hostname", "o", "0.0.0.0", "web server hostname")
	rootCmd.PersistentFlags().IntVarP(&portOption, "port", "p", 9000, "web server port number")
	rootCmd.MarkFlagRequired("log")
	rootCmd.MarkFlagRequired("port")
	rootCmd.MarkFlagRequired("hostname")
	rootCmd.MarkFlagRequired("key")
	rootCmd.Execute()
}
