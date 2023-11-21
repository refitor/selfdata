package sdwork

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"rshub/product/rsapp/selfdata/prowork/common"

	"github.com/manifoldco/promptui"
	"github.com/refitself/rslibs/rscrypto"
)

func AuthByEmail(authUrl, authID, mode string, pubKey []byte) (string, error) {
	if authID == "" {
		return "", errors.New("Dynamic authorization of operating private data failed, unable to obtain authorization ID, please switch to offline mode!!!")
	}

	if err := sdAuth(authUrl, authID, mode, pubKey); err != nil {
		return "", err
	}

	var promptCode promptui.Prompt
	promptCode = promptui.Prompt{
		Label:    "Authorization code",
		Validate: nil,
	}
	code, _ := promptCode.Run()
	return code, authCodeValidate(authID, code)
}

func authCodeValidate(email, code string) error {
	memCode := common.PopMemData(email, false)
	if code != fmt.Sprintf("%v", memCode) {
		return errors.New("Invalid authorization code")
	}
	return nil
}

func StrongAuthUrl(authUrl, authID, code string, pubKey []byte) string {
	// format: app+email+code
	param := fmt.Sprintf("selfdata+%s+%s", authID, code)
	paramBuf, err := rscrypto.EcdsaEncrypt([]byte(param), []byte(pubKey))
	if err != nil {
		return ""
	}
	eparam := base64.URLEncoding.EncodeToString(paramBuf)
	return fmt.Sprintf("%s?param=%s", authUrl, url.QueryEscape(eparam))
}

func sdAuth(authUrl, authID, mode string, pubKey []byte) error {
	code := common.GetRandom(6, true)
	PanicShutdown(common.PushMemDataByTime(authID, code, 5, time.Minute))

	dataMap := make(map[string]string, 0)
	dataMap["To"] = authID
	dataMap["Msg"] = code
	dataMap["Subject"] = "dynamic authorization"
	dataBuf, err := json.Marshal(dataMap)
	PanicShutdown(err)

	if tcore != nil && tcore.SDAuth != nil {
		if fHandle := tcore.SDAuth(); fHandle != nil {
			return fHandle(dataBuf)
		}
	}
	if mode == C_Mode_Offline {
		visitUrl := StrongAuthUrl(authUrl, authID, code, pubKey)
		common.InfoLog("If you cannot scan the QR code, please manually access the authorization in your browser: %s\n", visitUrl)
		return common.RenderString(visitUrl)
	}

	// send at local
	if err := common.EmailSend(authID, "dynamic authorization", fmt.Sprintf("[selfdata] code for dynamic authorization: %s", code)); err != nil {
		common.InfoLog("Dynamic authorization code sending abnormal: %s", err.Error())
		return err
	}
	common.InfoLog("Successfully send code, timeout: 5 minutes")
	return nil
}
