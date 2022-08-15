package appleTools

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/syyongx/php2go"
	"github.com/tidwall/gjson"
	"github.com/xml520/wqutils/httpclient"
	"github.com/xml520/wqutils/ipaTools"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
	"time"
)

const uploadBaseUrl = "https://contentdelivery.itunes.apple.com/WebObjects/MZLabelService.woa/json/MZITunesProducerService"
const metaTmp = `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://apple.com/itunes/importer" version="software5.4">
  <software_assets apple_id="{{.AppID}}" bundle_short_version_string="{{.VersionString}}" bundle_version="{{.VersionCode}}" bundle_identifier="{{.BundleID}}" app_platform="ios">
    <asset type="bundle">
      <data_file>
        <size>{{.FileSize}}</size>
        <file_name>{{.BaseName}}</file_name>
        <checksum type="md5">{{.FileMd5}}</checksum>
      </data_file>
    </asset>
  </software_assets>
</package>`

var uploaderClient *httpclient.HttpClient
var metaTemp *template.Template

const (
	uploadUserAgent = "iTMSTransporter"
	uploadVersion   = "2.3.0"
)

func init() {
	uploaderClient = httpclient.NewHttpClient().Defaults(map[interface{}]interface{}{
		"Accept": jsonContentType,

		httpclient.OPT_USERAGENT: uploadUserAgent + "/" + uploadVersion,
		httpclient.OPT_AFTER_REQUEST_FUNC: func(res *httpclient.Response) error {
			if res == nil {
				return errors.New("请求错误")
			}
			if !res.ToJson("result.Success").Bool() {
				return errors.New(res.ToJson("result.Errors.0").String())
			}
			return nil
		},
		httpclient.OPT_TIMEOUT: 30,
	})
	metaTemp, _ = template.New("").Parse(metaTmp)
}

type UploadAuth struct {
	Account  string `json:"username"`
	Password string `json:"password"`
	Api      *Api
}
type Uploader struct {
	auth         *UploadAuth
	sessionId    string
	sharedSecret string
}
type UploadMeta struct {
	appid    string
	FileName string
	IpaMete
}
type IpaMete struct {
	AppID          string `plist:"-"`
	BaseName       string `plist:"-"`
	FileName       string `plist:"-"`
	FileSize       int64  `plist:"-"`
	VersionCode    string `plist:"CFBundleVersion""`
	VersionString  string `plist:"CFBundleShortVersionString""`
	BundleID       string `plist:"CFBundleIdentifier"`
	FileMd5        string `plist:"-"`
	metaBuf        bytes.Buffer
	newPackageName string
}

func (ipa *IpaMete) init() error {
	f, err := os.Stat(ipa.FileName)
	if err != nil {
		return err
	}
	ipa.FileSize = f.Size()
	ipa.FileMd5, err = php2go.Md5File(ipa.FileName)
	ipa.BaseName = filepath.Base(ipa.FileName)
	if err != nil {
		return fmt.Errorf("无法计算Md5 %s", err)
	}
	if err = ipaTools.ParserIpaInfo(ipa.FileName, ipa); err != nil {
		return err
	}
	metaTemp.Execute(&ipa.metaBuf, ipa)
	return nil
}
func (ipa *IpaMete) getGzBase64() string {
	var in bytes.Buffer
	w := gzip.NewWriter(&in)
	w.Write(ipa.metaBuf.Bytes())
	w.Flush()
	w.Close()
	return base64.StdEncoding.EncodeToString(in.Bytes())
}

// NewIpaUploader 创建ipa上传器
func NewIpaUploader(auth *UploadAuth) *Uploader {
	return &Uploader{auth: auth}
}
func (u *Uploader) Upload(appid, filename string) (err error) {
	var mete = &IpaMete{AppID: appid, FileName: filename}
	if err = mete.init(); err != nil {
		return err
	}
	if err = u.authSession(); err != nil {
		return fmt.Errorf("authSession :%s", err)
	}

	//if err = u.step1validateMeta(mete); err != nil {
	//	return fmt.Errorf("step1validateMeta :%s", err)
	//}
	if err = u.step2validateAssets(mete); err != nil {
		return fmt.Errorf("step2validateMeta :%s", err)
	}
	if err = u.step3clientChecksumCompleted(mete); err != nil {
		return fmt.Errorf("step3clientChecksumCompleted :%s", err)
	}
	if err = u.step4createReservationAndUploadFiles(mete); err != nil {
		return fmt.Errorf("step4createReservation :%s", err)
	}
	if err = u.step6uploadDoneWithArguments(mete); err != nil {
		return fmt.Errorf("step5uploadDoneWithArguments :%s", err)
	}
	fmt.Println("新包名", mete.newPackageName)
	return nil
}
func (u *Uploader) authSession() error {
	//authenticateForSession
	var p map[string]any
	if u.auth.Api == nil {
		p = map[string]any{
			"Password": u.auth.Password,
		}
	}
	res, err := u.authDo("authenticateForSession", u.defineMap(p))
	if err != nil {
		return err
	} else {
		u.sessionId = res.ToJson("result.SessionId").String()
		u.sharedSecret = res.ToJson("result.SharedSecret").String()
		log.Println(u.sessionId, u.sharedSecret)
	}
	return nil
}
func (u *Uploader) step1validateMeta(meta *IpaMete) error {
	res, err := u.do("validateMeta", u.defineMap(map[string]interface{}{
		"Files": []string{
			meta.BaseName, "metadata.xml",
		},
		"MetadataChecksum":   php2go.Md5(meta.metaBuf.String()),
		"MetadataCompressed": meta.getGzBase64(),
		"MetadataInfo": map[string]interface{}{
			"app_platform":                "ios",
			"apple_id":                    meta.AppID,
			"asset_types":                 []string{"bundle"},
			"bundle_identifier":           meta.BundleID,
			"bundle_short_version_string": meta.VersionString,
			"bundle_version":              meta.VersionCode,
			"packageVersion":              "software5.4",
		},
		"PackageName": "app.itmsp",
		"PackageSize": meta.FileSize + int64(meta.metaBuf.Len()),
	}))
	if err != nil {
		return err
	}
	fmt.Println(res.ToString())

	meta.newPackageName = res.ToJson("result.NewPackageName").String()
	return err
}
func (u *Uploader) step2validateAssets(meta *IpaMete) error {
	res, err := u.do("validateAssets", u.defineMap(map[string]interface{}{
		"Files": []string{
			meta.BaseName, "metadata.xml",
		},
		"MetadataChecksum":   php2go.Md5(meta.metaBuf.String()),
		"MetadataCompressed": meta.getGzBase64(),
		"MetadataInfo": map[string]interface{}{
			"app_platform":                "ios",
			"apple_id":                    meta.AppID,
			"asset_types":                 []string{"bundle"},
			"bundle_identifier":           meta.BundleID,
			"bundle_short_version_string": meta.VersionString,
			"bundle_version":              meta.VersionCode,
			"packageVersion":              "software5.4",
		},
		"PackageName": "app.itmsp",
		"PackageSize": meta.FileSize + int64(meta.metaBuf.Len()),
	}))
	if err != nil {
		return err
	}
	fmt.Println(res.ToString())

	meta.newPackageName = res.ToJson("result.NewPackageName").String()
	return err
}
func (u *Uploader) step3clientChecksumCompleted(meta *IpaMete) error {
	_, err := u.do("clientChecksumCompleted", u.defineMap(map[string]any{
		"NewPackageName": meta.newPackageName,
	}))
	return err
}
func (u *Uploader) step4createReservationAndUploadFiles(meta *IpaMete) error {
	var f *os.File
	defer func() {
		if f != nil {
			defer f.Close()
		}
	}()
	res, err := u.do("createReservation", u.defineMap(map[string]any{
		"NewPackageName": meta.newPackageName,
		"fileDescriptions": []interface{}{
			map[string]interface{}{
				"checksum":          php2go.Md5(meta.metaBuf.String()),
				"checksumAlgorithm": "MD5",
				"contentType":       "application/xml",
				"fileName":          "metadata.xml",
				"fileSize":          meta.metaBuf.Len(),
			},
			map[string]interface{}{
				"checksum":          meta.FileMd5,
				"checksumAlgorithm": "MD5",
				"contentType":       "application/octet-stream",
				"fileName":          meta.BaseName,
				"fileSize":          meta.FileSize,
				"uti":               "com.apple.ipa",
			},
		},
	}))
	if err != nil {
		return err
	}
	for _, item := range res.ToJson("result.Reservations").Array() {
		fmt.Println("正在上传", item.Get("file").String())

		switch item.Get("file").String() {
		case "metadata.xml":
			if err = u.step5commitReservation(meta, bytes.NewReader(meta.metaBuf.Bytes()), item); err != nil {
				return fmt.Errorf("上传 metadata.xml 失败 %s", err)
			}
		default:
			f, err = os.Open(meta.FileName)
			if err != nil {
				return fmt.Errorf("无法读取文件 %s", err)
			}
			if err = u.step5commitReservation(meta, f, item); err != nil {
				return fmt.Errorf("上传 ipa文件 失败 %s", err)
			}
		}
		fmt.Println("上传完成", item.Get("file").String())
	}
	return err
}
func (u *Uploader) step5commitReservation(meta *IpaMete, at io.ReaderAt, result gjson.Result) error {
	for _, item := range result.Get("operations").Array() {
		fmt.Println("全部分片 ", len(result.Get("operations").Array()), " 当前上传 ", item.Get("partNumber").Int())
		fmt.Println("url", item.Get("uri").String())
		fmt.Println("Content-Type", item.Get("headers.Content-Type").String())
		var data = make([]byte, item.Get("length").Int())
		n, _ := at.ReadAt(data, item.Get("offset").Int())
		res, err := httpclient.WithHeaders(map[string]string{
			"Content-Type": item.Get("headers.Content-Type").String(),
			"Content-Size": strconv.FormatInt(item.Get("length").Int(), 10),
		}).Put(item.Get("uri").String(), bytes.NewBuffer(data[:n]))
		if err != nil {
			fmt.Println("error:", err)
			return err
		}
		if res.StatusCode != 200 {
			return fmt.Errorf("上传失败 状态码 %v", res.StatusCode)
		}
	}
	_, err := u.do("commitReservation", u.defineMap(map[string]any{
		"NewPackageName": meta.newPackageName,
		"reservations":   []string{result.Get("id").String()},
	}))
	return err
}
func (u *Uploader) step6uploadDoneWithArguments(meta *IpaMete) error {
	_, err := u.do("uploadDoneWithArguments", u.defineMap(map[string]any{
		"NewPackageName": meta.newPackageName,
		"FileSizeInfo": map[string]any{
			"['" + meta.BaseName + "']": meta.FileSize,
			"metadata.xml":              meta.metaBuf.Len(),
		},
		"ClientChecksumInfo": []interface{}{
			map[string]interface{}{
				"CalculatedChecksum": meta.FileMd5,
				"CalculationTime":    100,
				"FileLastModified":   (time.Now().Unix() - 10000) * 1000,
				"Filename":           meta.BaseName,
				"fileSize":           meta.FileMd5,
			},
		},
		"StatisticsArray":        []string{},
		"StreamingInfoList":      []string{},
		"PackagePathWithoutBase": nil,
		"TransferTime":           300,
		"NumberBytesTransferred": meta.FileSize + int64(meta.metaBuf.Len()),
	}))
	if err != nil {
		return err
	}
	return nil
}
func (u *Uploader) do(method string, data any) (*httpclient.Response, error) {
	var header = map[string]string{
		"x-session-version": "2",
		"x-request-id":      time.Now().Format("20060102150405") + "-000",
	}
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"id":      header["x-request-id"],
		"params":  data,
	}
	jsonByte, err := json.Marshal(body)
	if err != nil {
		return nil, errors.New("json 格式化失败")
	}
	if u.auth.Api != nil {
		//token, err := u.auth.Api.generateToken(300)
		//if err != nil {
		//	return nil, err
		//}
		//header["Authorization"] = "Bearer " + token
	}
	if u.sessionId != "" {
		h := md5.New()
		io.WriteString(h, u.sessionId)
		for _, b2 := range md5.Sum(jsonByte) {
			h.Write([]byte{b2})
		}
		io.WriteString(h, header["x-request-id"])
		io.WriteString(h, u.sharedSecret)
		header["x-session-digest"] = hex.EncodeToString(h.Sum(nil))
		header["x-session-id"] = u.sessionId
	}

	return uploaderClient.WithHeaders(header).PostJson(uploadBaseUrl, string(jsonByte))
}
func (u *Uploader) authDo(method string, data any) (*httpclient.Response, error) {
	id := time.Now().Format("20060102150405") + "-000"
	var header map[string]string
	//log.Println(data)
	if u.auth.Api != nil {
		token, err := u.auth.Api.generateToken(300)
		if err != nil {
			return nil, err
		}
		header = map[string]string{
			"Authorization": "Bearer " + token,
		}
	}

	return uploaderClient.WithHeaders(header).PostJson("https://contentdelivery.itunes.apple.com/WebObjects/MZLabelService.woa/json/MZITunesProducerService", map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"id":      id,
		"params":  data,
	})
}
func (u *Uploader) defineMap(m map[string]any) any {

	metaMap := map[string]any{
		"Application":         uploadUserAgent,
		"BaseVersion":         uploadVersion,
		"iTMSTransporterMode": "upload",
		"StreamingInfoList":   []string{},
		//"Password":            u.auth.Password,
		"Version":   uploadVersion,
		"Transport": "HTTP",
	}
	if u.auth.Api == nil {
		metaMap["Username"] = u.auth.Account
	}
	for k, v := range m {
		metaMap[k] = v
	}
	return metaMap
}
