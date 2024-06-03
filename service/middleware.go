package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/twmb/murmur3"
	"github.com/valyala/fasthttp"
	"github.com/xxtea/xxtea-go/xxtea"
	"net/url"
)

type middlewareFunc func(svr *HttpService, ctx *fasthttp.RequestCtx) error

func verify(str string) string {
	h32 := murmur3.SeedNew32(24)
	_, err := h32.Write([]byte(str))
	if err != nil {
		return ""
	}
	v := h32.Sum32()
	return fmt.Sprintf("%d", v)
}

func decryptMiddleware(svr *HttpService, ctx *fasthttp.RequestCtx) error {
	// 请求参数解密中间件

	allows := map[string]bool{
		"/heartbeat": true,
	}
	forbidden := errors.New("forbidden")

	path := string(ctx.Path())
	if _, ok := allows[path]; ok {
		return nil
	}

	version := string(ctx.Request.Header.Peek("v"))
	reqTime := string(ctx.Request.Header.Peek("X-Ca-Timestamp"))
	nonce := string(ctx.Request.Header.Peek("X-Ca-Nonce"))

	if svr.EncryptKey == "" {
		return forbidden
	}
	if version == "" {
		return forbidden
	}
	if reqTime == "" {
		return forbidden
	}
	if nonce == "" {
		return forbidden
	}

	uri := string(ctx.RequestURI())
	decodedValue, err := url.QueryUnescape(uri)
	if err != nil {
		fmt.Println("url.QueryUnescape = ", err.Error())
		return forbidden
	}

	if ctx.IsGet() || ctx.IsDelete() {
		str := fmt.Sprintf("%s%s%s%s", svr.EncryptKey, reqTime, decodedValue, version)
		h2 := verify(str)
		if h2 != nonce {
			fmt.Println("GET h2 = ", h2)
			fmt.Println("GET nonce = ", nonce)
			fmt.Println("GET str = ", str)
			return forbidden
		}
	} else if ctx.IsPost() || ctx.IsPut() {

		b := string(ctx.PostBody())
		str := fmt.Sprintf("%s%s%s%s%s", b, svr.EncryptKey, reqTime, decodedValue, version)
		h2 := verify(str)
		if h2 != nonce {
			fmt.Println("POST h2 = ", h2)
			fmt.Println("POST nonce = ", nonce)
			fmt.Println("POST str = ", str)
			return forbidden
		}
		data, err := base64.StdEncoding.DecodeString(b)
		if err != nil {
			fmt.Println("POST DecodeString err = ", err.Error())
			return forbidden
		}

		decryptData := xxtea.Decrypt(data, []byte(svr.EncryptKey))
		var args fasthttp.Args
		args.ParseBytes(decryptData)
		postArgs := ctx.PostArgs()
		postArgs.Reset()
		args.CopyTo(postArgs)
	} else {
		return forbidden
	}

	return nil
}
