package dispatcher

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/racoon-devel/venera/internal/rater"

	"github.com/ccding/go-logging/logging"
	"github.com/gorilla/mux"

	"github.com/racoon-devel/venera/internal/provider"
	"github.com/racoon-devel/venera/internal/utils"
	"github.com/racoon-devel/venera/internal/webui"
)

var log *logging.Logger

type taskItemHandler func(taskId uint) error

type mainContext struct {
	Providers []string
	Tasks     []TaskInfo
}

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
		session, err := provider.CreateSearchSession(r)
		if err != nil {
			webui.DisplayError(w, err)
			return
		}

		AppendTask(session, providerId)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		webui.DisplayNewTask(w, providerId, &webui.CreateContext{Raters: rater.GetRaters()})
	}
}

func editTaskHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Edit task")
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["task"], 10, 32)
	if err != nil {
		webui.DisplayError(w, err)
		return

	}
	err = taskAction(uint(id), func(task *Task) {
		updated, err := task.WebUpdate(w, r)
		if err != nil {
			webui.DisplayError(w, err)
			return
		}
		if updated {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	})

	if err != nil {
		webui.DisplayError(w, err)
		return
	}
}

func actionTaskHandler(w http.ResponseWriter, r *http.Request) {
	controlTaskHandler(w, r, func(taskId uint) error {
		vars := mux.Vars(r)
		action, ok := vars["action"]
		if !ok {
			return fmt.Errorf("Action not defined")
		}

		var nestedError error
		err := taskAction(taskId, func(task *Task) {
			nestedError = task.Action(action, r.URL.Query())
		})

		if nestedError != nil {
			err = nestedError
		}

		return err
	})
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
		log.Errorf("Task #%d control error: %+v", id, err)
		webui.DisplayError(w, err)
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
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

func exportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		provider := provider.All()["export"]
		session, err := provider.CreateSearchSession(r)
		if err != nil {
			webui.DisplayError(w, err)
			return
		}

		AppendTask(session, provider.ID())
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	var ctx mainContext
	ctx.Providers = provider.GetAvailable()
	ctx.Tasks = Describe()
	webui.DisplayExport(w, &ctx)
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

	router.HandleFunc("/", mainHandler).Methods("GET")

	router.HandleFunc("/results", resultsHandler).Methods("GET")
	router.HandleFunc("/result/{result}", resultHandler).Methods("GET")
	router.HandleFunc("/result/{result}/delete", deleteHandler).Methods("GET")
	router.HandleFunc("/result/{result}/favour", favourHandler).Methods("GET")

	providers := provider.All()
	for id, prov := range providers {
		prov.SetupRouter(router.PathPrefix("/" + id).Subrouter())
		router.HandleFunc("/task/new/"+id, newTaskHandler).Methods("GET", "POST")
	}

	router.HandleFunc("/task/{task}", editTaskHandler).Methods("GET", "POST")
	router.HandleFunc("/task/stop/{task}", stopTaskHandler).Methods("GET")
	router.HandleFunc("/task/delete/{task}", deleteTaskHandler).Methods("GET")
	router.HandleFunc("/task/pause/{task}", suspendTaskHandler).Methods("GET")
	router.HandleFunc("/task/run/{task}", runTaskHandler).Methods("GET")
	router.HandleFunc("/task/{task}/{action}", actionTaskHandler).Methods("GET")
	router.HandleFunc("/export", exportHandler).Methods("GET", "POST")

	router.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", http.FileServer(http.Dir(utils.Configuration.Directories.Content+"/web/"))))
	router.Use(setupAccessLog)
	router.Use(setupAccess)
	router.Use(setupPanicHandler)

	return router
}

// Run start HTTP server
func RunServer(logger *logging.Logger, wg *sync.WaitGroup) {
	router := InstanceRouter(logger)

	ip := utils.Configuration.Http.Ip
	port := utils.Configuration.Http.Port

	log.Infof("Start HTTP server { endpoint = %s:%d }", ip, port)

	dispatcher.httpServer = http.Server{Addr: ip + ":" + strconv.Itoa(port), Handler: router}

	wg.Add(1)

	go func() {
		defer wg.Done()
		log.Fatal(dispatcher.httpServer.ListenAndServe())
	}()

	var ctx context.Context
	ctx, dispatcher.cancelNightTimer = context.WithCancel(context.Background())
	wg.Add(1)
	go checkNightMode(ctx, wg)
}

func Stop() {
	dispatcher.log.Info("HTTP Server shutdowning...")
	dispatcher.httpServer.Shutdown(context.Background())

	dispatcher.cancelNightTimer()

	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	for _, task := range dispatcher.tasks {
		task.Shutdown()
	}
}
