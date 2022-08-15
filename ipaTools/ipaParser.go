package ipaTools

import (
	"archive/zip"
	"errors"
	"fmt"
	"howett.net/plist"
	"io"
	"os"
	"regexp"
)

type IpaParser struct {
	filename string
	file     *os.File
	list     []*zip.File
	info     *InfoPlist
}
type InfoPlist struct {
	header *zip.FileHeader
	path   string
	data   []byte
}

// ParserIpaInfo Unmarshal 解码plist  https://pkg.go.dev/howett.net/plist#Unmarshal
func ParserIpaInfo(filename string, data any) error {
	ipa, err := zip.OpenReader(filename)
	if err != nil {
		fmt.Errorf("无法读取ipa文件 %s", err)
	}
	defer ipa.Close()
	infoPlist := findZipFile(ipa.File, "Payload/[^/]*\\.app/Info.plist$")
	if infoPlist == nil {
		return fmt.Errorf("找不到info.plist")
	}
	pi, err := infoPlist.Open()
	if err != nil {
		return fmt.Errorf("无法打开info.plist %s", err)
	}
	defer pi.Close()
	buf, err := io.ReadAll(pi)
	if err != nil {
		return fmt.Errorf("无法读取info.plist %s", err)
	}
	_, err = plist.Unmarshal(buf, data)
	return err
}

// 查找ipa文件
func findZipFile(files []*zip.File, reg string) *zip.File {
	r, err := regexp.Compile(reg)
	if err != nil {
		return nil
	}
	for _, file := range files {
		//log.Println(file.Name)
		if r.MatchString(file.Name) {
			return file
		}
	}
	return nil
}
func (p *IpaParser) Close() error {
	return p.file.Close()
}
func NewIpaParser(filename string) (*IpaParser, error) {
	ipa := &IpaParser{filename: filename}
	return ipa, ipa.parser()
}
func (p *IpaParser) parser() (err error) {
	p.file, err = os.OpenFile(p.filename, os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	var fileSize int64
	if s, e := p.file.Stat(); e != nil {
		return e
	} else {
		fileSize = s.Size()
	}
	ipaZip, err := zip.NewReader(p.file, fileSize)
	if err != nil {
		return fmt.Errorf("无法打开文件 :%s", err)
	}

	p.list = ipaZip.File
	p.info = &InfoPlist{}

	p.info.header = p.find("Payload/[^/]*\\.app/Info.plist$")
	if p.info.header == nil {
		return errors.New("找不到info.plist")
	}
	f, err := ipaZip.Open(p.info.header.Name)
	if err != nil {
		return errors.New("无法打开 info.plist")
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return errors.New("无法读取 info.plist")
	}
	p.info.data = b
	return nil
}
func (p *IpaParser) find(reg string) *zip.FileHeader {
	r, err := regexp.Compile(reg)
	if err != nil {
		return nil
	}
	for _, file := range p.list {
		//log.Println(file.Name)

		if r.MatchString(file.Name) {
			return &file.FileHeader
		}
	}
	return nil
}

// Decode 解码plist  https://pkg.go.dev/howett.net/plist#Unmarshal
func (i *InfoPlist) Decode(data any) error {
	_, err := plist.Unmarshal(i.data, data)
	return err
}
func (p *IpaParser) WriteInfo(data map[string]any) error {
	d := map[string]interface{}{}
	if err := p.info.Decode(&d); err != nil {
		return err
	}
	for k, v := range data {
		d[k] = v
	}
	zw := zip.NewWriter(p.file)
	defer zw.Close()
	w, err := zw.CreateHeader(p.info.header)

	if err != nil {
		return fmt.Errorf("创建文件失败 %s", err)
	}
	b, _ := plist.Marshal(d, 1)
	if _, err = w.Write(b); err != nil {
		return fmt.Errorf("无法写入文件 %s", err)
	}
	return nil
}
