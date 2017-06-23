package v1

type (
	Application struct {
		Name         string
		Ver          string
		BuildVer     string
		Description  string
		Url          string `json:"Url,omitempty"`
		Apk          string `json:"Apk,omitempty"`
		CurrVerAllow bool
	}
	Addr struct {
		Ip     string
		Port   int
		BaseIp string
		Ssl    bool `json:"Ssl,omitempty"`
		Proxy  bool `json:"Proxy,omitempty"`
	}
	AddrList []*Addr
)
