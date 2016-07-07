package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/kovetskiy/lorg"
)

func (server *WebServer) HandleAPI(
	writer http.ResponseWriter, request *http.Request,
) {
	var (
		number = atomic.AddInt64(&server.requests, 1)
		logger = getLogger("api#%d", number)
	)

	logger.Infof(
		"-> %s %s",
		request.Method, request.URL,
	)

	status, response := server.routeAPI(
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

func (server *WebServer) routeAPI(
	logger *lorg.Log,
	request *http.Request,
) (status int, response interface{}) {
	requestURL := "/" + strings.TrimPrefix(request.URL.Path, pathAPI)

	switch {
	case requestURL == "/tasks/":
		switch request.Method {
		case "POST":
			logger.Infof("handled request: new task")
			return server.handleNewTask(logger, request)

		case "GET":
			logger.Infof("handled request: list tasks")
			return server.handleListTasks(logger)

		default:
			return http.StatusMethodNotAllowed, nil
		}

	case strings.HasPrefix(requestURL, "/tasks/"):
		logger.Infof("handled request: get task")
		return server.handleTask(
			logger,
			strings.Trim(strings.TrimPrefix(requestURL, "/tasks/"), "/"),
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

func (server *WebServer) handleTask(
	logger *lorg.Log,
	query string,
) (status int, response interface{}) {
	var task Task
	if !strings.Contains(query, "/") {
		taskID, err := strconv.Atoi(query)
		if err != nil {
			return http.StatusBadRequest, err
		}

		logger.Debugf("get task by unique id = %d", taskID)
		task = server.resources.queue.GetTaskByUniqueID(taskID)
	} else {
		logger.Debugf("get task by identifier = %s", query)
		task = server.resources.queue.GetTaskByIdentifier(query)
	}

	if task == nil {
		return http.StatusNotFound, nil
	}

	return http.StatusOK, ResponseTask{
		UniqueID:   task.GetUniqueID(),
		Identifier: task.GetIdentifier(),
		State:      task.GetState().String(),
		Title:      task.GetTitle(),
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
		Tasks: make([]ResponseTask, 0),
	}

	for i := len(server.resources.queue.tasks) - 1; i >= 0; i-- {
		task := server.resources.queue.tasks[i]

		tasksList.Tasks = append(
			tasksList.Tasks,
			ResponseTask{
				UniqueID:   task.GetUniqueID(),
				Identifier: task.GetIdentifier(),
				State:      task.GetState().String(),
				Title:      task.GetTitle(),
			},
		)
	}

	return http.StatusOK, tasksList
}
