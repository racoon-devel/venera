package dispatcher

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ccding/go-logging/logging"
	"github.com/gorilla/mux"

	"racoondev.tk/gitea/racoon/venera/internal/provider"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
	"racoondev.tk/gitea/racoon/venera/internal/webui"
)

var log *logging.Logger

type taskItemHandler func(taskId uint) error

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

func newTaskHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	providerId := parts[len(parts)-1]
	provider := provider.All()[providerId]

	if r.Method == "POST" {
		session, err := provider.GetSearchSession(log, r)
		if err != nil {
			webui.DisplayError(w, err)
			return
		}

		AppendTask(session, providerId)
		http.Redirect(w, r, "/", 300)
	} else {
		webui.DisplayNewTask(w, providerId)
	}
}

func controlTaskHandler(w http.ResponseWriter, r *http.Request, handler taskItemHandler) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["task"], 10, 32)
	if err != nil {
		webui.DisplayError(w, err)
		return

	}

	err = handler(uint(id))
	if err != nil {
		webui.DisplayError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func stopTaskHandler(w http.ResponseWriter, r *http.Request) {
	controlTaskHandler(w, r, func(taskId uint) error {
		log.Debugf("Stopping task %d", taskId)
		return StopTask(uint(taskId))
	})
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	controlTaskHandler(w, r, func(taskId uint) error {
		log.Debugf("Deleting task %d", taskId)
		return DeleteTask(uint(taskId))
	})
}

func suspendTaskHandler(w http.ResponseWriter, r *http.Request) {
	controlTaskHandler(w, r, func(taskId uint) error {
		log.Debugf("Suspending task %d", taskId)
		return SuspendTask(uint(taskId))
	})
}

func runTaskHandler(w http.ResponseWriter, r *http.Request) {
	controlTaskHandler(w, r, func(taskId uint) error {
		log.Debugf("Running task %d", taskId)
		return RunTask(uint(taskId))
	})
}

type mainContext struct {
	Providers []string
	Tasks     []TaskInfo
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	var ctx mainContext
	ctx.Providers = provider.GetAvailable()
	ctx.Tasks = Describe()
	webui.DisplayMain(w, &ctx)
}

// InstanceRouter - creation full HTTP handler, it is useful for tests
func InstanceRouter(logger *logging.Logger) http.Handler {
	log = logger

	router := mux.NewRouter()

	router.HandleFunc("/", mainHandler)

	providers := provider.All()
	for id, _ := range providers {
		router.HandleFunc("/task/new/"+id, newTaskHandler).Methods("GET", "POST")
	}

	router.HandleFunc("/task/stop/{task}", stopTaskHandler).Methods("GET")
	router.HandleFunc("/task/delete/{task}", deleteTaskHandler).Methods("GET")
	router.HandleFunc("/task/pause/{task}", suspendTaskHandler).Methods("GET")
	router.HandleFunc("/task/run/{task}", runTaskHandler).Methods("GET")

	router.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", http.FileServer(http.Dir(utils.Configuration.Other.Content+"/web/"))))
	router.Use(setupAccessLog)
	router.Use(setupAccess)
	router.Use(setupPanicHandler)

	return router
}

// Run start HTTP server
func RunServer(logger *logging.Logger) {

	router := InstanceRouter(logger)

	ip := utils.Configuration.Http.Ip
	port := utils.Configuration.Http.Port

	log.Infof("Start HTTP server { endpoint = %s:%d }", ip, port)

	logger.Fatal(http.ListenAndServe(ip+":"+
		strconv.Itoa(port), router))

}
