package tn

type (
	Confirm struct {
		Action  string   `json:"action"`
		Charset string   `json:"charset"`
		Hash    string   `json:"hash"`
		Teasers []string `json:"teasers"`
	}
	ConfirmResponse struct {
		Result bool `json:"result"`
	}
)

const (
	confirmAction  = "confirmView"
	confirmCharset = "utf-8"
)

func NewConfirm(resp *Response, shown []string) *Confirm {
	confirm := &Confirm{}
	confirm.Action = confirmAction
	confirm.Charset = confirmCharset
	confirm.Hash = resp.Hash
	confirm.Teasers = shown
	return confirm
}
