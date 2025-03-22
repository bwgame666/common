package service

import (
	b64 "encoding/base64"
	"fmt"
	"time"

	"github.com/bwgame666/common/libs"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/xxtea/xxtea-go/xxtea"
)

type HttpService struct {
	route           *router.Router
	MiddlewareList  []MiddlewareFunc
	ApiTimeoutMsg   string
	ApiTimeout      time.Duration
	EncryptResponse bool
	GzipResponse    bool
	EncryptKey      string
}

type RequestHandler interface{}

type ResponseData struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func New(middlewareList []MiddlewareFunc, encryptKey string, encryptResponse bool) *HttpService {
	return &HttpService{
		route:           router.New(),
		MiddlewareList:  middlewareList,
		ApiTimeoutMsg:   `{"code": 408001,"message":"The server response timed out. Please try again later."}`,
		ApiTimeout:      time.Second * 30,
		EncryptResponse: encryptResponse,
		GzipResponse:    false,
		EncryptKey:      encryptKey,
	}
}

func (that *HttpService) EnableGzip(gzipResponse bool) {
	that.GzipResponse = gzipResponse
}

func (that *HttpService) StartServer(addr string) {
	srv := &fasthttp.Server{
		Handler:            that.middlewareDecorator(that.route.Handler),
		ReadTimeout:        that.ApiTimeout,
		WriteTimeout:       that.ApiTimeout,
		MaxRequestBodySize: 51 * 1024 * 1024,
	}
	if err := srv.ListenAndServe(addr); err != nil {
		fmt.Println("Error in ListenAndServe: ", err)
	}
}

func (that *HttpService) Post(path string, handle RequestHandler) {
	that.route.POST(path, fasthttp.TimeoutHandler(validatorDecorator(that, handle), that.ApiTimeout, that.ApiTimeoutMsg))
}

func (that *HttpService) Get(path string, handle RequestHandler) {
	that.route.GET(path, fasthttp.TimeoutHandler(validatorDecorator(that, handle), that.ApiTimeout, that.ApiTimeoutMsg))
}

func (that *HttpService) Put(path string, handle RequestHandler) {
	that.route.PUT(path, fasthttp.TimeoutHandler(validatorDecorator(that, handle), that.ApiTimeout, that.ApiTimeoutMsg))
}

func (that *HttpService) Delete(path string, handle RequestHandler) {
	that.route.DELETE(path, fasthttp.TimeoutHandler(validatorDecorator(that, handle), that.ApiTimeout, that.ApiTimeoutMsg))
}

func (that *HttpService) Head(path string, handle RequestHandler) {
	that.route.HEAD(path, fasthttp.TimeoutHandler(validatorDecorator(that, handle), that.ApiTimeout, that.ApiTimeoutMsg))
}

func (that *HttpService) Options(path string, handle RequestHandler) {
	that.route.OPTIONS(path, fasthttp.TimeoutHandler(validatorDecorator(that, handle), that.ApiTimeout, that.ApiTimeoutMsg))
}

func (that *HttpService) Patch(path string, handle RequestHandler) {
	that.route.PATCH(path, fasthttp.TimeoutHandler(validatorDecorator(that, handle), that.ApiTimeout, that.ApiTimeoutMsg))
}

func (that *HttpService) Response(ctx *fasthttp.RequestCtx, data *ResponseData) {

	bytes, err := libs.JsonMarshal(data)
	if err != nil {
		ctx.SetBody([]byte(err.Error()))
		return
	}

	if !that.EncryptResponse {
		ctx.SetStatusCode(200)
		ctx.SetContentType("application/json")
		if that.GzipResponse {
			ctx.Response.Header.Set("Content-Encoding", "gzip")
			gzippedData := fasthttp.AppendGzipBytes(nil, bytes)
			ctx.SetBody(gzippedData)
		} else {
			ctx.SetBody(bytes)
		}
		return
	}

	if that.EncryptKey == "" {
		ctx.SetContentType("text/plain")
		ctx.SetBody([]byte(""))
		return
	}
	encryptData := xxtea.Encrypt(bytes, []byte(that.EncryptKey))
	sEnc := b64.StdEncoding.EncodeToString(encryptData)
	ctx.SetStatusCode(200)
	ctx.SetContentType("text/plain")
	ctx.SetBody([]byte(sEnc))
}

func (that *HttpService) middlewareDecorator(handler fasthttp.RequestHandler) fasthttp.RequestHandler {

	return func(ctx *fasthttp.RequestCtx) {

		for _, mFunc := range that.MiddlewareList {
			if err := mFunc(that, ctx); err != nil {
				data := &ResponseData{
					Code:    408002,
					Message: err.Error(),
				}
				switch err.Error() {
				case "forbidden":
					data.Code = 408001
				case "unauthorized":
					data.Code = 408002
				case "not-whitelisted":
					data.Code = 408003
				case "system-in-maintain":
					data.Code = 408004
				default:

				}
				that.Response(ctx, data)
				return
			}
		}

		startTime := time.Now()
		// 处理http请求
		handler(ctx)

		// 高耗时请求处理
		costTime := time.Since(startTime)
		if costTime > 2*time.Second {
			path := string(ctx.Path())
			info := fmt.Sprintf("path: %s, query args: %s, post args: %s, ts: %s, time cost: %v",
				path,
				ctx.QueryArgs().String(),
				ctx.PostArgs().String(),
				startTime.Format("2006-01-02 15:04:05"),
				costTime,
			)
			fmt.Println(info)
		}
	}
}
