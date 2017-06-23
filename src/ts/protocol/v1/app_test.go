package v1

import (
	"testing"
	"ts/data"
)

func TestAppLastBuild(t *testing.T) {
	req := Request{
		Lang:  data.LanguageRussian,
		Token: TEST_TOKEN,
		Query: &Query{
			"Ver":     TEST_VER,
			"WithApk": false,
		},
	}
	var app Application
	if err := getRpcClient().Call("AppController.LastBuild", &req, &app); err != nil {
		t.Fatal(err.Error())
	}
}

func TestAppIpList(t *testing.T) {
	req := Request{
		Token: TEST_TOKEN,
	}
	var addrList AddrList
	if err := getRpcClient().Call("AppController.IpList", &req, &addrList); err != nil {
		t.Fatal(err.Error())
	}
	if len(addrList) == 0 {
		t.Fatal("Ip list can't be empty")
	}
	for _, addr := range addrList {
		if addr.Ip == "" {
			t.Fatal("Ip can't be blank")
		}
		if addr.BaseIp == "" {
			t.Fatal("BaseIp can't be blank")
		}
	}
}
