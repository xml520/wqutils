package appleTools

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

const sss = `dssid2=7f50e672-5355-4bb8-b7b2-95ddf419ea9c; dssf=1; POD=cn~zh; itspod=55; searchTermState=0; s_vi=[CS]v1|31A178D312411188-400009A1D070D137[CE]; s_fid=0AAB0E224C32074A-0CBCF67ABAACD887; geo=CN; s_cc=true; s_sq=%5B%5BB%5D%5D; dslang=CN-ZH; site=CHN; acn01=EZfj6rv7026cP7HvXo4JtoeOOYRDp1XeXgAXapsxM2ce; myacinfo=DAWTKNV323952cf8084a204fb20ab2508441a07d02d3654edc302a2f56428ee3af34868977c9ed69c9461e1200a0bdba89a67f13933c1be625a93f1afebd5fe8dc4f5806350c5a38412cbbb7822921194eacdeb3dac7486161b023b0789d35a8dd0431313050880fd659a17d4a9c9a298f835a3bbeaca42bd50d4f08cccaa86fb1eb30248cd37088ac131c66d293c438fe5963ee027d1e80c2cc1edd2d4e142f6004ee4a78857503678b66251d220768956f0fed718484505124031e08519bae36c6a56d296a42516ccf6b9b8b4cf91d87dcd89cdb98430b31a4cdf385998a036af22afa5250aced83d475164449d31639577d361fb4568efdd7bb03d7ad80154f051487a597dd953cf5933d5695b741d7e6ed9e63853862d5ce9f7c51bec91da53fe98f98fda84b53659d484f893a35c18881f0368f8d1f6a4d6da93b14ad1dbf246a1454c567ba4bc25895e942dd9f84d55108aaf0efff3c1d64761df33b66dd5f2ceac4a486d14e387c6034a6d1fcca4233dd2fce99cfcb0d755fef5ace8bebe93429e29a16753168d3000ca837bd95da6fda6e6d4cd901b3c32eb54f51e13ada6ee7e7667dd166f33305b1537f69a9f81dbb90e3251a131ed952bd66127f463cfecf5f01fbea41faa1321de57e18e5f749b9d02b67ed8fe1242607715860eeffca4ab0126ec41c41e58e6322b3ebad3458f3250c298637b2d0537c49a9cb8afd6fb188d9585a47V3; dc=mr; itcdq=0; itctx=eyJjcCI6MTIyODc4NTExLCJkcyI6MTAwNTg3MzYyNDEsImV4IjoiMjAyMi0xMC0yNSAxMzo1NjozNCJ9|lbcibketgn8bhf00thjus83l88|t9TbZy4_HGrMVbb_rM_LxZF5U0g; dqsid=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE2NTg5NzEwMDgsImp0aSI6Ik5mdUFabkV6VjJKZU1oQ2ZISTFwcFEifQ.-oT4Pb6T8zQwoOirBNDyNkT3Qydgk6ho9wOSuRcjOzM; wosid=qkA2vLiCqnVKsMwiR4ET0w; woinst=5381`
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

// RandString 生成随机字符串
func RandString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
func TestWeb(t *testing.T) {
	var c string
	c = sss
	var w = &Auth{Web: Web{Cookie: &c}}
	var l = make(chan int, 1)
	var success int32
	var errorCount int32
	var i int
	go func() {
		for {
			i++
			l <- i
			go func(index int) {
				defer func() {
					<-l
				}()
				if _, err := w.Do("GET", "https://appstoreconnect.apple.com/testflight/v2/providers/122878511/apps/6443429052/builds", ""); err != nil {
					fmt.Println("index", index, err)
					atomic.AddInt32(&errorCount, 1)
				} else {
					atomic.AddInt32(&success, 1)
					fmt.Println("index", index, "ok")
				}
				fmt.Println("success", success, "error", errorCount)
				time.Sleep(time.Second * 5)

			}(i)
		}
	}()
	time.Sleep(time.Second * 6066)

}
