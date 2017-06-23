package main

import (
	"bytes"
	"errors"
	"gopkg.in/mgo.v2"
	"io"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
	"ts/data"
	api "ts/protocol/v1"
)

type serverCodec struct {
	rpc.ServerCodec
	conn  *serverConn
	token string
	ver   string
	*rpc.Request
}

type serverConn struct {
	net.Conn
	buf bytes.Buffer
}

func (sc *serverConn) Read(p []byte) (n int, err error) {
	sc.buf.Reset()
	tee := io.TeeReader(sc.Conn, &sc.buf)
	return tee.Read(p)
}

func NewServerCodec(conn net.Conn) *serverCodec {
	sc := &serverConn{Conn: conn}
	return &serverCodec{
		ServerCodec: jsonrpc.NewServerCodec(sc),
		conn:        sc,
	}
}

func (s *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	err := s.ServerCodec.ReadRequestHeader(r)
	s.Request = r
	return err
}

func (s *serverCodec) ReadRequestBody(body interface{}) error {
	if err := s.ServerCodec.ReadRequestBody(body); err != nil {
		return err
	}

	debug := false
	raw := s.conn.buf.String()

	defer func(debug *bool, raw *string) {
		if *debug {
			queryRingDebug.Push(*raw)
			return
		}
		queryRing.Push(*raw)
	}(&debug, &raw)

	if req, ok := body.(*api.Request); ok {
		ip := s.conn.RemoteAddr().String()
		if p := strings.Index(ip, ":"); p != -1 {
			ip = ip[0:p]
		}
		debug = ip == "206.54.164.71" || ip == "206.54.164.72" || ip == "127.0.0.1" || IsDebug(req.Token)
		req.Ip = ip
		s.token = req.Token
		s.ver = req.Ver

		// check access token except Reg2
		if "DeviceController.Reg2" != s.Request.ServiceMethod {
			if err := data.ValidateToken(s.token); err != nil {
				return err
			}
			err := data.Context.Devices.FindId(s.token).One(&req.Device)
			if err != nil {
				return errors.New("Token " + err.Error())
			}

			if req.Ip != req.Device.LastIp || req.Device.LastGeo == nil {
				req.Geo = Geo.GetRecord(req.Ip)
			} else {
				req.Geo = req.Device.LastGeo
			}
		}
	}

	return nil
}

func (s *serverCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	err := s.ServerCodec.WriteResponse(r, x)

	if r.Error != "" && r.Error != mgo.ErrNotFound.Error() {
		log.Err(r.ServiceMethod + " " + r.Error)
	}

	rpcStat := data.NewRpcStat(&data.RpcStatSource{
		Ver:    s.ver,
		Method: r.ServiceMethod,
	})

	if err != nil {
		rpcStat.Counters.ErrorCount++
		log.Err(err.Error())
	} else {
		rpcStat.Counters.SuccessCount++
	}
	rpcStat.UpsertInc()
	return err
}
