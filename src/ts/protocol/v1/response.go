package v1

type ResponseStatus string

const (
	StatusOk        ResponseStatus = "ok"
	StatusNotFound  ResponseStatus = "not found"
	StatusError     ResponseStatus = "error"
	StatusForbidden ResponseStatus = "forbidden"
)

type Response struct {
	Status ResponseStatus `json:"status"`
	Error  string         `json:"error,omitempty"`
	Data   interface{}
}
