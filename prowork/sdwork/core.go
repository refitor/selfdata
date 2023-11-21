package sdwork

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	rice "github.com/GeertJohan/go.rice"
	"github.com/manifoldco/promptui"
	"github.com/refitself/rslibs/libhttp"
	"github.com/refitself/rslibs/rscrypto"
	"github.com/spf13/afero"
	"net/url"
	"os"
	"path/filepath"
	"rshub/product/rsapp/selfdata/prowork/common"
	"strings"
)

// var for global
var (
	vBox     *rice.Box
	vVersion string
	vMetaKey string
	fMeta    func(string) string

	vSDFS    = afero.NewOsFs()
	vSysFS   = afero.NewOsFs()
	selfMeta = &common.TSDMeta{}

	Shutdown      func()
	PanicShutdown func(error)
)

const (
	C_Right_Relay  = "relay"
	C_Right_Web    = "web"
	C_Right_Visit  = "visit"
	C_Right_Sync   = "sync"
	C_Right_Strong = "strong"

	C_Cmd_Read  = "read"
	C_Cmd_Write = "write"

	C_Mode_Offline  = "offline"
	C_Mode_Standard = "standard"
)

const (
	C_Prefix_Self  = "self"
	C_Prefix_En    = "encryption"
	C_Prefix_De    = "decryption"
	C_Url_Auth     = "https://refitself.cn/api/user/auth"
	C_Url_Relay    = "https://relay.refitself.cn"
	C_Url_Download = "https://refitself.cn/download"
	C_Url_Official = "https://refitself.cn"
)

func Init(box *rice.Box, version, metaKey string, vMeta func(string) string) {
	vBox = box
	fMeta = vMeta
	vVersion = version
	vMetaKey = metaKey

	cmdName := ""
	if len(os.Args) > 1 && os.Args[1] != "-h" && os.Args[1] != "help" && os.Args[1] != "--help" && !strings.HasPrefix(os.Args[1], "--") {
		cmdName = os.Args[1]
	}
	vSupportList = initForMeta(cmdName)
}

func initForMeta(cmdName string) []string {
	vID = fMeta("ID")
	vSysPub = fMeta("Public")
	vSysPriv = fMeta("Private")
	common.DebugLog("private: %s, public: %s", vSysPriv, vSysPub)

	// from meta.sd
	vMetaPath, _ = filepath.Abs(filepath.Join("./sd", "meta.sd"))
	strMeta, err := SdDecryptData(vMetaPath, []byte{}, []byte(vSysPriv), []byte(vSysPub), nil, nil)
	common.DebugLog("complete parse meta.sd: %s, error: %s", strMeta, err)
	json.Unmarshal([]byte(strMeta), &selfMeta)

	if selfMeta.AuthID == "" {
		doInitMeta()
		if cmdName == C_Cmd_Read || cmdName == C_Cmd_Write {
			PanicShutdown(errors.New("\nPlease initialize the system through web dynamic authorization first, run: selfdata"))
		}
	}

	// 校验命令是否可执行
	paidRightList := make([]string, 0)
	for _, v := range selfMeta.PaidInfoMap {
		paidRightList = append(paidRightList, initRightList(strings.Split(v, ","))...)
	}
	supportList := strings.Split(vSupport, ", ")
	supportList = append(supportList, paidRightList...)
	if !strings.Contains(strings.Join(supportList, ","), cmdName) {
		PanicShutdown(fmt.Errorf("Invalid operation, supported command: %s", strings.Join(supportList, ",")))
	}
	return supportList
}

func initRightList(paidRights []string) []string {
	ret := make([]string, 0)
	for _, right := range paidRights {
		switch right {
		case C_Right_Relay:
			ret = append(ret, C_Right_Relay)
			fallthrough
		case C_Right_Visit:
			ret = append(ret, C_Right_Visit)
			fallthrough
		case C_Right_Sync:
			ret = append(ret, C_Right_Sync)
			fallthrough
		case C_Right_Strong:
			ret = append(ret, C_Right_Strong)
		}
	}
	return ret
}

func HasRight(right string) bool {
	for _, v := range vSupportList {
		if v == right {
			return true
		}
	}
	return false
}

func GetSettings() map[string]interface{} {
	ret := make(map[string]interface{}, 0)
	ret["Name"] = selfMeta.Name
	ret["AuthID"] = selfMeta.AuthID
	ret["Language"] = selfMeta.Language
	ret["StoreDir"] = selfMeta.StoreDir
	ret["SysRight"] = strings.Join(vSupportList, ", ")
	ret["ListenPort"] = selfMeta.ListenPort
	ret["AutoUpgrade"] = selfMeta.AutoUpgrade
	ret["DataPublic"] = selfMeta.DataPublic
	ret["DataPrivate"] = selfMeta.DataPrivate
	for _, cmd := range vSupportList {
		switch cmd {
		case C_Right_Relay:
			ret["Relay"] = selfMeta.Relay
			ret["NetPublic"] = selfMeta.NetPublic
			ret["NetPrivate"] = selfMeta.NetPrivate
			fallthrough
		case C_Right_Visit:
			ret["VisitAuthList"] = selfMeta.VisitAuthList
			fallthrough
		case C_Right_Sync:
			ret["Mode"] = selfMeta.Mode
			ret["Relay"] = selfMeta.Relay
			ret["RelayPub"] = selfMeta.RelayPub
			ret["NetPublic"] = selfMeta.NetPublic
			ret["NetPrivate"] = selfMeta.NetPrivate
			ret["SyncAuthList"] = selfMeta.SyncAuthList
			fallthrough
		case C_Right_Strong:
			ret["Mode"] = selfMeta.Mode
		}
	}
	return ret
}

func CheckUpgrade() bool {
	needUpgrade := selfMeta.AutoUpgrade
	if needUpgrade {
		prompt := promptui.Prompt{
			//Label: fmt.Sprintf("latest version: %s, current version: %s, confirm update?(y/n)", serverVersion, vVersion),
			Label:    "confirm upgrade?(y/n)",
			Validate: nil,
		}
		confirm, _ := prompt.Run()
		if confirm == "y" {
			needUpgrade = true
		}
	}
	return needUpgrade

	//doCheckUpgrade := func(serverVersion string) bool {
	//	needUpgrade := selfMeta.AutoUpgrade
	//	if !needUpgrade {
	//		prompt := promptui.Prompt{
	//			//Label: fmt.Sprintf("latest version: %s, current version: %s, confirm update?(y/n)", serverVersion, vVersion),
	//			Label: "confirm upgrade?(y/n)",
	//			Validate: nil,
	//		}
	//		confirm, _ := prompt.Run()
	//		if confirm == "y" {
	//			needUpgrade = true
	//		}
	//	}
	//	return needUpgrade
	//}
	//
	//var bUpgrade bool
	//var latestVersion string
	//retErr := common.HttpGetByFunc(C_Url_Download + "/files/selfdata/version.txt", func(buf []byte) {
	//	if len(strings.Split(string(buf), ".")) == 3 {
	//		latestVersion = strings.TrimSpace(string(buf))
	//		if latestVersion != strings.TrimSpace(vVersion) {
	//			common.InfoLog("upgrade: %s ---> %s", vVersion, latestVersion)
	//			bUpgrade = doCheckUpgrade(latestVersion)
	//		}
	//	}
	//})
	//if retErr != nil {
	//	common.ErrorLog(retErr.Error())
	//}
	//if bUpgrade {
	//	return latestVersion
	//}
	//return ""
}

func doInitMeta() {
	if selfMeta.AuthID != "" {
		return
	}
	common.InfoLog("\n===>init selfdata metadata")
	vNetPriv, vNetPub, ecdsaErr := rscrypto.GenerateEcdsaKey()
	vSelfPriv, vSelfPub, rsaErr := rscrypto.GenerateRsaKey()
	PanicShutdown(ecdsaErr)
	PanicShutdown(rsaErr)
	selfMeta = &common.TSDMeta{
		Name:          vID,
		Mode:          "standard",
		StoreDir:      "sd",
		Language:      "zh-CN",
		NetPublic:     string(vNetPub),
		NetPrivate:    string(vNetPriv),
		DataPublic:    string(vSelfPub),
		DataPrivate:   string(vSelfPriv),
		AutoUpgrade:   false,
		Relay:         C_Url_Relay,
		ListenPort:    "7351",
		RelayPub:      getRelayPublic(),
		SysRight:      "",
		GoogleSecret:  "",
		PaidInfoMap:   make(map[string]string, 0),
		VisitAuthList: make(map[string]string, 0),
		SyncAuthList:  make(map[string]*common.TSDPublic, 0),
	}
	if doUpdateMeta(selfMeta, false) {
		common.InfoLog("Successfully update selfdata metadata, name: %s, authID: %s, googleSecret: %s", selfMeta.Name, selfMeta.AuthID, selfMeta.GoogleSecret)
	} else {
		common.InfoLog("Failed to update selfdata metadata, user canceled")
	}
}

func doUpdateMeta(newTsdm *common.TSDMeta, bVerify bool) bool {
	if newTsdm.Name != selfMeta.Name && newTsdm.Name != "" {
		selfMeta.Name = newTsdm.Name
	}
	if newTsdm.Language != selfMeta.Language && newTsdm.Language != "" {
		selfMeta.Language = newTsdm.Language
	}
	if newTsdm.AutoUpgrade != selfMeta.AutoUpgrade {
		selfMeta.AutoUpgrade = newTsdm.AutoUpgrade
	}
	if newTsdm.Relay != selfMeta.Relay && newTsdm.Relay != "" {
		selfMeta.Relay = newTsdm.Relay
	}
	if newTsdm.AuthID != selfMeta.AuthID && newTsdm.AuthID != "" {
		selfMeta.AuthID = newTsdm.AuthID
	}
	if selfMeta.AuthID != "" && selfMeta.GoogleSecret == "" {
		selfMeta.GoogleSecret = NewGoogleAuth().GetSecret()
		visitUrl := NewGoogleAuth().GetQrcode(selfMeta.AuthID, selfMeta.GoogleSecret)
		common.InfoLog("If you cannot scan the QR code, please enter account and key manually: %s, %s\n", selfMeta.AuthID, selfMeta.GoogleSecret)
		PanicShutdown(common.RenderString(visitUrl))
	}
	if newTsdm.StoreDir != selfMeta.StoreDir && newTsdm.StoreDir != "" {
		selfMeta.StoreDir = newTsdm.StoreDir
	}
	if newTsdm.DataPublic != selfMeta.DataPublic && newTsdm.DataPublic != "" {
		selfMeta.DataPublic = newTsdm.DataPublic
	}
	if newTsdm.DataPrivate != selfMeta.DataPrivate && newTsdm.DataPrivate != "" {
		selfMeta.DataPrivate = newTsdm.DataPrivate
	}
	if newTsdm.RelayPub != selfMeta.RelayPub && newTsdm.RelayPub != "" {
		selfMeta.RelayPub = newTsdm.RelayPub
	}
	if newTsdm.NetPublic != selfMeta.NetPublic && newTsdm.NetPublic != "" {
		selfMeta.NetPublic = newTsdm.NetPublic
	}
	if newTsdm.NetPrivate != selfMeta.NetPrivate && newTsdm.NetPrivate != "" {
		selfMeta.NetPrivate = newTsdm.NetPrivate
	}
	if newTsdm.PaidInfoMap != nil && len(newTsdm.PaidInfoMap) > 0 {
		for k, v := range newTsdm.PaidInfoMap {
			selfMeta.PaidInfoMap[k] = v
		}
	}

	confirm := "y"
	if bVerify {
		prompt := promptui.Prompt{
			Label:    "selfdata metadata meta.sd will be updated, confirm execution?(y/n)",
			Validate: nil,
		}
		promptResult, err := prompt.Run()
		PanicShutdown(err)
		confirm = promptResult
	}

	if confirm == "y" {
		oldFS := vSDFS
		SetSDFS(vSysFS)
		selfMetaBuf, err := json.Marshal(selfMeta)
		if err != nil {
			PanicShutdown(fmt.Errorf("runForUpdate failed at json.Marshal, detail: %s", err.Error()))
		}
		if _, err := vSDFS.Stat(filepath.Join(selfMeta.StoreDir, C_Prefix_En, C_Prefix_Self)); err != nil {
			PanicShutdown(vSDFS.MkdirAll(filepath.Join(selfMeta.StoreDir, C_Prefix_En, C_Prefix_Self), os.ModePerm))
		}
		if _, err := vSDFS.Stat(filepath.Join(selfMeta.StoreDir, C_Prefix_De, C_Prefix_Self)); err != nil {
			PanicShutdown(vSDFS.MkdirAll(filepath.Join(selfMeta.StoreDir, C_Prefix_De, C_Prefix_Self), os.ModePerm))
		}
		vSDFS.Remove(vMetaPath)
		common.DebugLog("Start updating meta.sd: %s", string(selfMetaBuf))
		SDEncryptData(vMetaPath, vMetaKey, selfMetaBuf, []byte(vSysPriv), []byte(vSysPub), nil, nil)
		SetSDFS(oldFS)
		return true
	}
	return false
}

func GetRSPublic() string {
	public := `LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFNWp1Ky8yRXRMcEl3djY1R3JHK0krWHRMbWEzSQpZQUtleHJ0d1FLSDdxd0czOTVFZk0wY01ieE50NFdFV25BOXlWblVRWWY3NUJaKzY3VkRzT1R6U2ZnPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg==`
	pubBuf, _ := base64.StdEncoding.DecodeString(public)
	return string(pubBuf)
}

func getRelayPublic() string {
	return ""
}

func SetSDFS(fs afero.Fs) afero.Fs {
	vSDFS = fs
	return vSDFS
}

func SDDecryptMeta(fs afero.Fs, dstPath string) (*common.TSDMeta, error) {
	tsdmBuf, err := json.Marshal(GetSettings())
	if err != nil {
		return nil, err
	}
	common.DebugLog("SDDecryptMeta: %+v", string(tsdmBuf))
	if err := common.CreateFile(fs, fs, "", dstPath, tsdmBuf); err != nil {
		return nil, err
	}
	return selfMeta, nil
}

func SDEncryptMeta(buf []byte) error {
	memMeta := &common.TSDMeta{
		PaidInfoMap: make(map[string]string, 0),
	}
	if err := json.Unmarshal(buf, &memMeta); err != nil {
		return err
	}
	common.DebugLog("SDEncryptMeta: %+v", string(buf))

	if key, val, err := verifyPayAuthCode(buf); key != "" && val != "" && err == nil {
		memMeta.PaidInfoMap[key] = val
	} else if err != nil {
		return err
	}
	doUpdateMeta(memMeta, false)
	return nil
}

func verifyPayAuthCode(buf []byte) (key, val string, retErr error) {
	memMetaMap := make(map[string]interface{}, 0)
	if err := json.Unmarshal(buf, &memMetaMap); err != nil {
		return key, val, nil
	}
	if memMetaMap["PayAuthCode"] != nil {
		param := fmt.Sprintf("%s+%s", selfMeta.AuthID, memMetaMap["PayAuthCode"])
		verifyParam := fmt.Sprintf("%s-%s", selfMeta.AuthID, memMetaMap["PayAuthCode"])
		paramBuf, err := rscrypto.EcdsaEncrypt([]byte(param), []byte(GetRSPublic()))
		if err != nil {
			return key, val, err
		}
		eparam := hex.EncodeToString(paramBuf)
		common.DebugLog("ready verify for PayAuthCode: %s, param=%s", fmt.Sprintf("%s/api/paid/service", C_Url_Official), eparam)

		form := make(url.Values)
		form.Add("param", eparam)
		if err := libhttp.HttpPostByFunc(fmt.Sprintf("%s/api/paid/service", C_Url_Official), form, func(buf []byte) {
			common.DebugLog("++++++verify for PayAuthCode result: %s", string(buf))
			retMap := make(map[string]interface{}, 0)
			if err := json.Unmarshal(buf, &retMap); err != nil {
				common.DebugLog("verifyPayAuthCode failed, detail: %s, buf: %s", err.Error(), string(buf))
			} else if strings.Contains(fmt.Sprintf("%v", retMap["Data"]), "+") {
				retDatas := strings.Split(fmt.Sprintf("%v", retMap["Data"]), "+")
				if err := rscrypto.EcdsaVerify(verifyParam, retDatas[0], []byte(GetRSPublic())); err != nil {
					common.DebugLog("verifyPayAuthCode failed, detail: %s, buf: %s", err.Error(), string(buf))
				} else {
					key = retDatas[0]
					val = common.GetPaidRight(retDatas[1])
				}
			}
		}); err != nil {
			return key, val, err
		}
	}
	return key, val, nil
}

func CloneSelfMetaForReadonly() *common.TSDMeta {
	return selfMeta
}
