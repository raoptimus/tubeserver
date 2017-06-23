package v1

import (
	"fmt"
	"testing"
	"ts/data"
)

func TestUserUpdate(t *testing.T) {
	req := Request{
		Token: TEST_TOKEN,
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

	// check that empty data will not override existing data
	req.Object = User{
		Avatar:   "",
		UserName: "",
		Email:    "",
		Tel:      "911",
		Lang:     data.LanguageEnglish,
	}

	if err := getRpcClient().Call("UserController.UpdateInfo", &req, nil); err != nil {
		t.Fatal(err.Error())
	}

	user, err := data.GetUserByToken(TEST_TOKEN)
	if err != nil {
		t.Fatal(err)
	}
	switch {
	case len(user.Avatar.Data) == 0:
		t.Log(user.Avatar)
		t.Fatal("Avatar.Data is empty")
	case user.UserName == "":
		t.Fatal("UserName is empty")
	case user.Email == "":
		t.Fatal("Email is empty")
	case user.Tel != "911":
		t.Fatal("Tel is not 911")
	case user.Language != data.LanguageEnglish:
		t.Fatal("Language must be [en]")
	}
}

func ExampleGetUser() {
	runExample(`{"method": "UserController.GetInfo", "params": [{"Token": "` + TEST_TOKEN + `"}], "id": "1"}`)
	// Output:
}

func TestUserGet(t *testing.T) {
	var u User
	req := Request{
		Token: TEST_TOKEN,
	}

	if err := getRpcClient().Call("UserController.GetInfo", req, &u); err != nil {
		t.Fatal(err.Error())
	}

	if u.Id <= 0 {
		t.Fatal("Data is not valid")
	}

	if u.Avatar != TEST_AVATAR {
		fmt.Println(toJson(u))
		t.Fatal("Data (Avatar) is not valid")
	}
}

func ExamplePremiumStatus() {
	runExample(`{"method": "UserController.GetPremiumStatus", "params": [{"Token": "` + TEST_TOKEN + `"}], "id": "1"}`)
	// Output:
}

func TestUserGetPremiumStatus(t *testing.T) {
	var premium Premium
	req := Request{
		Token: TEST_TOKEN,
	}

	if err := getRpcClient().Call("UserController.GetPremiumStatus", req, &premium); err != nil {
		t.Fatal(err.Error())
	}

	switch premium.Type {
	case data.PremiumTypeNone:
		if premium.Duration > 0 {
			t.Fatal("Duration must not be positive for UserTypeNone")
		}
	case data.PremiumTypeTrial, data.PremiumTypeSignup:
		if premium.Duration == 0 {
			t.Fatal("Duration must be positive for UserTypeTrial/Signup")
		}
	default:
		t.Fatalf("Unknown type: %q", premium.Type)
	}

}
