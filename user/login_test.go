package user

import (
	"os"
	"testing"
)

func BenchmarkLogin(b *testing.B) {
	loginData := LoginData{
		Userid: os.Getenv("USERID"),
		Userpw: os.Getenv("USERPW"),
	}

	for i := 0; i < b.N; i++ {
		runLogin(loginData)
	}
}
