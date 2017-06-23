package tn

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/raoptimus/gserv/service"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	apiUrl              = "http://..."
	httpGetTimeout      = 3000 * time.Millisecond
	httpConfirmTimeout  = 10000 * time.Millisecond
	maxIdleConnsPerHost = 1000
)

type Transport struct {
	confirmTransport *http.Transport
	doTransport      *http.Transport
}

func NewTransport() *Transport {
	return &Transport{
		doTransport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: httpGetTimeout,
			}).Dial,
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		confirmTransport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: httpConfirmTimeout,
			}).Dial,
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func (s *Transport) Do(req *Request) (resp *Response, err error) {
	b, err := json.Marshal(&req)
	if err != nil {
		err = errors.New("The data of request can't be encode to json: " + err.Error())
		return
	}
	isDebug := req.User.Ip == "...." || req.User.Ip == "..."

	q := string(b)
	data := url.Values{"json": {q}}

	//req.Protocol
	b, err = s.download(apiUrl, data, s.doTransport)
	if err != nil {
		err = errors.New("The response can't be read: " + err.Error())
		return
	}

	if isDebug {
		log.Println("\n\nRequest:\n", q, "\nResponse:\n", string(b))
	}

	if err = json.Unmarshal(b, &resp); err != nil {
		err = errors.New("The response can't be encod to json: " + err.Error())
		if isDebug {
			log.Println(err)
		}
		return
	}

	if resp.Error != "" {
		err = errors.New("The response has returned the error: " + resp.Error)
		if isDebug {
			log.Println(err)
		}
		return
	}

	if len(resp.Teasers) < 1 {
		err = fmt.Errorf("Little ads %d of %d", len(resp.Teasers), req.Ad.Amount)
		if isDebug {
			log.Println(err)
		}
		return
	}

	return
}

func (s *Transport) DoConfirm(resp *Response, shown []string) error {
	defer service.DontPanic()

	confirm := NewConfirm(resp, shown)
	b, err := json.Marshal(&confirm)
	if err != nil {
		return errors.New("The data of confirm can't be serialization: " + err.Error())
	}

	data := url.Values{"json": {string(b)}}
	b, err = s.download(apiUrl, data, s.confirmTransport)
	if err != nil {
		return fmt.Errorf("It isn't downloading: %v", err)
	}

	var status ConfirmResponse
	if err = json.Unmarshal(b, &status); err != nil {
		return errors.New("Result of confirm isn't unserialization: " + err.Error())
	}

	if !status.Result {
		return errors.New("Result confirmation is not ok: " + string(b))
	}

	return nil
}

func (s *Transport) download(url string, p url.Values, t *http.Transport) ([]byte, error) {
	client := &http.Client{}
	client.Transport = t
	httpReq, err := http.NewRequest("POST", url, strings.NewReader(p.Encode()))
	httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Add("Accept", "application/json")

	timer := time.AfterFunc(httpConfirmTimeout, func() {
		t.CancelRequest(httpReq)
	})
	httpResp, err := client.Do(httpReq)
	timer.Stop()

	if err != nil {
		return []byte{}, err
	}

	if httpResp.StatusCode != 200 {
		t.CancelRequest(httpReq)
		return []byte{}, fmt.Errorf("Response is returning status %d", httpResp.StatusCode)
	}

	defer httpResp.Body.Close()
	return ioutil.ReadAll(httpResp.Body)
}
