package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/kovetskiy/lorg"
	"github.com/seletskiy/hierr"
)

type HTTPHandler struct {
	address   string
	logger    *lorg.Log
	resources *Resources
	handled   int64
}

func NewHTTPHandler(
	logger *lorg.Log,
	resources *Resources,
) *HTTPHandler {
	handler := &HTTPHandler{
		logger:    logger,
		resources: resources,
	}

	return handler
}

func (handler *HTTPHandler) Listen(address string) error {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return hierr.Errorf(
			err, "can't resolve %s", address,
		)
	}

	handler.logger.Debugf("opening connection at %s", address)

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't listen at %s", addr,
		)
	}

	handler.logger.Infof("listening at %s", address)

	return http.Serve(listener, handler)
}

func (handler *HTTPHandler) ServeHTTP(
	response http.ResponseWriter, request *http.Request,
) {
	var (
		requestID = atomic.AddInt64(&handler.handled, 1)
		logger    = coreLogger.NewChildWithPrefix(
			fmt.Sprintf("[request#%d]", requestID),
		)
	)

	handler.logger.Debugf("request #%d handled", requestID)
	defer func() {
		handler.logger.Debugf("request #%d released", requestID)
	}()

	logger.Infof(
		"-> %s %s",
		request.Method, request.URL.String(),
	)

	data, status := handler.handle(
		logger, request, request.URL.Path,
	)

	response.WriteHeader(status)

	if data != nil {
		if _, ok := data.(error); ok {
			data = map[string]interface{}{"error": data}
		}

		err := json.NewEncoder(response).Encode(data)
		if err != nil {
			logger.Error(err)
		}
	}

	logger.Infof(
		"<- %d %s",
		status, http.StatusText(status),
	)
}

func (handler *HTTPHandler) handle(
	logger *lorg.Log,
	request *http.Request,
	requestURL string,
) (interface{}, int) {
	switch {
	case strings.HasPrefix(requestURL, "/x/"):
		return handler.handleNewTask(
			logger,
			strings.TrimPrefix(requestURL, "/x/"),
		)

	case requestURL == "/task/":
		return handler.handleTaskList(logger)

	case strings.HasPrefix(requestURL, "/task/"):
		return handler.handleTaskStatus(
			logger,
			strings.TrimPrefix(requestURL, "/task/"),
		)

	default:
		return nil, http.StatusNotAcceptable
	}
}

func (handler *HTTPHandler) handleNewTask(
	logger *lorg.Log,
	requestURL string,
) (interface{}, int) {
	task, err := NewTaskStashPullRequest(
		strings.TrimPrefix(requestURL, "/x/"),
	)
	if err != nil {
		logger.Error(err)
		return err, http.StatusBadRequest
	}

	taskID := handler.resources.queue.Push(task)

	return ResponseTaskQueued{
		ID: taskID,
	}, http.StatusOK
}

func (handler *HTTPHandler) handleTaskStatus(
	logger *lorg.Log,
	requestURL string,
) (interface{}, int) {
	taskID, err := strconv.Atoi(
		strings.Trim(strings.TrimPrefix(requestURL, "/task/"), "/"),
	)
	if err != nil {
		return err, http.StatusBadRequest
	}

	if taskID > len(handler.resources.queue.tasks) || taskID < 1 {
		return nil, http.StatusNotFound
	}

	task := handler.resources.queue.tasks[taskID-1]

	return ResponseTask{
		ID:    task.GetID(),
		State: task.GetState().String(),
		Title: task.GetTitle(), 
		Logs: strings.Split(
			strings.TrimSuffix(task.GetBuffer().String(), "\n"),
			"\n",
		),
	}, http.StatusOK
}

func (handler *HTTPHandler) handleTaskList(
	logger *lorg.Log,
) (interface{}, int) {
	response := ResponseTaskList{}
	for i := len(handler.resources.queue.tasks) - 1; i >=  0; i-- {
		task  := handler.resources.queue.tasks[i]

		response.Tasks = append(
			response.Tasks,
			ResponseTask{
				ID: task.GetID(),
				State: task.GetState().String(),
				Title: task.GetTitle(),
			},
		)
	}

	return response, http.StatusOK
}
