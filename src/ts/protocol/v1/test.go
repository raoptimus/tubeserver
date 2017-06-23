package v1

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/raoptimus/gserv/config"
	"github.com/raoptimus/gserv/service"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"strconv"
	"strings"
	"ts/data"
)

type (
	addr struct {
		Ip     string
		Port   int
		BaseIp string
		Ssl    bool `json:"Ssl,omitempty"`
		Proxy  bool `json:"Proxy,omitempty"`
	}
	addrs     []*addr
	rpcClient struct {
		*rpc.Client
		conn io.ReadWriteCloser
	}
)

var TEST_DEVICE = Device{
	Id:            "506deee230739fa_",
	Os:            "Android",
	Type:          "tablet",
	VerOs:         "5.0.2",
	Serial:        "015d4a5ed91c1419",
	Manufacture:   "asus",
	Model:         "Nexus 7",
	SerialGsm:     "015d4a5ed91c1419",
	FileName:      "...cz0xMnRyYWZmaWNfNiZhPTU2Nzg5Jmw9MQ==.apk",
	WifiMac:       "3c:97:0e:b0:5d:92",
	AdvertisingId: "5177d563-e4b0-4b80-842a-6bdf62ed001a",
}

var TEST_TOKEN, _ = TEST_DEVICE.GenNewToken()

const TEST_AVATAR string = "AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAQAQAAAAAAAAAAAAAAAAAAAAAAAAiKCcJIignmSIoJ+0iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign7SIoJ5kiKCcJIignkCIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/IignjSIoJ/MiKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ+ciKCf/Iign/yIoJ/8iKCf/R0xL/0dMS/8iKCf/Iign/yIoJ/87QED/Sk9O/0pPTv9KT07/P0VE/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/0pPTv9KT07/Iign/yctLP9HTEv/Sk9O/0pPTv9KT07/Sk9O/0pPTv9HTEv/Iign/yIoJ/8iKCf/Iign/yIoJ/9KT07/Sk9O/yIoJ/8/RUT/Sk9O/0pPTv8sMjH/JSsq/ywxMP9KT07/R0xL/yIoJ/8iKCf/Iign/xFYk/8AiP//Lp3//zx2pv8iKCf/Sk9O/0pPTv8sMTD/EViT/wCI//8AiP//EViT/yIoJ/8iKCf/Iign/yIoJ/8AiP//AIj//y6d//8unf//IC41/0pPTv9KT07/IDZD/wCI//8Tkf//AIj//xOR//8iKCf/Iign/yIoJ/8iKCf/AIj//wCI//8unf//Lp3//x40Qv9KT07/Sk9O/yMwN/8AiP//Lp3//wCI//8unf//Iign/yIoJ/8iKCf/Iign/xFYk/8AiP//Lp3//zx2pv8iKCf/Sk9O/0pPTv8sMTD/EViT/wCI//8AiP//EViT/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/0pPTv9KT07/Iign/0dMS/9KT07/Sk9O/y40M/8lKyr/LTMy/0pPTv9HTEv/Iign/yIoJ/9KT07/Sk9O/0pPTv9KT07/Sk9O/0pPTv9KT07/Sk9O/0pPTv9KT07/Sk9O/0pPTv9KT07/Sk9O/yIoJ/8iKCf/R0xL/0pPTv9KT07/Sk9O/0pPTv9KT07/Sk9O/0dMS/87QED/Sk9O/0pPTv9KT07/P0VE/yIoJ/8iKCf/Iign8yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign8yIoJ4oiKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ4oiKCcJIignmSIoJ+0iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign/yIoJ/8iKCf/Iign7SIoJ5kiKCcJAAD//wAA//8AAP//AAD//wAA//8AAP//AAD//wAA//8AAP//AAD//wAA//8AAP//AAD//wAA//8AAP//AAD//w=="
const TEST_VER string = "1.0"

var client *rpcClient

func init() {
	defer service.DontPanic()
	data.Init(false)
}

func getBindAddr() (*addr, error) {
	intAddrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, errors.New("Can't read interface addrs: " + err.Error())
	}
	var bindAddrs addrs
	if err := config.Object("BindIp", &bindAddrs); err != nil {
		return nil, errors.New("Can't read BindIp data from config: " + err.Error())
	}
	for _, b := range bindAddrs {
		if b.Proxy {
			continue //proxy is not supporting
		}
		if b.Ip == "" {
			return b, nil
		}
		ip := net.ParseIP(b.Ip)
		for _, intAddr := range intAddrs {
			var intIp net.IP
			switch v := intAddr.(type) {
			case *net.IPNet:
				intIp = v.IP
			case *net.IPAddr:
				intIp = v.IP
			}
			if intIp.Equal(ip) {
				return b, nil
			}
		}
	}
	return nil, errors.New("Not found addr for binding")
}

func getRpcConn() io.ReadWriteCloser {
	addr, err := getBindAddr()
	if err != nil {
		log.Fatalln(err)
	}
	addrPort := addr.Ip + ":" + strconv.Itoa(addr.Port)

	if addr.Ssl {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: true,
		}
		conn, err := tls.Dial("tcp", addrPort, tlsCfg)
		if err != nil {
			log.Fatalln(fmt.Errorf("Can't connect to rpc tls server %s: %s", addrPort, err))
		}
		return conn
	}

	conn, err := net.Dial("tcp", addrPort)
	if err != nil {
		log.Fatalln(fmt.Errorf("Can't connect to rpc server %s: %s", addrPort, err))
	}
	return conn
}

func getRpcClient() *rpcClient {
	if client != nil {
		return client
	}
	conn := getRpcConn()
	return &rpcClient{
		conn:   conn,
		Client: jsonrpc.NewClient(conn),
	}
}

func checkVideoFileUrl(rawUrl string) error {
	_, err := url.Parse(rawUrl)
	if err != nil {
		return err
	}
	resp, err := http.Get(rawUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("Status code is no correct: " + resp.Status)
	}
	if resp.ContentLength == 0 {
		return errors.New("Content-Length is zero")
	}
	if !strings.Contains(resp.Header.Get("Content-Type"), "video/") {
		return errors.New("Content type is no correct: " + resp.Header.Get("Content-Type"))
	}
	return nil
}

func toJson(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
