package server

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"racoondev.tk/gitea/racoon/venera/internal/dispatcher"

	"github.com/ccding/go-logging/logging"
	"github.com/gorilla/mux"

	"racoondev.tk/gitea/racoon/venera/internal/provider"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
	"racoondev.tk/gitea/racoon/venera/internal/webui"
)

var log *logging.Logger

func setupAccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("[%s] %s %s", r.Method, r.RemoteAddr, r.URL.Path)

		next.ServeHTTP(w, r)
	})
}

func setupAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Add("WWW-Authenticate", `Basic realm="RtservUserRole"`)
			w.WriteHeader(http.StatusUnauthorized)
		} else if user != utils.Configuration.Http.UserName || password != utils.Configuration.Http.Password {
			time.Sleep(1 * time.Second)
			log.Errorf("Access denied for '%s:%s@%s' to '%s'", user, password, r.RemoteAddr, r.URL.Path)
			w.Header().Add("WWW-Authenticate", `Basic realm="RtservUserRole"`)
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			next.ServeHTTP(w, r)
		}

	})
}

func setupPanicHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				//dispatcher.PanicHandler(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func NewTaskHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	providerId := parts[len(parts)-1]
	provider := provider.All()[providerId]

	if r.Method == "POST" {
		session, err := provider.GetSearchSession(r)
		if err != nil {
			webui.ShowError(w, err)
			return
		}

		dispatcher.NewTask(session, providerId)
		http.Redirect(w, r, "/", 300)
	} else {
		provider.ShowSearchPage(w)
	}
}

// InstanceRouter - creation full HTTP handler, it is useful for tests
func InstanceRouter(logger *logging.Logger) http.Handler {
	log = logger

	router := mux.NewRouter()

	router.HandleFunc("/", webui.MainPageHandler)

	providers := provider.All()
	for id, _ := range providers {
		router.HandleFunc("/task/new/"+id, NewTaskHandler).Methods("GET", "POST")
	}

	router.Use(setupAccessLog)
	router.Use(setupAccess)
	router.Use(setupPanicHandler)

	return router
}

// Run start HTTP server
func Run(logger *logging.Logger) {

	router := InstanceRouter(logger)

	ip := utils.Configuration.Http.Ip
	port := utils.Configuration.Http.Port

	log.Infof("Start HTTP server { endpoint = %s:%d }", ip, port)

	logger.Fatal(http.ListenAndServe(ip+":"+
		strconv.Itoa(port), router))

}
