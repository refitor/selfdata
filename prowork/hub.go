package prowork

import (
	"context"
	"encoding/base64"
	"net/url"
	"runtime"
	"time"

	// "encoding/base64"
	"errors"
	"fmt"

	// "net/url"
	// "runtime"
	// "time"

	"rshub/product/rsapp/selfdata/prowork/sdweb"

	rice "github.com/GeertJohan/go.rice"
	"github.com/jpillora/overseer/fetcher"
	"github.com/refitself/rslibs/libpush"
	"github.com/refitself/rslibs/rscrypto"

	// "github.com/refitself/rslibs/rscrypto"
	"rshub/product/rsapp/selfdata/prowork/common"
	"rshub/product/rsapp/selfdata/prowork/sdwork"
)

var fExit func()

func Init(box *rice.Box, version, debugCode string, funcExit func()) {
	fExit = funcExit
	common.InitLog()
	common.DebugCode = debugCode
	common.EnableShutdown = true
	libpush.InitEmail("smtp.126.com:465", "refitself@126.com", "ROJNHEGZDGMOBALQ")

	// init for component
	sdwork.V_Web_Run = sdweb.Run
	sdwork.V_Web_Init = sdweb.Init

	// 1. open by the app.metaKey from meta.sd of app
	// 2. open by the local.metaKey from meta.sd
	// 3. update the app.metaKey to meta.sd
	metaKey, fMeta, err := getMetaFunc(box, "", "/meta.sd")
	panicShutdown(err)
	if fMeta != nil && fMeta("Code") != "" {
		metaKeyLocal := getLocalMetaKey(fMeta("Code"), "./sd/meta.sd")
		_, fMeta, err = getMetaFunc(box, metaKeyLocal, "/meta.sd")
		panicShutdown(err)
	}
	sdwork.Shutdown = shutdown
	sdwork.PanicShutdown = panicShutdown
	sdwork.Init(box, version, metaKey, fMeta)
}

func Run(ctx context.Context) {
	sdwork.Run()
	shutdown()
}

func panicShutdown(err error) {
	if err != nil {
		common.InfoLog(err.Error())
		fmt.Println("")
		//rslog.Error(err.Error())
		//state.GracefulShutdown <- true
		if common.EnableShutdown {
			shutdown()
		}
	}
}

func shutdown() {
	beforeExit()
	fExit()
}

func beforeExit() {

}

func GetFetcher() fetcher.Interface {
	if sdwork.CloneSelfMetaForReadonly().AutoUpgrade {
		paramBuf, err := rscrypto.EcdsaEncrypt([]byte(common.UpgradeCode), []byte(sdwork.GetRSPublic()))
		common.DebugLog("===>code: %s, err: %v", common.UpgradeCode, err)
		if err != nil {
			return nil
		}
		ext := common.GetAppExt()
		if ext != "" {
			ext = "-" + ext
		}
		eparam := base64.URLEncoding.EncodeToString(paramBuf)
		common.DebugLog("ready upgrade for selfdata: %s", fmt.Sprintf("%s/selfdata/selfdata_%s_%s%s?code=%s", sdwork.C_Url_Download, runtime.GOOS, runtime.GOARCH, common.GetAppExt(), url.QueryEscape(eparam)))
		return &fetcher.HTTP{
			URL:      fmt.Sprintf("%s/selfdata/selfdata_%s_%s%s?code=%s", sdwork.C_Url_Download, runtime.GOOS, runtime.GOARCH, common.GetAppExt(), url.QueryEscape(eparam)),
			Interval: 5 * time.Second, //60 * time.Minute,
		}
	}
	return nil
}

func GetPreUpgrade() func(tempBinaryPath string) error {
	return func(tempBinaryPath string) error {
		common.DebugLog("===>before upgrade: %s", tempBinaryPath)
		if sdwork.CheckUpgrade() {
			return nil
		}
		return errors.New("permission denied")
	}
}
