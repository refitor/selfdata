package sdweb

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/gorilla/sessions"
	"github.com/phyber/negroni-gzip/gzip"
	"image/jpeg"
	"net/http"
	"path/filepath"
	"rshub/product/rsapp/selfdata/prowork/sdwork"
	"strings"
	"time"

	"rshub/product/rsapp/selfdata/prowork/common"

	rice "github.com/GeertJohan/go.rice"
	"github.com/didip/tollbooth"
	limiter2 "github.com/didip/tollbooth/limiter"
	"github.com/didip/tollbooth_negroni"
	"github.com/julienschmidt/httprouter"
	"github.com/refitself/rslibs/libhttp"
	//"github.com/rs/cors"
	"github.com/urfave/negroni"
)

func init() {
	v_Session_Store = sessions.NewCookieStore([]byte(fmt.Sprintf("%v", time.Now().UnixNano())))
	v_Session_Store.MaxAge(3600)
}

func Run(ctx context.Context, box *rice.Box, port string) string {
	vBox = box
	if port == "" {
		port = "7351"
	}

	n := negroni.New()
	n.UseFunc(newCors)
	n.UseFunc(newGzip)
	n.Use(newRateLimite())
	n.UseFunc(newAPILog)
	//n.Use(negroni.NewLogger())
	n.UseFunc(newPermissionCheck)
	n.UseHandlerFunc(initRouter().ServeHTTP)
	server := &http.Server{Addr: fmt.Sprintf(":%v", port), Handler: n}

	common.InfoLog("run web server successed, listen: %v", port)
	if err := server.ListenAndServe(); err != nil {
		common.ErrorLog(fmt.Errorf("run web server failed, detail: %s", err.Error()).Error())
	}
	<-ctx.Done()
	server.Shutdown(ctx)
	return fmt.Sprintf("%v", port)
}

func newAPILog(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if !strings.HasPrefix(r.URL.Path, "/api") {
		next(rw, r)
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/api/user") && !WebStatusCheck(r) {
		if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
			libhttp.ResponseRedirect(rw, r, "/")
		} else {
			http.Redirect(rw, r, "/", http.StatusFound)
		}
		return
	}
	common.DebugLog("%s %s, request: %+v", r.Method, r.URL.Path, r)

	ts := time.Now()
	next(rw, r)
	common.InfoPrint("%s %s, time: %v ms", r.Method, r.RequestURI, time.Now().Sub(ts).Milliseconds())
}

func newGzip(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if strings.HasPrefix(r.URL.Path, "/api") {
		next(rw, r)
		return
	}
	gzip.Gzip(gzip.DefaultCompression).ServeHTTP(rw, r, next)
}

func newRateLimite() negroni.Handler {
	limiter := tollbooth.NewLimiter(1, &limiter2.ExpirableOptions{DefaultExpirationTTL: time.Hour, ExpireJobInterval: time.Second})
	limiter.SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"}).SetMethods([]string{"GET", "POST"})
	limiter.SetMessage("You have reached maximum request limit.")
	return tollbooth_negroni.LimitHandler(limiter)
}

func newCors(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	bAllow := false
	fmt.Println(r.RemoteAddr)
	if strings.Contains(r.RemoteAddr, "::1") || strings.HasPrefix(r.RemoteAddr, "127.0.0.1") {
		bAllow = true
	} else if sdwork.HasRight(sdwork.C_Right_Strong) {
		bOrigin1 := strings.HasPrefix(r.RemoteAddr, "10.")
		bOrigin2 := strings.HasPrefix(r.RemoteAddr, "172.16.")
		bOrigin3 := strings.HasPrefix(r.RemoteAddr, "192.168.")
		bAllow = bOrigin1 || bOrigin2 || bOrigin3
	}
	if !bAllow {
		libhttp.ResponseError(rw, r, libhttp.C_Error_Denied)
		return
	}
	next(rw, r)
}

func newPermissionCheck(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	reqPath := libhttp.WebParams(r).Get("path")
	if reqPath == "" {
		reqPath = libhttp.WebParams(r).Get("filepath")
	}
	if selfAuthID != "" && selfAuthID != popFromSession(r) && reqPath == filepath.Join(decryptMemDir, "meta.json") {
		libhttp.ResponseError(rw, r, libhttp.C_Error_Denied)
		return
	}
	// permission check for sync
	// permission check for visit
	// permission check for relay
	next(rw, r)
}

//func newDecryptBefore(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
//	if !strings.HasPrefix(r.URL.Path, "/api") {
//		next(rw, r)
//		return
//	}
//	WebDecryptBefore(r)
//	next(rw, r)
//}

func initRouter() *httprouter.Router {
	router := httprouter.New()
	router.NotFound = indexHandler

	// api visit
	//router.HandlerFunc(http.MethodPost, "/api/kv/set", ApiKVSet)
	//router.HandlerFunc(http.MethodPost, "/api/kv/get", ApiKVGet)
	//router.HandlerFunc(http.MethodPost, "/api/data/get", ApiDataGet)

	// user status
	router.HandlerFunc(http.MethodPost, "/api/user/auth", WebUserAuth)
	router.HandlerFunc(http.MethodGet, "/api/user/logout", WebUserLogout)
	router.HandlerFunc(http.MethodGet, "/api/user/status", WebUserStatus)
	router.HandlerFunc(http.MethodGet, "/api/cert/public", func(w http.ResponseWriter, r *http.Request) {
		respBuf := base64.StdEncoding.EncodeToString([]byte(popFromSession(r)))
		libhttp.ResponseDirect(w, r, []byte(C_Encrypt_Prefix+respBuf), "text/plain")
	})

	// file resource
	router.HandlerFunc(http.MethodGet, "/api/file/view", WebFileView)
	router.HandlerFunc(http.MethodGet, "/api/file/scan", WebFileScan)
	router.HandlerFunc(http.MethodPost, "/api/file/upload", WebFileUpload)
	router.HandlerFunc(http.MethodGet, "/api/file/preview", WebFilePreview)
	router.HandlerFunc(http.MethodPost, "/api/resource/operate", WebResOperate)
	return router
}

func (p microFS) Open(name string) (http.File, error) {
	if name == "/" {
		name = ""
	}
	return vBox.HTTPBox().Open("/web" + name)
}

func WebUserAuth(w http.ResponseWriter, r *http.Request) {
	kind := libhttp.WebParams(r).Get("kind")
	authID := libhttp.WebParams(r).Get("authID")

	switch kind {
	case c_Kind_Qrcode:
		if !sdwork.HasRight(sdwork.C_Right_Strong) {
			libhttp.ResponseError(w, r, libhttp.C_Error_Denied)
			return
		}
		common.PopMemData(authID, true)
		code := common.GetRandom(6, true)
		if err := common.PushMemDataByTime(authID, code, 1, time.Minute); err != nil {
			libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "code create failed").Error())
			return
		}
		visitUrl := sdwork.StrongAuthUrl(sdwork.C_Url_Auth, authID, code, []byte(sdwork.GetRSPublic()))

		// 生成二维码
		qrCode, _ := qr.Encode(visitUrl, qr.M, qr.Auto)
		qrCode, _ = barcode.Scale(qrCode, 150, 150)

		// 二维码编码成base64字符串
		imgBuff := bytes.NewBuffer(nil)
		jpeg.Encode(imgBuff, qrCode, &jpeg.Options{100})
		qrCodeImg := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(imgBuff.Bytes())
		webServeResponse(w, r, libhttp.WebResponse{Data: qrCodeImg}, "application/json")
		return
	case c_Kind_Send:
		if selfMeta != nil && selfMeta.Mode == sdwork.C_Mode_Offline {
			libhttp.ResponseError(w, r, libhttp.Error("selfMeta.Mode is offline", "offline mode, please use code scanning authentication").Error())
			return
		}
		common.PopMemData(authID, true)
		code := common.GetRandom(6, true)
		if err := common.PushMemDataByTime(authID, code, 1, time.Minute); err != nil {
			libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "code create failed").Error())
			return
		}
		if err := common.EmailSend(authID, "dynamic authorization", fmt.Sprintf("[selfdata] code for dynamic authorization: %s", code)); err != nil {
			libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "code send failed, please use code scanning authentication").Error())
			return
		}
	case c_Kind_Verify:
		WebUserLogin(w, r)
		return
	default:
		libhttp.ResponseError(w, r, libhttp.C_Error_Denied)
		return
	}
	webServeResponse(w, r, libhttp.WebResponse{Data: libhttp.C_Data_Success}, "application/json")
}

func WebUserLogin(w http.ResponseWriter, r *http.Request) {
	pwd := libhttp.WebParams(r).Get("pwd")
	authID := libhttp.WebParams(r).Get("authID")
	justVerify := libhttp.WebParams(r).Get("justVerify")

	// verify code
	memCode := common.PopMemData(authID, false)
	if fmt.Sprintf("%v", memCode) == "" {
		if _, err := sdwork.AuthByGoogle(selfMeta.GoogleSecret, pwd); err != nil {
			libhttp.ResponseError(w, r, libhttp.Error("Invalid authorization code", "code verify failed").Error())
			return
		}
	} else if pwd != fmt.Sprintf("%v", memCode) {
		libhttp.ResponseError(w, r, libhttp.Error("Invalid authorization code", "code verify failed").Error())
		return
	}
	// if successed, remove memCode
	common.PopMemData(authID, true)

	//if common.IsDebug() && pwd != "" {
	//	libhttp.ResponseError(w, r, libhttp.Error("Invalid authorization code", "code verify failed").Error())
	//	return
	//}

	if justVerify == "" || justVerify == "false" {
		if err := pushToSession(w, r, authID, "", ""); err != nil {
			libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "session init failed").Error())
			return
		}
		if err := initAuthEnv(authID); err != nil {
			libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "session environment init failed").Error())
			return
		}
		if err := initDecryptEnv(authID); err != nil {
			libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "encryption environment init failed").Error())
			return
		}
	}
	webServeResponse(w, r, libhttp.WebResponse{Data: libhttp.C_Data_Success}, "application/json")
}

func WebUserLogout(w http.ResponseWriter, r *http.Request) {
	if err := memfs.RemoveAll(filepath.Join(decryptMemDir, popFromSession(r))); err != nil {
		libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "logout failed").Error())
		return
	}
	if err := removeSession(w, r); err != nil {
		libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "logout failed").Error())
		return
	}
	webServeResponse(w, r, libhttp.WebResponse{Data: libhttp.C_Data_Success}, "application/json")
}

func WebUserStatus(w http.ResponseWriter, r *http.Request) {
	if !WebStatusCheck(r) {
		libhttp.ResponseError(w, r, libhttp.C_Error_Denied)
		return
	}
	webServeResponse(w, r, libhttp.WebResponse{Data: libhttp.C_Data_Success}, "application/json")
}

func WebStatusCheck(r *http.Request) bool {
	return popFromSession(r) != ""
}

func pushToSession(w http.ResponseWriter, r *http.Request, id, private, public string) error {
	if session, err := v_Session_Store.Get(r, c_Session_ID); session != nil {
		session.Values[c_Session_AuthID] = id
		//session.Values[c_Session_Public] = public
		//session.Values[c_Session_Private] = private
		if err := session.Save(r, w); err != nil {
			return err
		}
		return nil
	} else {
		return err
	}
}

func popFromSession(r *http.Request) string {
	if session, _ := v_Session_Store.Get(r, c_Session_ID); session != nil {
		if sid := session.Values[c_Session_AuthID]; sid != nil {
			return fmt.Sprintf("%v", sid)
		}
	}
	return ""
}

func removeSession(w http.ResponseWriter, r *http.Request) error {
	if sid := popFromSession(r); sid != "" {
		if session, _ := v_Session_Store.Get(r, c_Session_ID); session != nil {
			delete(session.Values, c_Session_AuthID)
			if err := session.Save(r, w); err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("permission denied")
}
