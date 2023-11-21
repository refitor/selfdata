package prowork

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"rshub/product/rsapp/selfdata/prowork/common"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	"github.com/refitself/rslibs/rscrypto"
)

func getLocalMetaKey(code, metaPath string) string {
	metaSDPath, _ := filepath.Abs(metaPath)
	fMetaData, err := os.Open(metaSDPath)
	if err != nil {
		common.DebugLog("getLocalMetaKey failed: %s", err.Error())
		return ""
	}

	dataBuf, err := ioutil.ReadAll(fMetaData)
	if err != nil {
		common.DebugLog("getLocalMetaKey failed: %s", err.Error())
		return ""
	}
	bytesArr := make([][]byte, 0)
	if err := json.Unmarshal(dataBuf, &bytesArr); err != nil {
		common.DebugLog("getLocalMetaKey failed: %s", err.Error())
		return ""
	}
	if len(bytesArr) > 0 {
		return string(rscrypto.AesDecryptECB(bytesArr[len(bytesArr) - 1], []byte(code)))
	}
	return ""
}

func doGetAESKey(aesKey []byte) []byte {
	if len(aesKey) > 100 {
		aesKeys := append([]rune(string(aesKey))[5:37], []rune(string(aesKey))[55:]...)
		sharedRunes := []rune(string(aesKeys))[67:]
		if len(sharedRunes) > 32 {
			aesKey = []byte(string(sharedRunes[:32]))
		} else if len(sharedRunes) < 32 {
			strSharedKey := string(sharedRunes)
			for i := 0; i < 32 - len(sharedRunes); i++ {
				strSharedKey += "0"
			}
			aesKey = []byte(strSharedKey)
		}
	}
	return aesKey
}

func getDecodeMeta(metaKey string, data []byte) (string, []byte) {
	var mkey []byte
	if metaKey == "" {
		metaArr := make([][]byte, 0)
		json.Unmarshal(data, &metaArr)
		if len(metaArr) == 2 {
			metaKey = string(metaArr[0])
			data = metaArr[1]
		}
	}
	mkey, _ = hex.DecodeString(metaKey)
	outBuf := rscrypto.AesDecryptECB(data, doGetAESKey(mkey))
	return metaKey, []byte(strings.TrimSpace(string(outBuf)))
}

func getMetaFunc(box *rice.Box, metaKey, fullPath string) (string, func(string)string, error) {
	fMeta, err := box.Open(fullPath)
	if err != nil {
		return "", nil, fmt.Errorf("open===>system processing exception: %s", err.Error())
	}

	dataBuf, err := ioutil.ReadAll(fMeta)
	if err != nil {
		return "", nil, fmt.Errorf("read===>system processing exception: %s", err.Error())
	}
	if strings.HasSuffix(fullPath, ".sd") {
		metaKey, dataBuf = getDecodeMeta(metaKey, dataBuf)
	}

	ret := make(map[string]string, 0)
	if !strings.HasSuffix(fullPath, "meta.sd") {
		ret["data"] = string(dataBuf)
	} else {
		if err = json.Unmarshal(dataBuf, &ret); err != nil {
			return "", nil, fmt.Errorf("parse===>system processing exception: %s", err.Error())
		}
	}

	return metaKey, func(key string) string {
		if data, ok := ret[key]; !ok {
			return ""
		} else {
			return data
		}
	}, nil
}
