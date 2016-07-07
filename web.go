package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/kovetskiy/lorg"
)

func (server *WebServer) HandleWeb(
	writer http.ResponseWriter, request *http.Request,
) {
	var (
		number = atomic.AddInt64(&server.requests, 1)
		logger = getLogger("web#%d", number)
	)

	logger.Infof(
		"-> %s %s",
		request.Method, request.URL,
	)

	requestURL := request.URL.Path

	switch {
	case strings.HasPrefix(requestURL, pathStatus):
		logger.Infof("handled request: get task status")
		server.handleStatus(
			writer, logger,
			strings.Trim(strings.TrimPrefix(requestURL, pathStatus), "/"),
		)

	case strings.HasPrefix(requestURL, pathBadge):
		server.handleBadge(
			writer,
			logger,
			strings.Trim(strings.TrimPrefix(requestURL, pathBadge), "/"),
		)

	default:
		writeStatus(writer, logger, http.StatusNotFound)
	}
}

func (server *WebServer) getTask(logger *lorg.Log, query string) (Task, error) {
	var task Task
	if !strings.Contains(query, "/") {
		taskID, err := strconv.Atoi(query)
		if err != nil {
			return nil, err
		}

		logger.Debugf("get task by unique id = %d", taskID)
		task = server.resources.queue.GetTaskByUniqueID(taskID)
	} else {
		logger.Debugf("get task by identifier = %s", query)
		task = server.resources.queue.GetTaskByIdentifier(query)
	}

	return task, nil
}

func (server *WebServer) handleStatus(
	writer http.ResponseWriter,
	logger *lorg.Log,
	query string,
) {
	task, err := server.getTask(logger, query)
	if err != nil {
		writeStatus(writer, logger, http.StatusBadRequest)
		logger.Error(err)
		return
	}

	if task == nil {
		writeStatus(writer, logger, http.StatusNotFound)
		return
	}

	writeStatus(writer, logger, http.StatusOK)
	fmt.Fprintf(writer, "%s\n----\n%s", task.GetState(), task.GetBuffer())
}

func (server *WebServer) handleBadge(
	writer http.ResponseWriter,
	logger *lorg.Log,
	query string,
) {
	task, err := server.getTask(logger, query)
	if err != nil {
		writeStatus(writer, logger, http.StatusBadRequest)
		logger.Error(err)
		return
	}

	if task == nil {
		writeStatus(writer, logger, http.StatusNotFound)
		return
	}

	var path string
	switch task.GetState() {
	case TaskStateSuccess:
		path = pathStaticBadgeBuildPassing

	case TaskStateError:
		path = pathStaticBadgeBuildFailure

	default:
		path = pathStaticBadgeBuildProcessing
	}

	logger.Infof("<- %s", path)
	writer.Header().Set("Location", path)
	writeStatus(writer, logger, http.StatusTemporaryRedirect)
}

func writeStatus(writer http.ResponseWriter, logger lorg.Logger, status int) {
	logger.Infof("<- %d %s", status, http.StatusText(status))
	writer.WriteHeader(status)
}
