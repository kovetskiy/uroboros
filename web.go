package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/kovetskiy/lorg"
	"github.com/seletskiy/hierr"
)

const (
	prefixAPI    = "/api/v1/"
	prefixBadges = "/badges/"
)

type WebServer struct {
	mux       *http.ServeMux
	address   string
	logger    *lorg.Log
	resources *resources
	requests  int64
}

func NewWebServer(
	logger *lorg.Log,
	resources *resources,
) *WebServer {
	server := &WebServer{
		mux:       http.NewServeMux(),
		logger:    logger,
		resources: resources,
	}

	server.mux.HandleFunc(prefixAPI, server.Handle)
	server.mux.HandleFunc(
		prefixBadges,
		http.FileServer(http.Dir(".")).ServeHTTP,
	)

	return server
}

func (server *WebServer) Serve(address string) error {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return hierr.Errorf(
			err, "can't resolve '%s'", address,
		)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't listen '%s'", addr,
		)
	}

	server.logger.Infof("listening at %s", address)

	return http.Serve(listener, server.mux)
}

func (server *WebServer) Handle(
	writer http.ResponseWriter, request *http.Request,
) {
	var (
		number = atomic.AddInt64(&server.requests, 1)
		logger = getLogger("request#%d", number)
	)

	logger.Infof(
		"-> %s %s",
		request.Method, request.URL,
	)

	status, response := server.route(
		logger, request,
	)

	logger.Infof(
		"<- %d %s",
		status, http.StatusText(status),
	)

	writer.WriteHeader(status)

	if response != nil {
		if err, ok := response.(error); ok {
			response = map[string]interface{}{"error": err}
		}

		err := json.NewEncoder(writer).Encode(response)
		if err != nil {
			logger.Error(err)
		}
	}
}

func (server *WebServer) route(
	logger *lorg.Log,
	request *http.Request,
) (status int, response interface{}) {
	requestURL := "/" + strings.TrimPrefix(request.URL.Path, prefixAPI)

	switch {
	case requestURL == "/tasks/":
		switch request.Method {
		case "POST":
			return server.handleNewTask(logger, request)

		case "GET":
			return server.handleListTasks(logger)

		default:
			return http.StatusMethodNotAllowed, nil
		}

	case strings.HasPrefix(requestURL, "/tasks/"):
		return server.handleTaskStatus(
			logger,
			strings.TrimPrefix(requestURL, "/tasks/"),
		)

	default:
		return http.StatusNotFound, nil
	}
}

func (server *WebServer) handleNewTask(
	logger *lorg.Log,
	request *http.Request,
) (status int, response interface{}) {
	err := request.ParseForm()
	if err != nil {
		logger.Error(err)
		return http.StatusNotFound, nil
	}

	task, err := NewTaskStashPullRequest(
		request.PostForm.Get("url"),
	)
	if err != nil {
		logger.Error(err)
		return http.StatusBadRequest, err
	}

	taskID := server.resources.queue.Push(task)

	return http.StatusOK, ResponseTaskQueued{ID: taskID}
}

func (server *WebServer) handleTaskStatus(
	logger *lorg.Log,
	requestURL string,
) (status int, response interface{}) {
	taskID, err := strconv.Atoi(
		strings.Trim(strings.TrimPrefix(requestURL, "/task/"), "/"),
	)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if taskID > len(server.resources.queue.tasks) || taskID < 1 {
		return http.StatusNotFound, nil
	}

	task := server.resources.queue.tasks[taskID-1]

	return http.StatusOK, ResponseTask{
		ID:    task.GetID(),
		State: task.GetState().String(),
		Title: task.GetTitle(),
		Logs: strings.Split(
			strings.TrimSuffix(task.GetBuffer().String(), "\n"),
			"\n",
		),
	}
}

func (server *WebServer) handleListTasks(
	logger *lorg.Log,
) (status int, response interface{}) {
	tasksList := ResponseTaskList{
		Tasks: make([]ResponseTask, len(server.resources.queue.tasks)),
	}

	for i := len(server.resources.queue.tasks) - 1; i >= 0; i-- {
		task := server.resources.queue.tasks[i]

		tasksList.Tasks = append(
			tasksList.Tasks,
			ResponseTask{
				ID:    task.GetID(),
				State: task.GetState().String(),
				Title: task.GetTitle(),
			},
		)
	}

	return http.StatusOK, tasksList
}
