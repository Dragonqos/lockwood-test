package websocket

type Request struct {
	RID     any    `json:"rID" validate:"required"`
	Cmd     string `json:"cmd" validate:"required,max=64"`
	User    string `json:"user" validate:"omitempty,max=128"`
	Name    string `json:"name" validate:"omitempty,max=128"`
	Message string `json:"message" validate:"omitempty,max=4096"`
}

type response struct {
	RID    any    `json:"rID"`
	Cmd    string `json:"cmd"`
	Status string `json:"status"`
}
