package appleTools

import (
	"fmt"
	"testing"
)

func TestItunesLogin(t *testing.T) {
	if data, err := ItunesLogin("luisalfonso.boffill@gmail.com", "Boffill@12"); err != nil {
		fmt.Println(err)
	} else if m, err := data.GeyMoney(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("余额", m)
	}
}
