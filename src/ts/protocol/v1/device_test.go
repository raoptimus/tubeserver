package v1

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"testing"
	"time"
	"ts/data"
	"ts/data/push"
)

func TestDoubleReg(t *testing.T) {
	antonNew := Device{
		Id:            "2f828c95089617df",
		Os:            "Android",
		Type:          "tablet",
		VerOs:         "4.4.2",
		Serial:        "IBUWU4Y5LV5DAELF",
		Manufacture:   "TCL",
		Model:         "I216X",
		SerialGsm:     "865499021029234",
		WifiMac:       "b0:e0:3c:53:87:f0",
		AdvertisingId: "ae8e6575-ed21-4291-87be-4c2b0c2035cb",
	}
	antonOld := antonNew
	antonOld.AdvertisingId = ""
	oldToken, err := antonOld.GenOldToken()
	if err != nil {
		t.Fatal(err)
	}
	oldToken2, err := antonNew.GenOldToken()
	if err != nil {
		t.Fatal(err)
	}

	if oldToken != oldToken2 {
		t.Fatalf("Old tokens %s != %s", oldToken, oldToken2)
	}
	newToken, err := antonNew.GenNewToken()
	if err != nil {
		t.Fatal(err)
	}

	data.Context.Devices.RemoveId(oldToken)
	data.Context.Devices.RemoveId(newToken)
	t.Logf("Old token: %s", oldToken)
	t.Logf("New token: %s", newToken)

	//reg2 old
	req := Request{
		Object: antonOld,
	}
	var token ReturnToken
	if err := getRpcClient().Call("DeviceController.Reg2", &req, &token); err != nil {
		t.Fatal(err)
	}
	if token.Exists {
		t.Fatalf("Token cant be exists")
	}
	if token.Token != oldToken {
		t.Fatalf("Token after reg %s != %s", token.Token, oldToken)
	}
	user, _ := data.GetUserByToken(oldToken)
	if user != nil {
		t.Logf("User %d is found by token %s", user.Id, oldToken)
	} else {
		t.Logf("User not found by old token %s", oldToken)
	}
	//reg2 new
	var token2 ReturnToken
	req2 := Request{
		//		Token:  oldToken,
		Object: antonNew,
	}
	if err := getRpcClient().Call("DeviceController.Reg2", &req2, &token2); err != nil {
		t.Fatal(err)
	}
	if !token2.Exists {
		t.Fatalf("Token can be exists")
	}
	if token2.Token != newToken {
		t.Fatalf("Token after reg %s != %s", token2.Token, newToken)
	}
	user2, _ := data.GetUserByToken(newToken)
	if user2 != nil {
		t.Logf("User %d is found by token %s", user2.Id, newToken)
	} else {
		t.Logf("User not found by new token %s", newToken)
	}
	if user != nil && user2 == nil {
		t.Fatalf("User not found after replace token %s to %", oldToken, newToken)
	}
	if user != nil && user2 != nil && user.Id != user2.Id {
		t.Fatalf("User %d != %d", user.Id, user2.Id)
	}
	n, err := data.Context.Devices.FindId(oldToken).Count()
	if err != nil {
		t.Fatal(err)
	}
	if n > 0 {
		t.Fatalf("Old device is exists after replace")
	}
}

//новые функции обяз после этой
func TestDeviceReg2(t *testing.T) {
	req := Request{
		Object: TEST_DEVICE,
	}
	//1
	var token ReturnToken
	if err := getRpcClient().Call("DeviceController.Reg2", &req, &token); err != nil {
		t.Fatal(err.Error())
	}
	if len(token.Token) != 32 {
		t.Fatal("Token format is incorrect")
	}
	//2
	var token2 ReturnToken
	if err := getRpcClient().Call("DeviceController.Reg2", &req, &token2); err != nil {
		t.Fatal(err.Error())
	}
	if len(token2.Token) != 32 {
		t.Fatal("Token format is not correct")
	}
	//
	if token2.Token != token.Token {
		t.Fatalf("Generate algoritm token error %v != %v", token.Token, token2.Token)
	}
}

func deviceRegRollback(oldToken, newToken string) error {
	d1, err := data.FindDevice(newToken)
	if err != nil {
		return err
	}
	if d1 == nil {
		return nil
	}

	user, err := data.GetUserByToken(newToken)
	if err != nil && err != mgo.ErrNotFound {
		return err
	}

	if user != nil {
		userTokens := []string{oldToken}
		for _, t := range user.Tokens {
			if *t == oldToken || *t == newToken {
				continue
			}
			userTokens = append(userTokens, *t)
		}
		err = data.Context.Users.UpdateId(user.Id, bson.M{"$set": bson.M{"Tokens": userTokens}})
		if err != nil {
			return err
		}
	}

	if err := data.Context.Devices.RemoveId(newToken); err != nil {
		return err
	}
	d1.Id = oldToken
	return data.Context.Devices.Insert(d1)
}

// https://workflowboard.com/jira/browse/TUBS-88
func TestDeviceReg2Migration(t *testing.T) {
	d := TEST_DEVICE
	oldToken, err := d.GenOldToken()
	if err != nil {
		t.Fatal(err)
	}
	newToken, err := d.GenNewToken()
	if err != nil {
		t.Fatal(err)
	}

	if err := deviceRegRollback(oldToken, newToken); err != nil {
		t.Fatal(err)
	}

	req := Request{
		Token: oldToken,
		Object: User{
			UserName: "testUser",
			Avatar:   TEST_AVATAR,
			Tel:      "7363736373",
			Email:    "test@test.com",
			Lang:     data.LanguageRussian,
		},
	}

	if err := getRpcClient().Call("UserController.UpdateInfo", &req, nil); err != nil {
		t.Fatal(err.Error())
	}

	oldUser, err := data.GetUserByToken(oldToken)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		apk, os         string
		serial, gsm, id string
		advId           string
		token           string
		expect          string
		user            *data.User
	}{
		{
			apk:    "old",
			os:     "new",
			serial: d.Serial,
		},
		{
			apk:    "old",
			os:     "old",
			serial: d.Serial,
			gsm:    d.SerialGsm,
			id:     d.Id,
			expect: oldToken,
			user:   oldUser,
		},
		{
			apk:    "new",
			os:     "old",
			serial: d.Serial,
			gsm:    d.SerialGsm,
			id:     d.Id,
			advId:  d.AdvertisingId,
			expect: newToken,
			user:   oldUser,
		},
		{
			apk:    "new",
			os:     "new",
			serial: d.Serial,
			advId:  d.AdvertisingId,
			token:  oldToken,
			expect: newToken,
			user:   oldUser,
		},
		{
			apk:    "new",
			os:     "new",
			serial: d.Serial,
			advId:  d.AdvertisingId,
			token:  newToken,
			expect: newToken,
			user:   oldUser,
		},
	}
	for _, c := range cases {
		dev := TEST_DEVICE
		dev.Serial = c.serial
		dev.SerialGsm = c.gsm
		dev.Id = c.id
		dev.AdvertisingId = c.advId

		if c.expect == "" {
			c.expect, _ = dev.GenTmpToken()
			c.user, _ = data.GetUserByToken(c.expect)
		}

		req := Request{
			Token:  c.token,
			Object: dev,
		}
		var token ReturnToken

		if err := getRpcClient().Call("DeviceController.Reg2", &req, &token); err != nil {
			t.Fatal(err)
		}

		var logInfo = func() {
			t.Logf("Test req: %+v\n", req)
			t.Logf("Test case: %+v\n", c)
		}
		if token.Token != c.expect {
			logInfo()
			t.Fatalf("Expected token %s but got %s", c.expect, token.Token)
		}

		if c.user != nil {
			user, err := data.GetUserByToken(token.Token)
			if err != nil {
				t.Fatal(err)
			}

			if user.Id != c.user.Id {
				logInfo()
				t.Fatalf("Expected user %s but got %s", c.user.Id, user.Id)
			}
		}
	}
}

func TestDeviceClickPush(t *testing.T) {
	var task push.Task
	err := data.Context.PushTask.Find(nil).One(&task)
	if err != nil {
		t.Fatal(err.Error())
	}
	req := Request{
		Token: TEST_TOKEN,
		Object: struct {
			Id string `bson:"Id"`
		}{
			Id: strconv.Itoa(task.Id),
		},
	}

	var result bool
	if err := getRpcClient().Call("DeviceController.ClickPush", &req, &result); err != nil {
		t.Fatal(err.Error())
	}
}

func TestDeviceUpdateInfo(t *testing.T) {
	req := Request{
		Token: TEST_TOKEN,
		Object: DeviceInfo{
			Loc:      "Etc/GMT-3",
			GoogleId: "123",
		},
	}

	var result bool
	if err := getRpcClient().Call("DeviceController.UpdateInfo", &req, &result); err != nil {
		t.Fatal(err.Error())
	}
}

func TestAddAction(t *testing.T) {
	req := Request{
		Token: TEST_TOKEN,
		Object: DeviceEvent{
			Details: "launch",
		},
	}
	var res bool
	if err := getRpcClient().Call("DeviceController.AddAction", &req, &res); err != nil {
		t.Fatal(err.Error())
	}
	if !res {
		t.Fatal("Result is false")
	}
}

func TestAddActionFirstLaunch(t *testing.T) {
	req := Request{
		Token: TEST_TOKEN,
		Object: DeviceEvent{
			Details: "first-launch",
		},
	}
	var res bool
	if err := getRpcClient().Call("DeviceController.AddAction", &req, &res); err != nil {
		t.Fatal(err.Error())
	}
	if !res {
		t.Fatal("Result is false")
	}
	var event data.DeviceEvent
	err := data.Context.DeviceEvents.Find(
		bson.M{"Action": data.DeviceActionFLaunch, "DeviceId": TEST_TOKEN}).
		Sort("-AddedDate").Limit(1).One(&event)
	if err != nil {
		t.Fatal(err)
	}

	if time.Now().UTC().Sub(event.AddedDate.UTC()) > 1*time.Second {
		t.Fatalf("First launch event is not found")
	}
}

func ExampleUpdateNet() {
	runExample(`{"method": "DeviceController.UpdateNet", "params": [{"Token": "` + TEST_TOKEN + `", "Object": {"ISP": "AT&T", "Carrier": "AT&T"}}], "id": "1"}`)
	// Output:
}

func TestUpdateNet(t *testing.T) {
	req := Request{
		Token: TEST_TOKEN,
		Object: DeviceNet{
			ISP:     "AT&T",
			Carrier: "AT&T",
		},
	}
	var res bool
	if err := getRpcClient().Call("DeviceController.UpdateNet", &req, &res); err != nil {
		t.Fatal(err.Error())
	}
	if !res {
		t.Fatal("Result is false")
	}
}
