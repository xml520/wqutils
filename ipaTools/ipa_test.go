package ipaTools

import (
	"log"
	"testing"
)

func TestNewIpaParser(t *testing.T) {
	ipa, err := NewIpaParser("./test.ipa")
	if err != nil {
		log.Fatalln(err)
	}
	defer ipa.Close()
	err = ipa.WriteInfo(map[string]any{"123": "123"})
	if err != nil {
		log.Fatalln("写失败", err)
	}
}
