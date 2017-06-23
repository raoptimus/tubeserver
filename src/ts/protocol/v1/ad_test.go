package v1

import (
	"github.com/raoptimus/gserv/config"
	"strings"
	"testing"
	"ts/data"
)

func ExampleAdList() {
	runExample(`{"method": "AdController.List", "params": [{"Lang": "ru", "Token": "` + TEST_TOKEN + `", "Page": {"Skip": 0, "Limit": 2}}], "id": "1"}`)
	// Output:
}

func ExampleBannerInfo() {
	runExample(`{"method": "AdController.BannerInfo", "params": [{"Token": "` + TEST_TOKEN + `"}], "id": "1"}`)
	// Output:
}

func TestAdList(t *testing.T) {
	data := []struct {
		mustFail  bool
		mustEmpty bool
		lang      data.Language
		token     string
		skip      int
		limit     int
		carrier   string
		isp       string
	}{
		{false, false, data.LanguageRussian, TEST_TOKEN, 0, 10, "", "ISP"},
		{false, false, data.LanguageRussian, TEST_TOKEN, 1, 10, "", "ISP"},
		{false, true, data.LanguageRussian, TEST_TOKEN, 0, 10, "", ""},
		{false, false, data.LanguageRussian, TEST_TOKEN, 0, 10, "mobile provider", ""},
		{false, false, data.LanguageRussian, TEST_TOKEN, 1, 10, "mobile provider", ""},
		{false, true, data.LanguageEnglish, TEST_TOKEN, 0, 10, "", ""},
		{false, false, data.LanguageEnglish, TEST_TOKEN, 0, 10, "mobile provider", ""},
		{false, false, data.LanguageEnglish, TEST_TOKEN, 0, 10, "", "ISP"},
		{true, true, data.LanguageEnglish, "nonsence", 0, 10, "", ""},
	}
	for _, q := range data {

		// для теста таргетинга требуется обновить данные об используемом интернет-соединении
		updateNetReq := Request{
			Token: TEST_TOKEN,
			Object: DeviceNet{
				ISP:     q.isp,
				Carrier: q.carrier,
			},
		}
		var updateNetSuccess bool
		if err := getRpcClient().Call("DeviceController.UpdateNet", &updateNetReq, &updateNetSuccess); err != nil {
			t.Fatal("UpdateNet failed: " + err.Error())
		}
		if !updateNetSuccess {
			t.Fatal("UpdateNet failed")
		}

		// normal test
		req := Request{
			Lang:  q.lang,
			Token: q.token,
			Page: &Page{
				Skip:  q.skip,
				Limit: q.limit,
			},
		}
		t.Log("Test data:", q)
		var list AdList
		err := getRpcClient().Call("AdController.List", &req, &list)

		if q.mustFail {
			if err == nil {
				t.Fatal("Expected error but got nil")
			} else {
				continue
			}
		}
		if err != nil {
			t.Fatalf("call error: %s", err.Error())
		}

		//		if q.mustEmpty && len(list) != 0 {
		//			t.Fatal("Returned data is not empty")
		//		}
		//		if !q.mustEmpty && len(list) == 0 {
		//			t.Fatalf("Returned data is empty (skip: %d, limit: %d)", q.skip, q.limit)
		//		}

		checks := func(k, v string) {
			if v == "" {
				t.Fatalf("Data is not valid: %s is empty", k)
			}
		}
		for _, v := range list {
			// t.Logf("%#v\n", v)
			checks(v.Title, "Title")
			checks(v.Name, "Name")
			checks(v.Desc, "Desc")
			if !strings.HasPrefix(v.Icon, "http://") {
				t.Fatalf("Icon url is invalid: %s", v.Icon)
			}
			if len(v.Images) == 0 {
				t.Fatal("No images")
			}
			for _, url := range v.Images {
				if !strings.HasPrefix(url, "http://") {
					t.Fatalf("Image url is invalid: %s", url)
				}
			}
		}
	}
}

func TestBannerInfo(t *testing.T) {
	req := Request{Token: TEST_TOKEN}
	var bannerInfo BannerInfo

	if err := getRpcClient().Call("AdController.BannerInfo", &req, &bannerInfo); err != nil {
		t.Fatalf("call error: %v", err)
	}
	if bannerInfo.NeedBanner != config.Bool("NeedBanner", false) {
		t.Fatalf("Want NeedBanner = %v but got %v",
			config.Bool("NeedBanner", false), bannerInfo.NeedBanner)
	}
	if bannerInfo.BannerFrequency != config.Int("BannerFrequency", 0) {
		t.Fatalf("Want BannerFrequency = %v but got %v",
			bannerInfo.BannerFrequency, config.Int("BannerFrequency", 0))
	}
	if !strings.Contains(bannerInfo.BannersUrl, config.String("BannersUrl", "")) {
		t.Fatalf("Want BannersUrl = %v but got %v",
			bannerInfo.BannersUrl, config.String("BannersUrl", ""))
	}
}
