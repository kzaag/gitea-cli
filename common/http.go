package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

/*

m: http method
u: http url
req: request body
res: response body [ptr]
hdr: headers to add, content-type is already added
ec: expected status code
*/
func HttpRequest(m, u string, req, res interface{}, hdr http.Header, ec int) error {

	var reqr io.Reader
	var err error

	if req != nil {
		rb, err := json.Marshal(req)
		if err != nil {
			return err
		}
		reqr = bytes.NewReader(rb)
	}

	httpReq, err := http.NewRequest(m, u, reqr)
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	httpReq.Header.Set("Accept", "application/json; charset=utf-8")
	for x := range hdr {
		for y := range hdr[x] {
			httpReq.Header.Add(x, hdr[x][y])
		}
	}

	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}

	defer httpRes.Body.Close()

	if httpRes.StatusCode != ec {
		var msgFmt = "invalid status code, expected %d got %d;\n%s"
		var msgB []byte
		var msg string
		if httpRes.Body != nil {
			msgB, err = ioutil.ReadAll(httpRes.Body)
			if err != nil {
				return err
			}
			msg = fmt.Sprintf(msgFmt, ec, httpRes.StatusCode, string(msgB))
		} else {
			msg = fmt.Sprintf(msgFmt, ec, httpRes.StatusCode, "<nothing in body>")
		}
		return errors.New(msg)
	}

	if res == nil {
		return nil
	}

	bt, err := ioutil.ReadAll(httpRes.Body)
	//fmt.Println(string(bt))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bt, res); err != nil {
		return err
	}

	return nil
}
