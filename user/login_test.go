package user

import (
	"context"
	"os"
	"testing"

	"github.com/s-owl/skhus-backend/browser"
)

func BenchmarkLogin(b *testing.B) {
	userData := loginData{
		Userid: os.Getenv("USERID"),
		Userpw: os.Getenv("USERPW"),
	}
	Browser := browser.NewBrowser(context.Background())
	defer Browser.Close()

	for i := 0; i < b.N; i++ {
		tabForest, fcf := Browser.NewContext()
		tabSam, scf := Browser.NewContext()
		defer fcf()
		defer scf()

		forestResult := loginOnForest(tabForest, &userData)
		samResult := loginOnSam(tabSam, &userData)
		<-forestResult
		<-samResult
	}
}
