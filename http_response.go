package main

type ResponseTaskQueued struct {
	ID int64
}

type ResponseTask struct {
	ID    int64  `json:"id"`
	State string `json:"state"`
	Title string `json:"title"`
	Logs  []string `json:"logs,omitempty"`
}


type ResponseTaskList struct {
	Tasks []ResponseTask `json:"tasks"`
}
