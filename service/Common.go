package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Deansquirrel/goTextHandle/global"
	"github.com/Deansquirrel/goTextHandle/repository"
	"github.com/Deansquirrel/goToolCommon"
	log "github.com/Deansquirrel/goToolLog"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func Start() {
	go startWebService()
	go startAssets()
	//rep := repository.ProxyConfigDB{}
	//rep.Test()
}

func startAssets() {
	startAssetsWebService(
		global.SysConfig.Assets.Path,
		global.SysConfig.Assets.Port,
		global.SysConfig.Assets.LogLevel)
}

func startAssetsWebService(path string, port int, logLevel string) {
	log.Debug(fmt.Sprintf("start web service,port: %d,path: %s", port, path))
	defer log.Debug(fmt.Sprintf("start web service,port: %d,path: %s Complete", port, path))
	app := iris.New()
	app.Use(recover.New())
	app.Use(logger.New())
	app.Logger().SetLevel(logLevel)
	app.StaticWeb("/", path)

	go func() {
		_ = app.Run(
			iris.Addr(":"+strconv.Itoa(port)),
			iris.WithoutInterruptHandler,
			iris.WithoutServerError(iris.ErrServerClosed),
			iris.WithOptimizations,
		)
	}()
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch,
			os.Interrupt,
			syscall.SIGINT,
			os.Kill,
			syscall.SIGKILL,
			syscall.SIGTERM,
		)
		select {
		case <-ch:
			stopAssetsWebService(app, port, path)
		case <-global.Ctx.Done():
			stopAssetsWebService(app, port, path)
		}
	}()
}

func stopAssetsWebService(app *iris.Application, port int, path string) {
	log.Debug(fmt.Sprintf("stop web service,port: %d,path: %s", port, path))
	defer log.Debug(fmt.Sprintf("stop web service,port: %d,path: %s Complete", port, path))
	timeout := 3 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_ = app.Shutdown(ctx)
}

func startWebService() {
	app := iris.New()

	app.Use(recover.New())
	app.Use(logger.New())

	irisInit(app)
	irisRouter(app)
	irisStart(app)

	select {
	case <-global.Ctx.Done():
	}
}

func irisInit(app *iris.Application) {
	setIrisLogWrap(app)
	setIrisLogLevel(app)
	setIrisLogTimeFormat(app)
	setIrisLogFile(app)
}

func irisRouter(app *iris.Application) {
	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, //允许通过的主机名称
		AllowCredentials: true,
	})

	v1 := app.Party("/", crs).AllowMethods(iris.MethodOptions)
	{
		v1.Post("/text", textHandle)
	}
}

type textRequest struct {
	Phone string `json:"phone"`
}

type textResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func textHandle(ctx iris.Context) {
	var req textRequest
	err := ctx.ReadJSON(&req)
	if err != nil {
		writeResponse(ctx, &textResponse{
			ErrCode: -1,
			ErrMsg:  fmt.Sprintf("read json error: %s", err.Error()),
		})
		return
	}

	repHxDb := repository.HxDB{}
	rCodeList, err := repHxDb.GetRCode(req.Phone)
	if err != nil {
		writeResponse(ctx, &textResponse{
			ErrCode: -1,
			ErrMsg:  err.Error(),
		})
		return
	}
	if len(rCodeList) != 1 {
		errMsg := fmt.Sprintf("rCode list num error: exp 1 act %d", len(rCodeList))
		log.Error(errMsg)
		writeResponse(ctx, &textResponse{
			ErrCode: -1,
			ErrMsg:  errMsg,
		})
		return
	}
	rCode := rCodeList[0]
	log.Debug(fmt.Sprintf("update rcode is %s", rCode))

	crmDzService := NewCrmDz(global.SysConfig.UpdateInfo.CrmDz)
	err = crmDzService.UpdateWxMembershipNo(rCode)
	if err != nil {
		writeResponse(ctx, &textResponse{
			ErrCode: -1,
			ErrMsg:  err.Error(),
		})
		return
	}

	writeResponse(ctx, &textResponse{
		ErrCode: 0,
		ErrMsg:  "success",
	})
}

func writeResponse(ctx iris.Context, r *textResponse) {
	rBody, err := json.Marshal(r)
	if err != nil {
		errStr := fmt.Sprintf("tran obj to str err: %s", err.Error())
		log.Error(errStr)
		log.Warn(fmt.Sprintf("errCode: %d,errMsg: %s", r.ErrCode, r.ErrMsg))
		rBody = []byte("{\"errcode\"：-1，\"errmsg\":\"tran obj to str err\"}")
	}
	_, err = ctx.Write(rBody)
	if err != nil {
		log.Error(fmt.Sprintf("write response error: %s", err.Error()))
		log.Warn(fmt.Sprintf("errCode: %d,errMsg: %s", r.ErrCode, r.ErrMsg))
	}
	return
}

func irisStart(app *iris.Application) {
	log.Warn("StartWebService")
	defer log.Warn("StartWebService Complete")
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch,
			os.Interrupt,
			syscall.SIGINT,
			os.Kill,
			syscall.SIGKILL,
			syscall.SIGTERM,
		)
		select {
		case <-ch:
			irisStop(app)
			defer global.Cancel()
		case <-global.Ctx.Done():
			irisStop(app)
		}
	}()
	go func() {
		_ = app.Run(
			iris.Addr(":"+strconv.Itoa(global.SysConfig.Iris.Port)),
			iris.WithoutInterruptHandler,
			iris.WithoutServerError(iris.ErrServerClosed),
			iris.WithOptimizations,
		)
	}()
}

func irisStop(app *iris.Application) {
	log.Warn("StopWebService")
	defer log.Warn("StopWebService complete")
	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_ = app.Shutdown(ctx)
}

func setIrisLogLevel(app *iris.Application) {
	app.Logger().SetLevel(global.SysConfig.Iris.LogLevel)
}

func setIrisLogTimeFormat(app *iris.Application) {
	app.Logger().SetTimeFormat("2006-01-02 15:04:05")
}

func setIrisLogWrap(app *iris.Application) {
	app.Logger().DisableNewLine()
	app.Logger().SetPrefix(goToolCommon.GetWrapStr())
}

func setIrisLogFile(app *iris.Application) {
	reSetLogFile(app)
	time.AfterFunc(getRemainingTime(), func() {
		setIrisLogFile(app)
	})
}

//获取当日所剩时间
func getRemainingTime() time.Duration {
	todayLast := goToolCommon.GetDateStr(time.Now()) + " 23:59:59"
	todayLastTime, err := time.ParseInLocation("2006-01-02 15:04:05", todayLast, time.Local)
	if err != nil {
		log.Warn("获取当日所剩时间时发生错误:" + err.Error())
		return time.Minute
	}
	return time.Duration(todayLastTime.Unix()-time.Now().Local().Unix()+1) * time.Second
}

//设置日志输出对象
func reSetLogFile(app *iris.Application) {
	path, err := getIrisLogPath()
	if err != nil {
		log.Warn("刷新iris日志对象,获取当前路径时遇到错误:" + err.Error())
		return
	}
	fileName := "iris_" + goToolCommon.GetDateStr(time.Now()) + ".log"
	w, err := getIrisLogWriter(path, fileName)
	if err != nil {
		log.Warn("刷新iris日志对象,获取io.Writer遇到错误:" + err.Error())
		return
	}
	if w != nil {
		app.Logger().SetOutput(w)
		if global.SysConfig.Total.StdOut {
			app.Logger().AddOutput(os.Stdout)
		}
		log.Debug("SetLogFile")
	}
}

//获取iris写日志对象
func getIrisLogWriter(path string, fileName string) (io.Writer, error) {
	f, err := os.OpenFile(path+fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	return f, nil
}

//获取iris日志路径
func getIrisLogPath() (string, error) {
	path := log.Path
	if strings.Trim(path, " ") == "" {
		path, err := goToolCommon.GetCurrPath()
		if err != nil {
			return "", err
		}
		return path + "\\" + "log" + "\\", nil
	}
	if strings.HasSuffix(path, "\\") {
		path = path + "\\"
	}
	return path, nil
}
