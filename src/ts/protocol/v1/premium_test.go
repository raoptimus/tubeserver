package v1

import (
	"testing"
	"ts/data"
)

func ExampleTariffList() {
	runExample(`{"method": "PremiumController.TariffList", "params": [{"Token": "` + TEST_TOKEN + `"}], "id": "1"}`)
	// Output:
}

func TestTariffList(t *testing.T) {
	req := Request{
		Lang:  data.LanguageEnglish,
		Token: TEST_TOKEN,
	}
	var list TariffList
	if err := getRpcClient().Call("PremiumController.TariffList", &req, &list); err != nil {
		t.Fatal("Rpc call error:", err)
	}

	if len(list) == 0 {
		t.Fatal("Returned data is empty")
	}
	for _, tariff := range list {
		if tariff.Title == "" {
			t.Fatal("Tariff's title is empty:", tariff)
		}
		if tariff.Duration == 0 {
			t.Fatal("Tariff's duration is zero:", tariff)
		}
		if tariff.Price == "" {
			t.Fatal("Tariff's price is empty:", tariff)
		}
	}
}
