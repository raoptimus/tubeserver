package main

import (
	"errors"
	"fmt"
	"github.com/raoptimus/gserv/service"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"ts/data"
)

const MOBIONET_API_URL = "http://core.mobionetwork.me/receive/coreSec.php?offer_id=%s&transaction_id=%s&device_model=%s&device_brand=%s&device_os=%s&device_os_version=%s"
const CPAPLANET_API_URL = "http://cpaplanet.net/postback/?o=%s&a=%s&transaction_id=%s&p=114"
const WAPEMPIRE_API_URL = "http://track.wapempire.com/aff_lsr?transaction_id=%s"

const HTTP_TIMEOUT = 10 * time.Second
const HTTP_MAX_IDLE_CONNS_PER_HOST = 1024

type (
	postBack struct {
		t *http.Transport
	}
)

var PostBack *postBack

func init() {
	transport := &http.Transport{}
	dial := &net.Dialer{Timeout: time.Duration(HTTP_TIMEOUT)}
	transport.Dial = dial.Dial
	transport.MaxIdleConnsPerHost = HTTP_MAX_IDLE_CONNS_PER_HOST
	PostBack = &postBack{t: transport}
}

func (s *postBack) GoSend(d *data.Device) {
	u, err := s.getUrl(d)
	if err != nil {
		fmt.Println(err)
		return
	}
	if u != nil {
		go func(u *url.URL) {
			var err error
			//			if u.Host == "core.mobionetwork.me" {
			//				err = s.sendGet(u)
			//			} else {
			//				err = s.sendPost(u)
			//			}
			err = s.sendGet(u)
			if err != nil {
				fmt.Println(err)
			}
		}(u)
	}
}

func (s *postBack) Send(d *data.Device) (err error) {
	u, err := s.getUrl(d)
	if err != nil {
		return err
	}
	if u != nil {
		return s.sendGet(u)
		//		if u.Host == "core.mobionetwork.me" {
		//			return s.sendGet(u)
		//		} else {
		//			return s.sendPost(u)
		//		}
	}
	return nil
}

func (s *postBack) getUrl(d *data.Device) (u *url.URL, err error) {
	src := d.Source
	switch src.Partner {
	case "mobionetwork":
		{
			return url.Parse(fmt.Sprintf(MOBIONET_API_URL, src.OffId, src.TransId, d.Model, d.Manufacture, d.Os, d.VerOs))
		}
	case "cpaplanet":
		{
			return url.Parse(fmt.Sprintf(CPAPLANET_API_URL, src.OffId, src.AffId, src.TransId))
		}
	case "wapempire":
		{
			return url.Parse(fmt.Sprintf(WAPEMPIRE_API_URL, src.TransId))
		}
	default:
		{
			return nil, nil
		}
	}
}

func (s *postBack) sendPost(u *url.URL) (err error) {
	defer service.DontPanic()
	req, err := http.NewRequest("POST", u.Scheme+"://"+u.Host+u.Path, strings.NewReader(u.Query().Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return s.sendReq(req)
}

func (s *postBack) sendGet(u *url.URL) (err error) {
	defer service.DontPanic()
	req, err := http.NewRequest("GET", u.String(), nil)
	req.Header.Add("Content-Type", "text/plain")
	return s.sendReq(req)
}

func (s *postBack) sendReq(req *http.Request) (err error) {
	timer := time.AfterFunc(time.Duration(HTTP_TIMEOUT), func() {
		s.t.CancelRequest(req)
	})
	client := &http.Client{
		Transport: s.t,
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("Postback err: " + err.Error())
	}
	timer.Stop()
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Postback err: Response status is %d", resp.StatusCode))
	}
	ctx, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("Postback err: " + err.Error())
	}
	fmt.Println(req.URL, resp.StatusCode, "'", string(ctx), "'")
	return nil
}
