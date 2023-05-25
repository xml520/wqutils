// Copyright 2014-2019 Liu Dong <ddliuhb@gmail.com>.
// Licensed under the MIT license.

// Test httpclient with httpbin(http://httpbin.org)
package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

// common response format on httpbin.org
type ResponseInfo struct {
	Gzipped   bool              `json:"gzipped"`
	Method    string            `json:"method"`
	Origin    string            `json:"origin"`
	Useragent string            `json:"user-agent"` // http://httpbin.org/user-agent
	Form      map[string]string `json:"form"`
	Files     map[string]string `json:"files"`
	Headers   map[string]string `json:"headers"`
	Cookies   map[string]string `json:"cookies"`
}

var _data = `{
  "kbSync" : "AAQAALVGHOdf76eSEhLq1BPxYtl+8pYOBxTEwWl4PMl2rFQkZX2MBBcPcrwTLeqWD9gQpGXogSMkHKf+3xEay6SCYaNL3Saq7iCRF1xn9kYIN+5yp7dP7g\/COfT0cUlF15xfmnjse3P8bdPkXW8eR8vcFRYlRJWAsPqJBxB4bhuRguaWrvc8iqvrZ\/e3kyA5pdqyQG7hruXlD785ce9+x\/uAgtDEhIr7nqChFuiao+qWw7oQUx8DAaRjdv1m\/Hv5cPw7FHgUuncRImfK2WdwZ3cQpF4=",
  "storefrontId" : "143465-19,29",
  "purchaseKbSync" : "AAQAAGSBrPsbJbeO4NTSOVF6oQXIMRVDjwPJVL8HDl+3ZuJ4XXbvCnFVoU1CQqpzl\/S3VRaPVmsDUiWqa1+FOTctk+Cpf\/ve6QU8Kg9JvrDnkWdTjZnEsnOqEJDCmVIZ7Q\/N6ffBqfwXnUd1xDjzT\/GFQYsvRuXSqsUBfAU0qAJRt9fPwUNB25BYXgACU3zsR\/A0WnWW5JZeddR9uoXDqAzjHwfp1iLCxqAgqKUV2AjJizGvU4DLkwKFaIN\/nnTQLs95klBD4N\/KUIeusHGYVqDSKaX\/zlPXGyzTbFFhETM+XQKw\/rlArMubr6DL6rZs2A87ddAi2WLhm\/okMz0yjxloRwads1\/yNVzeGKRIUr4PeaK5Uma7chu9kgYy0ZU5DBuvPkBXUX4\/5ZAFA\/McS33VQh96mhR31ekgGCGeLzRdpPfaDu0yFhD0l\/P\/eIJB42xv1W5wIX+KM4sCOwXJG1LrH0WOOuPNHcIAS4+ydeTaUOkJsFFr0HY5gGNr0nZ48yKChQ\/27T+oVJvmJlN7nZp3AMyE0xUipklaSHtkyFeV146R0szS2T9WCFg8fMCqEVkuePqhwBX2taeDRMmDfWNwjbAqpE79NH1iAtBeoLqrW0VxHDsbi1ObQqJXnJPwA5WLU3opqf9TBa8PmjHqg4fPs8fC8h0rtNEemmeAVqQMtBZ6rTDpm+cyTOC\/kumhVGuEHzoUSc8PPI6yKhZf+wSctxlyVw4paqssItFrHswvBJUZ2dBsP+IDU8yne4cdoVCxyzSRfiuovxz\/NeNnbY\/tOmRTOaOjdfpgkuyP4Mrgz2xDBu1pXeuXySsN\/nKliddJEnDPH87eAOn\/2E9tatM5YpyJ3ezWks1vUemjUIo1WWmY43rllcr8yVlRztcW1d57Mq6fpqLvYuq3nj9AQE7y7w1EqYEZ+j76UJ7rQ8jf5ZAjwTr4FoLADIbB5Rb7zXiPTMUuvTv4DirVPMaoJXr+0XIrB2WO\/a7KkIFl2JbetcVXk4Jso+xpW\/zMoPDRiay5H9\/XiFUlYLCfrxFEvpHTcwXYdBDmnGtYkOh64S3n2tcDgA4ECUWUQtS2RG3MOv9uoqT3teLlEk3Y4wSNtNOVqDoesnR8OLcwJzix9MD5bRBA9xS7UXLUvkXWeohmpHu5WVACNrq+Z2v8yGowjy9G2u8K3QN1PJVPAr6AxiEYxr8EO9NcJEbFX80cpGTBZ6pa41EXpD1eIMo\/wHKcOAIFGbjX4R0qWGPgfp4WGC3JfSerMDk565ll37Nhc4GZtkjrBpLCL3sQUuVwmQqZqjJP0CjAMnUtCz7B1BnME\/PHovmG9bnlRHQjo7HVxtV7bZ0G11U6R3Rv9GuTKMZkUlYlRWiY4CZenUgKafoLvpk5POECLHxgK4DmkLGeiHQOp2s9LveXEiwJp1nT0Quobxuv8JbT2MXScTdylqPARAUSZ6HL4bjcWWblp5Hzen7XTWM4fpE5uSA0ahw2+qL+cTjREhGUtml4UIingcDSGaC6Xb7xPPJ1RlUQeW3vpWzQ4TBYsaDNlsDDRG4ysn7nDOjzl09ZHuDAoK+fGp1rPlal611S07PGsRw3a\/FFc9kSQGhL4RjlTFfpKxk4kvUBCVGKTUNv4plkpUUVYV9BLvpIbdpTnVW8XXH\/3puqgG1deGJNbD2Y5F9LAkQWtDzI+PMMyK4NbzU7BSu0MtWiHptspA+IeOJV5XqQWI5xDm1pGKkWo1dNenUzHwc9EVJaWwQPsVRV8bEcdBswBH2VblCyrWxiTRkeYPmcXgkxCn3Uu6WI8IFRaIQwHQ\/s3BGMEP7jaEnvt+JeDX7OZ49aNO4aEuGa3DEZwavoUk6fY8zhda1OH+fpsLoBzlZGeQS6krt6gIEZ+1F6ZrBlrGKDAndzAfSL7hRRwq4ozvGo9TI+5VsREwyQea7z\/svo17LdOEcWvKUEtmfmkVxAxwV9isFjnsq8q05sNNvtcBkFiOouorYZM0MSBnYmvH9ZBpTZvX0B4UC05J2Is8NmY+Ye9iVTKB9Xy\/jh7ZXgPVTXRMZtDVJ\/JaHqsb7NeXbM8Ii3G4cthvu+F4U1hv5xHVc0Y+XOCGhB0i+Wv7CdiDTi1AvVI7vDm5XaicgT9A\/3m7IT7bnCl7cYj\/2ObtljvslUZCdZEyidD62Q4+vTXunIcB6qxkzCjeKqya8s+BkEUM+8dN5J2eA50CscbrjO7tOf\/VYdhmxMSQ\/cGUkaFuhB0Pb8PoNPDmSLTxi4ruJY6\/KzusNN5F4Qltes7DvI0LxkMd3CQgJZ1OK04xwT5og+tjMjyFH3IdXjM8VVVZZn6WFs6bQcpjiQn+nTn7FtbUv2\/V3q4Q+34OXZoKchOo0C1xnLPN7CVrzm0x4u25wyc+DizFG0z2R91sGH9vFhUKd\/UkpiPKjbDmaiSV5PGPob0l+2n6euqtuHrbKy3ZPt49bSXG3W\/sv3Xk15q822XQwgi3zskK9h86mUnrko5+3UdI1Pz6DhEHR8mdmy32QLwEJHUqdwTUPObkOLDVwmukAXBzwAy8TCR1nE3RuOgI8Qo3ToJI9tWq4UHQ6poGCiApphiSoq21PpIo31gXwf8V5sABFSsOrcfAG\/e7eBBNUTzk2jKLHhXePCd2CA8\/0BIWfvgo\/qnzVbpyOmK9N2dKZLs71rP6plrpZlNv578KnmGfvhAvFESSJSBaY4YhBBGQMyKOzah7K3LJk2mFW4FxiwKVDLvRCbkMNUTBvDV4ZvCdzo\/39u0pP5Y732Udfr2tfbK6gjKEyAQ53pNg8A1rXsWVy8F6w4b+RSa4FtG7Tu7lmVZnlt4v\/8n2DaGPxhSWWh65fjg0Sn+t14Z55t71\/wTcbkDVRo1izxG5rjnbWoA+lVQ6obIbH1FfLM\/ZzNLcasGRffOylKlDipOKb7IQ1cWWGAFNJ46XN+GVEzdopGrszY3FrKIBvApWaZ3oVPQG3GsN997t4GnzrC2xEwD74F4xvXM0LI0tXypIAfYgBhg3d8\/gB+a95\/nkXaVYir62NuXHnxeJ6XMC7k7\/LaXkI+FbeOrGL6+v1TF\/GnTQZd1NUHUUffpUHItYwCvK1Rj4k6e4N1DQDnMWARrNFwDGlxJN5g3RoOTRHD47p9WmzEzSGh6F+piH9PUdBYo\/kIdAlghG+WvHmtQsjm0qGFtuATy75SH2IyFgPWoMp0hWk0TYXTpHZ6KDXHXz0YUlmdnt1BDCeDxfg3z8mRYlopwnCVQY31xnvkN6oLndLt2ID6IkBjhUnW0BAg2GQi30QzifqJgw1d0oUuq3XqrUQgoUF+udUTf7UFOzW0aznplKq2eg\/qRHITsJUuRraqW1DTpm8gp6JrHpX4fcORKN+M6m6ToXQdmueuVNgsnlQUTeYGmLA0DsdWxP4wsQxqMB41bRd1+mSnozA\/vdaW57iMfWwIT9L7xH8zZIX7kC+oiK8Yolh9phSubVKY928NyPMgxM9KwAzxhXTwXA+o49B2eIEX7jZ1vcq8WWxFleJqm\/HuyVlC4s7VEfe8rVIE\/ryPg7iR58maqLuunwOiGs1RCW5mHRrs\/PU3\/JONKxuKcorG90CpOv0b02\/PWmD6al46OYkCkuoCHFy4gh6ydU8+cCnyuTAGofN2zGuSTFu9Vxgl6XjPgyjipd\/oek1BYxZ9xzAjlD06jW5ubwEUtljV2ir\/qx1ChppOpV0Sy+8jaGf9gPKl0Dd7Kvsc\/ekOqKU1yiHNON0VstJ95O7BbvtxviE2mAGUqTvg\/oGeF4B7FpEMPjrCN2i6xJn5rgyTD4BykFOWhRjX7OOJ4yHwLM6UsXh1A5mO\/l2Wm6nnDgOz9kzGW4B7dr4E6Hm942h1k7CEX2seoHydz+DjRzIMLgM2j1kK6vY++3DmJyBs6\/1po5wg8aRzhjCiAzQMMS2xGYwmsy7dPs6\/XoEJnzw3ogPJADmt+e26lwF+OHThPb2ClELDO4usfth94UHc\/TqO\/fzSQcKe03aTAiwO5b10BOfxN3MoH3M5PXtxtnMEezJlp8IGZ67C0AxQQqyn4yCd1lH7wX7QOq+CLmMHefjfHqvigHIG1HSG60rtGyV9cVL9OvhiidzA7zPbEvPzmhBo8q83k2zPWaV+SUC9rlQ86b5eL36cXBpFTqsN4ZpI1Drl\/GebM7Q\/uJmxETLeuwHSeWNoRLMREg0O9D9vk8hk6CvWe2L+NW+qFJSPBOMPp7nY5+ECeeTp9APpNEhol5C\/rwM1I2hmyavPfFGg91VE7j836uTgyYkXWa8QmXzy\/2fk\/vAjyQdrb1mHB\/g9LQoLTS5N5LGtMSk+2DcVPyT0FWQE8ld3qMN4g5Ut72NF+8spI3MypJ3J41c50QETLCFsBHldmXL+trJUAh+K6f9tOovZqBokdkm64BbfDO6i7lRCSfvAfZz8\/j\/\/D4MSBBbHi3RUYLmxukh58WBe0zExtUxso8LhXguzT4qHsIKjRABP+IuQ7Qp7nLe6ZvyCtdO4uD\/vQxX5CVLc03iSxgnVJGPgf1BhjbrTLKGTheWi1Td8mMnrKWqx82CPnFywQ1i8VZBrG+1xHe8BHKmRq8dOCLOJp0VcD2rmeUVDePrGf\/k0qz1Bjo03Wd2u4GoSwVoAKIJFqVaDgn8Mv+RimrhuheI3usLFKaJxPbwSzgftblxYsGDrSF3gOUR6C8d4zWZTosh3z2TicyuZfzeZUW536jo72CrhYyyiYZSejvJRVHAN74OckWsTcyrX6tJw\/l7IDMj5v7hiQP04i77pWRHiwtzv\/6YdIF6zVFf+DxVECKixKyrHZkn+RM7f93wIg9N6a\/0AJ\/8qR1oXoo8uL4fum23x7gBKnoZ1JLVbFx4OKfsTkHWeMupL8hiEgxL00KP\/SpbU2Yii6vmwSdHvB7BmQeE70vlD3ZaTWhq75V5GdR065qRhh6Sz1pM3ScP+gEvuTauF4Kg87AI4UTJW92YUvSnjkYh9K9VBmnzzlYBNmqI9+Cu5NNGKBhxPsoBm+0lrirl6p0Bj0LaHkzduPxHlhTPcRtFgyTNN00C0KucN0ihjK6GtmtC7F1FHV3kkWGNwdaevl7n4PTIPsqJ35qX5X4taWiJkFoZzQ29XAQX\/gr61FIoX+r9hroSS7Gcpt\/k+9210WS+IUkKoMldfzFn6hSlUovnpoBdD3ZTVB\/6141V8bv9yhwKEzfo\/lKsW+qIhnRGYL0xocWWU7Rb6kaYlWQuMh0CorSUR03JtoX7qh1eUJQ3KnGlGPv5Aw6dubtGhC5MKaWnf4rWOwKNvlgKs\/PXua3FSArRE9UTs9PrRozrJFCQ8us5x4nfTyh0pRhNUs4vOJINELVvgLrXhlI63E3NOhhCE7mEnnrWMvwRYowxFVy3dtq3h7p4qnFgVdLRBPkyj+NHt1poEUAEpaTYG8Kc52+0H\/\/LeLCDgz+ax0FVwTEzJrosdzvESRxskTUVFFSRyOHTgUnwqSLMD2WbfOG\/VS7fS70psTv8txie\/\/\/RMhtQ7UZJc5UQ6uOiW4S07osAWmG5tjS9k56rxYmFbLDc5pGNxnHG+l7UDn\/MSr38eyf\/YCC2E1Fu8VfKlqf5gaPh1avDTv6QC5PToM4RY5POO41vJ+XeFQBa1iToxZtMYvjuJMCWpyvKmw5e+q9\/KeNsyqS\/KPFPXq7fzlFY8QjTPkr\/aZP+NnPJd30EaKXJPy3BCfvs\/b3zGLE\/pNcXUIUktKuThJ5aLKfZ5PooUyLPNmcQA9sCaVw5szvg8NGAkE2mAV0mVQz0CCjSsxqiHGsbKQOABDq2vCwGS00dtpWVS2DIE\/gjc15I5WGo0EiBAtgBfmxUcy7Ve78j7+nBI4trrCeVDkGWQRGIRuwI6wg+F9cEiZgPKvzQ1eORPjxGT\/N9nu6lI1nhlzJMXBtzru7MTx5LN+VhkXOZ9Yn95lYKXIcRsap2UgUzTXj+Qpau\/WFHrpvP0HSzElX\/c7eK4mqEIrv55BWlzj9GV2VN9oVo+c1JIKuEt\/VBDFizC0iVHFprehJsdpVsD++kGZ4rRqh3cNUrRKcHzsQS0rILEBIlYPGGxHvQGjcAmbD\/EdJusTM6t3cNV1yiRhpF\/Ix0AJYkNDSUevfGxLvVK8lERabJRGaG7nuVMOQZrEtSa37E9Jq5ZKPU90lIQcyEDMEbTsWlF\/UAT7yukdgflD7ulP1jnHQR6wRU6+PCt+mv39rIStwartZWkPVS\/TUXdRyKhI4Y55U8R5KBv1JwNE4bvZVZGlc7b4oe37hurMZEMZKjCtw4aUww2UECmdP3bWAPpbcELh2cDi5AwswpOn1rR3A2MdXMTeFLqtBSk8BJt+kG6a6DqzF9oyMcfH68ZjdjPeWRpaxePemN\/IuPOummCvHSPbyDELWuhYtjSD6rhiQOZ9KTOmIJUlfn3ztGPBn\/SVBc6o6GZbMqeXXIiCHdtXUJ7isB77XlyxXu2pUam0e1zPwu07EsyvebwwkO3+Ry3s6lWuRR6kOBObTgRuba9MfPwLDl4rZVnGnrsCnSSxFaIVvXaW\/pnO2ebZGk12xYqwqPZq8+tWq7O8I5uY43hYOsz7FjGn0keV5\/jhakLNFREPSeN5F1M\/hJoBSgq2tqJqeAXPqzA\/uzYVpO45k3z79+sdc1iriWncIG8MB2\/+EBb\/IJo0Dbc9HEgqRM3BLWM9w0+b7N343qFlgd9Yjs+YiSB6DNcLsfPuGYd8VVYf7X+34SZ3dtcFeQ77mmQNiE3emRHvoeaIDSl6Od4dnkjG2IFH9Um3HmBzyAJ8mggC+UCVGcUr8htYSVGquKTx0i5R5RHiNZMOgvZrGKorPNZsxwPLVLAe\/d2yRiIX5GQZyegtUr+EwL872wIIE0WOFS48Dx9bosaAQJiUHNbP9aW5A78ti+stuXbn5C+GYAzo\/nbIKTRXw0Hrj\/BnPkjKzBeWcylcTxds\/O0LlxkDm1OAArzTawQ1mToM044dK48tSbnsYvzDINdWHpQwvO0zpfF0Flq2Km8q7VO9YjYgCFUjmk4TbbdwNoFaDLpn76XpzNWwQ3TL+sxNgrxe0VO77\/NEt9HNiOKWfZQNhS53EsAsnXIIv\/0kwNhv2BSGp0LKp5H0UtGRYXijvxqYUBVb5PMKLGfxRd+TJjTqFMVHmH8yqOhaL+Iltdd40eK9qQw3SB7QbHY6jHc+FvdrphG\/C\/tpKh0E7kFWSd7jVhuSzCBPRqPCrtBIOsbBDZL3EZrgHulgrxhdCXZqjIl1ARq7AJgdGFMrvrSbVZvWxcgxb8BU6m2ekIGQhOSG83UJf0xclFd9Ze7hkdIKaDpL5N05foM7ajCqcfcFNMB\/gCV4TQmVFDlDU45o1RWTFRIG3QG12\/d08Jaz0G2rFulQX6xu4Kk\/DfZe4oLfEPj5wfpPYzcvs7wlvfstOyE4vqQtGmRn86Ymb1h7juWgEnQ7\/UlpN0uCvrW9cZ\/G6RzIsqa2zVm\/0YgCJjrlC+QZiqrGbdyMII4UD9g1wwUhakLznp2Cvbpw0F+J1OlUe2X8tXtJKF+8yAIulffbmJKbFLyQ2rSINXENc1t4IhmGtzAL4dFKn52PRjoHNDB7tBbwTuOuicJpYCHhT\/aDXCPUfX55p7g0mAfY7Ht0fHJCig5hCkGfwiBp+e1JWAuKrFhJXFadMRPejCxXKY5dPh1kmhGx6rTb419YuuoepmNNuA+ecJ3EocSYqc7eQiMsSuAKfmJjk5fWS95bVXy6pihVQZ1",
  "isFromAutoUpdate" : false,
  "udid" : "00008101-001119A42E00009E",
  "deviceName" : "iPhone"
}`

func TestRequestTF(t *testing.T) {
	var data = make(map[string]any)
	if err := json.Unmarshal([]byte(_data), &data); err != nil {
		log.Panicln(err)
	}

	for i := 0; i < 100; i++ {
		data["udid"] = fmt.Sprintf("00008101-001119A42E%05dE", i)
		fmt.Println(data["udid"])
		var body, _ = json.Marshal(data)
		http := NewHttpClient()
		http = http.WithHeaders(map[string]string{
			"Accept":            "application/json",
			"X-Session-Id":      "CL2sDBIQjtl+siJ9RfOly5e7Mci2Ig==",
			"X-Session-Digest":  "3532ad7b9a6163015253f95d369c2d681a3d290b",
			"X-Apple-TA-Device": " iPhone13,2 iPhone12,5",
			"Content-Type":      " application/json",
			"X-Request-Id":      "C27CB945-3A4C-44A6-9AB5-37514362ED23",
			"User-Agent":        "Oasis/3.2.3 OasisBuild/4 iOS/16.1.1 model/iPhone13,2 hwp/t8101 build/20B101 (6; dt:229) AMS/1 TSE/0",
		})
		res, err := http.Post("https://testflight.apple.com/v2/accounts/a83179c1-a9b1-495c-ab7c-d8a0b9aab671/apps/1670048808/builds/106904795/install", body)
		if err != nil {
			fmt.Println("res err", err)
		} else {
			if res.StatusCode == 200 {
				fmt.Println("ok", i)
			} else {
				fmt.Println("StatusCode", res.StatusCode)
			}
		}
	}

}

func TestRequest(t *testing.T) {
	// get
	res, err := NewHttpClient().
		Get("http://httpbin.org/get")

	if err != nil {
		t.Error("get failed", err)
	}

	if res.StatusCode != 200 {
		t.Error("Status Code not 200")
	}

	// post
	res, err = NewHttpClient().
		Post("http://httpbin.org/post", map[string]string{
			"username": "dong",
			"password": "******",
		})

	if err != nil {
		t.Error("post failed", err)
	}

	if res.StatusCode != 200 {
		t.Error("Status Code not 200")
	}

	body, err := res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	var info ResponseInfo

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if username, ok := info.Form["username"]; !ok || username != "dong" {
		t.Error("form data is not set properly")
	}

	// post, multipart
	res, err = NewHttpClient().
		Post("http://httpbin.org/post", map[string]string{
			"message": "Hello world!",
			"@image":  "README.md",
		})

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Status Code is not 200")
	}

	body, err = res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	image, ok := info.Files["image"]
	if !ok {
		t.Error("file not uploaded")
	}

	imageContent, err := ioutil.ReadFile("README.md")
	if err != nil {
		t.Error(err)
	}

	if string(imageContent) != image {
		t.Error("file is not uploaded properly")
	}
}

func TestResponse(t *testing.T) {
	c := NewHttpClient()
	res, err := c.
		Get("http://httpbin.org/user-agent")

	if err != nil {
		t.Error(err)
	}

	// read with ioutil
	defer res.Body.Close()
	body1, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Error(err)
	}

	res, err = c.
		Get("http://httpbin.org/user-agent")

	if err != nil {
		t.Error(err)
	}

	body2, err := res.ReadAll()

	res, err = c.
		Get("http://httpbin.org/user-agent")

	if err != nil {
		t.Error(err)
	}

	body3 := res.ToString()

	if err != nil {
		t.Error(err)
	}
	if string(body1) != string(body2) || string(body1) != body3 {
		t.Error("Error response body")
	}
}

func TestHead(t *testing.T) {
	c := NewHttpClient().Defaults(Map{
		OPT_CONNECTTIMEOUT: 30,
	})
	res, err := c.WithOption(OPT_TIMEOUT, 40).WithOption(OPT_SELECT_IP, "0.0.0.0").Head("http://httpbin.org/get")
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Status code is not 200")
	}

	if body := res.ToString(); err != nil || body != "" {
		t.Error("HEAD should not get body")
	}

}

func TestDelete(t *testing.T) {
	c := NewHttpClient()
	res, err := c.Delete("http://httpbin.org/delete")
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Status code is not 200")
	}
}

func TestOptions(t *testing.T) {
	c := NewHttpClient()
	res, err := c.Options("http://httpbin.org")
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Errorf("Status code is not 200: %d", res.StatusCode)
	}
}

func TestPatch(t *testing.T) {
	c := NewHttpClient()
	res, err := c.Patch("http://httpbin.org/patch")
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Errorf("Status code is not 200: %d", res.StatusCode)
	}
}

func TestPostJson(t *testing.T) {
	c := NewHttpClient()
	type jsonDataType struct {
		Name string
	}

	jsonData := jsonDataType{
		Name: "httpclient",
	}

	res, err := c.PostJson("http://httpbin.org/post", jsonData)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Status code is not 200")
	}
}

func TestPostText(t *testing.T) {
	c := NewHttpClient()

	res, err := c.Post("http://httpbin.org/post", "hello")
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Status code is not 200")
	}
}

func TestPutJson(t *testing.T) {
	c := NewHttpClient()
	type jsonDataType struct {
		Name string
	}

	jsonData := jsonDataType{
		Name: "httpclient",
	}

	res, err := c.PutJson("http://httpbin.org/put", jsonData)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Status code is not 200")
	}
}

func TestPatchJson(t *testing.T) {
	c := NewHttpClient()
	type jsonDataType struct {
		Name string
	}

	jsonData := jsonDataType{
		Name: "httpclient",
	}

	res, err := c.PatchJson("http://httpbin.org/patch", jsonData)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Status code is not 200")
	}
}

func TestHeaders(t *testing.T) {
	// set referer in options
	res, err := NewHttpClient().
		WithHeader("header1", "value1").
		WithOption(OPT_REFERER, "http://google.com").
		Get("http://httpbin.org/get")

	if err != nil {
		t.Error(err)
	}

	var info ResponseInfo

	body, err := res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	referer, ok := info.Headers["Referer"]
	if !ok || referer != "http://google.com" {
		t.Error("referer is not set properly")
	}

	useragent, ok := info.Headers["User-Agent"]
	if !ok || useragent != USERAGENT {
		t.Error("useragent is not set properly")
	}

	value, ok := info.Headers["Header1"]
	if !ok || value != "value1" {
		t.Error("custom header is not set properly")
	}
}

func _TestProxy(t *testing.T) {
	proxy := "127.0.0.1:1080"

	res, err := NewHttpClient().
		WithOption(OPT_PROXY, proxy).
		Get("http://httpbin.org/get")

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("StatusCode is not 200")
	}

	res, err = NewHttpClient().
		WithOption(OPT_PROXY_FUNC, func(*http.Request) (int, string, error) {
			return PROXY_HTTP, proxy, nil
		}).
		Get("http://httpbin.org/get")

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("StatusCode is not 200")
	}
}

func TestTimeout(t *testing.T) {
	// connect timeout
	res, err := NewHttpClient().
		WithOption(OPT_CONNECTTIMEOUT_MS, 1).
		Get("http://httpbin.org/get")

	if err == nil {
		t.Error("OPT_CONNECTTIMEOUT_MS does not work")
	}

	if !IsTimeoutError(err) {
		t.Error("Maybe it's not a timeout error?", err)
	}

	res, err = NewHttpClient().
		WithOption(OPT_CONNECTTIMEOUT, time.Millisecond).
		Get("http://httpbin.org/get")

	if err == nil {
		t.Error("OPT_CONNECTTIMEOUT (time.Duration) does not work")
	}

	if !IsTimeoutError(err) {
		t.Error("Maybe it's not a timeout error?", err)
	}

	// timeout
	res, err = NewHttpClient().
		WithOption(OPT_TIMEOUT, 3).
		Get("http://httpbin.org/delay/3")

	if err == nil {
		t.Error("OPT_TIMEOUT does not work")
	}

	if !strings.Contains(err.Error(), "timeout") {
		t.Error("Maybe it's not a timeout error?", err)
	}

	res, err = NewHttpClient().
		WithOption(OPT_TIMEOUT, 3*time.Second).
		Get("http://httpbin.org/delay/3")

	if err == nil {
		t.Error("OPT_TIMEOUT (time.Duration) does not work")
	}

	if !strings.Contains(err.Error(), "timeout") {
		t.Error("Maybe it's not a timeout error?", err)
	}

	// no timeout
	res, err = NewHttpClient().
		WithOption(OPT_TIMEOUT, 100).
		Get("http://httpbin.org/delay/3")

	if err != nil {
		t.Error("OPT_TIMEOUT does not work properly")
	}

	if res.StatusCode != 200 {
		t.Error("StatusCode is not 200")
	}
}

// Disabled because of the redirection issue of httpbin: https://github.com/postmanlabs/httpbin/issues/617
func _TestRedirect(t *testing.T) {
	c := NewHttpClient().Defaults(Map{
		OPT_USERAGENT: "test redirect",
	})
	// follow locatioin
	res, err := c.
		WithOptions(Map{
			OPT_FOLLOWLOCATION: true,
			OPT_MAXREDIRS:      10,
		}).
		Get("http://httpbin.org/redirect/3")

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 || res.Request.URL.String() != "http://httpbin.org/get" {
		t.Error("Redirect failed")
	}

	// should keep useragent
	var info ResponseInfo

	body, err := res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if useragent, ok := info.Headers["User-Agent"]; !ok || useragent != "test redirect" {
		t.Error("Useragent is not passed through")
	}

	// no follow
	res, err = c.
		WithOption(OPT_FOLLOWLOCATION, false).
		Get("http://httpbin.org/relative-redirect/3")

	if err == nil {
		t.Error("Must not follow location")
	}

	if !strings.Contains(err.Error(), "redirect not allowed") {
		t.Error(err)
	}

	if res.StatusCode != 302 || res.Header.Get("Location") != "/relative-redirect/2" {
		t.Error("Redirect failed: ", res.StatusCode, res.Header.Get("Location"))
	}

	// maxredirs
	res, err = c.
		WithOption(OPT_MAXREDIRS, 2).
		Get("http://httpbin.org/relative-redirect/3")

	if err == nil {
		t.Error("Must not follow through")
	}

	if !IsRedirectError(err) {
		t.Error("Not a redirect error", err)
	}

	if !strings.Contains(err.Error(), "stopped after 2 redirects") {
		t.Error(err)
	}

	if res.StatusCode != 302 || res.Header.Get("Location") != "/relative-redirect/1" {
		t.Error("OPT_MAXREDIRS does not work properly")
	}

	// custom redirect policy
	res, err = c.
		WithOption(OPT_REDIRECT_POLICY, func(req *http.Request, via []*http.Request) error {
			if req.URL.String() == "http://httpbin.org/relative-redirect/1" {
				return fmt.Errorf("should stop here")
			}

			return nil
		}).
		Get("http://httpbin.org/relative-redirect/3")

	if err == nil {
		t.Error("Must not follow through")
	}

	if !strings.Contains(err.Error(), "should stop here") {
		t.Error(err)
	}

	if res.StatusCode != 302 || res.Header.Get("Location") != "/relative-redirect/1" {
		t.Error("OPT_REDIRECT_POLICY does not work properly")
	}
}

func TestCookie(t *testing.T) {
	c := NewHttpClient()

	res, err := c.
		WithCookie(&http.Cookie{
			Name:  "username",
			Value: "dong",
		}).
		Get("http://httpbin.org/cookies")

	if err != nil {
		t.Error(err)
	}

	body, err := res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	var info ResponseInfo

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if username, ok := info.Cookies["username"]; !ok || username != "dong" {
		t.Error("cookie is not set properly")
	}

	if c.CookieValue("http://httpbin.org/cookies", "username") != "dong" {
		t.Error("cookie is not set properly")
	}

	// get old cookie
	res, err = c.
		Get("http://httpbin.org/cookies", nil)

	if err != nil {
		t.Error(err)
	}

	body, err = res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if username, ok := info.Cookies["username"]; !ok || username != "dong" {
		t.Error("cookie lost")
	}

	if c.CookieValue("http://httpbin.org/cookies", "username") != "dong" {
		t.Error("cookie lost")
	}

	// update cookie
	res, err = c.
		WithCookie(&http.Cookie{
			Name:  "username",
			Value: "octcat",
		}).
		Get("http://httpbin.org/cookies")

	if err != nil {
		t.Error(err)
	}

	body, err = res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if username, ok := info.Cookies["username"]; !ok || username != "octcat" {
		t.Error("cookie update failed")
	}

	if c.CookieValue("http://httpbin.org/cookies", "username") != "octcat" {
		t.Error("cookie update failed")
	}
}

func TestGzip(t *testing.T) {
	c := NewHttpClient()
	res, err := c.
		WithHeader("Accept-Encoding", "gzip, deflate").
		Get("http://httpbin.org/gzip")

	if err != nil {
		t.Error(err)
	}

	body, err := res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	var info ResponseInfo

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if !info.Gzipped {
		t.Error("Parse gzip failed")
	}
}

func _TestCurrentUA(ch chan bool, t *testing.T, c *HttpClient, ua string) {
	res, err := c.
		Begin().
		WithOption(OPT_USERAGENT, ua).
		Get("http://httpbin.org/headers")

	if err != nil {
		t.Error(err)
	}

	body, err := res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	var info ResponseInfo
	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if resUA, ok := info.Headers["User-Agent"]; !ok || resUA != ua {
		t.Error("TestCurrentUA failed")
	}

	ch <- true
}

func TestConcurrent(t *testing.T) {
	total := 100
	chs := make([]chan bool, total)
	c := NewHttpClient()
	for i := 0; i < total; i++ {
		chs[i] = make(chan bool)
		go _TestCurrentUA(chs[i], t, c, fmt.Sprint("go-httpclient UA-", i))
	}

	for _, ch := range chs {
		<-ch
	}
}

func TestIssue10(t *testing.T) {
	var testString = "gpThzrynEC1MdenWgAILwvL2CYuNGO9RwtbH1NZJ1GE31ywFOCY%2BLCctUl86jBi8TccpdPI5ppZ%2Bgss%2BNjqGHg=="
	c := NewHttpClient()
	res, err := c.Post("http://httpbin.org/post", map[string]string{
		"a": "a",
		"b": "b",
		"c": testString,
		"d": "d",
	})

	if err != nil {
		t.Error(err)
	}

	body, err := res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	var info ResponseInfo

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if info.Form["c"] != testString {
		t.Error("error")
	}
}

func TestOptDebug(t *testing.T) {
	c := NewHttpClient()
	c.
		WithOption(OPT_DEBUG, true).
		Get("http://httpbin.org/get")
}

func TestUnsafeTLS(t *testing.T) {
	unsafeUrl := "https://expired.badssl.com/"
	c := NewHttpClient()
	_, err := c.
		Get(unsafeUrl, nil)
	if err == nil {
		t.Error("Unexcepted unsafe url:" + unsafeUrl)
	}

	res, err := c.
		WithOption(OPT_UNSAFE_TLS, true).
		Get(unsafeUrl)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("OPT_UNSAFE_TLS error")
	}
}

func TestPutJsonWithCharset(t *testing.T) {
	c := NewHttpClient()
	type jsonDataType struct {
		Name string
	}

	jsonData := jsonDataType{
		Name: "httpclient",
	}

	contentType := "application/json; charset=utf-8"
	res, err := c.
		WithHeader("Content-Type", contentType).
		PutJson("http://httpbin.org/put", jsonData)
	if err != nil {
		t.Error(err)
	}

	body, err := res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	var info ResponseInfo

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if info.Headers["Content-Type"] != contentType {
		t.Error("Setting charset not working: " + info.Headers["Content-Type"])
	}
}

func TestCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c := NewHttpClient()

	ch := make(chan error)
	go func() {
		_, err := c.Begin().
			WithOption(OPT_CONTEXT, ctx).
			Get("http://httpbin.org/delay/3")
		ch <- err
	}()

	time.Sleep(1 * time.Second)
	cancel()

	err := <-ch

	if err == nil || !strings.Contains(err.Error(), "cancel") {
		t.Error("Cancel error")
	}
}

func TestIssue41(t *testing.T) {
	c := NewHttpClient()
	c.Begin().Get("http://httpbin.org")
	c.Get("http://httpbin.org")
}

func TestBeforeRequestFunc(t *testing.T) {
	c := NewHttpClient()
	res, err := c.Begin().WithOption(OPT_BEFORE_REQUEST_FUNC, func(c *http.Client, r *http.Request) {
		r.Header.Add("test", "test")
	}).Get("http://httpbin.org/get")

	if err != nil {
		t.Error(err)
	}

	var info ResponseInfo

	body, err := res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(body, &info)

	if info.Headers["Test"] != "test" {
		t.Error("header not added")
	}
}
