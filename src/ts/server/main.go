package main

import (
	"crypto/tls"
	"expvar"
	"fmt"
	"github.com/SpruceHealth/go-proxy-protocol/proxyproto"
	"github.com/raoptimus/gserv/config"
	"github.com/raoptimus/gserv/service"
	"github.com/raoptimus/rlog"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"ts/data"
	"ts/detect"
	api "ts/protocol/v1"
	"ts/ring"
)

const SERVER_TIMEOUT = time.Minute * 1

var log *rlog.Logger
var Geo *detect.Geo
var Mem *memContext
var Ips *IpTable

type profileOptions struct {
	Enabled bool
	Ip      string
	Port    int
}

var queryRing = ring.New(1000)
var queryRingDebug = ring.New(1000)

func init() {
	expvar.Publish("queries", expvar.Func(func() interface{} {
		return queryRing.List()
	}))
	expvar.Publish("debug-queries", expvar.Func(func() interface{} {
		return queryRingDebug.List()
	}))
}

func main() {
	if service.Exists() {
		os.Exit(0)
	}

	cs := config.String("MongoLogServer", config.String("MongoAllServer", "localhost/Logs"))
	var err error
	log, err = rlog.NewLoggerDial(rlog.LoggerTypeMongoDb, "", cs, "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	service.Init(
		&service.BaseService{
			Start:  start,
			Logger: log,
		})

	Ips = NewIpTable()
	data.Init(true)
	service.Start(true)
}

func start() {
	var (
		err         error
		profOptions profileOptions
	)
	if err := config.Object("Profiling", &profOptions); err != nil {
		fmt.Println("I don't read options for pprof: " + err.Error())
	} else if profOptions.Enabled {
		service.StartProfiler(profOptions.Ip + ":" + strconv.Itoa(profOptions.Port))
	}

	Geo, err = detect.NewGeo()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	Mem, err = NewMemContext()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	inIpList, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var list api.AddrList
	if err := config.Object("BindIp", &list); err != nil {
		fmt.Println("Can't read BindIp data from config ", err)
		os.Exit(1)
	}
	for _, t := range list {
		if t.Ip == "" {
			go listenJsonRpcServer(t.Ip, t.Port, t.Ssl, t.Proxy)
			break
		}
		ip := net.ParseIP(t.Ip)
		for _, inT := range inIpList {
			var inIp net.IP
			switch v := inT.(type) {
			case *net.IPNet:
				inIp = v.IP
			case *net.IPAddr:
				inIp = v.IP
			}
			if inIp.Equal(ip) {
				go listenJsonRpcServer(t.Ip, t.Port, t.Ssl, t.Proxy)
				break
			}
		}
	}
	log.Info("The server is ready to accept connections")
}

func listenJsonRpcServer(ip string, port int, ssl, proxy bool) {
	addr := ip + ":" + strconv.Itoa(port)
	server := newRpcServer()
	l, err := net.Listen("tcp", addr)
	fmt.Printf("Bind addr (ssl=%v,proxy=%v): %v\n", ssl, proxy, addr)

	if err != nil {
		fmt.Println("Listen error: " + err.Error())
		os.Exit(1)
	}

	var tlsCfg *tls.Config
	if ssl {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]) + "/../cert")
		if err != nil {
			fmt.Println("Cant find file in '" + os.Args[0] + "/../cert'")
			os.Exit(1)
		}
		cert, err := tls.LoadX509KeyPair(dir+"/cert.pem", dir+"/key.pem")
		if err != nil {
			fmt.Println("Error loading certificate: " + err.Error())
			os.Exit(1)
		}
		tlsCfg = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	if proxy {
		l = &proxyproto.Listener{
			Listener: l,
			Config:   tlsCfg,
			Timeout:  SERVER_TIMEOUT,
		}
	} else if ssl {
		l = tls.NewListener(l, tlsCfg)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Err(err.Error())
			continue
		}

		go func(conn net.Conn) {
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(SERVER_TIMEOUT))
			server.ServeCodec(NewServerCodec(conn))
		}(conn)
	}
}

func newRpcServer() *rpc.Server {
	controllers := [...]interface{}{
		new(CategoryController),
		new(VideoController),
		new(AdController),
		new(DeviceController),
		new(UserController),
		new(AppController),
		new(SearchController),
		new(PremiumController),
	}
	server := rpc.NewServer()
	for _, controller := range controllers {
		server.Register(controller)
	}
	return server
}
