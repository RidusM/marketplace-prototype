package entity

import (
	"github.com/go-playground/assert/v2"
	"testing"

	"github.com/google/uuid"
	"github.com/thiagozs/go-phonegen"
)

func TestNew(t *testing.T) {
	globalID := uuid.New()
	username := "Username"
	email := "randomEmail@gmail.com"
	fullname := Fullname{Firstname: "Dmitriy", Middlename: "Olegovich", Lastname: "Pushkin"}
	phoneNumber := newPhoneNum()

	desiredOutput := &UserProfile{UserID: globalID, Username: "Username",
		Fullname:    Fullname{Firstname: "Dmitriy", Middlename: "Olegovich", Lastname: "Pushkin"},
		Email:       "randomEmail@gmail.com",
		PhoneNumber: phoneNumber,
	}

	output := New(globalID, username, fullname.Firstname, fullname.Middlename, fullname.Lastname, phoneNumber, email)

	assert.Equal(t, output.Username, desiredOutput.Username)
	assert.Equal(t, output.Email, desiredOutput.Email)
	assert.Equal(t, output.PhoneNumber, desiredOutput.PhoneNumber)
	assert.Equal(t, output.Fullname, desiredOutput.Fullname)
	assert.Equal(t, output.UserID, desiredOutput.UserID)
}

func TestValid(t *testing.T) {
	type testCase struct {
		input       *UserProfile
		throwsError bool
	}

	testCases := []testCase{
		{input: &UserProfile{Username: "", PhoneNumber: newPhoneNum(), Email: "default@gmail.com"},
			throwsError: true},
		{input: &UserProfile{Username: "username", PhoneNumber: "", Email: "default@gmail.com"}, throwsError: true},
		{input: &UserProfile{Username: "username", PhoneNumber: newPhoneNum(), Email: ""}, throwsError: true},
		{input: &UserProfile{Username: "username", PhoneNumber: newPhoneNum(), Email: "default@gmail.com"}, throwsError: false},
	}

	for _, test := range testCases {
		err := Valid(test.input)
		if test.throwsError != (err != nil) {
			t.Errorf("Incorrect result")
		}
	}
}

func newPhoneNum() string {
	number, _ := phonegen.New().RandomE164(1, "421")
	return number[0]
}
