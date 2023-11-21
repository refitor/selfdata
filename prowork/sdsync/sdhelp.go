package sdsync

import (
	"encoding/json"
	"fmt"

	"rshub/product/rsapp/selfdata/prowork/common"

	"github.com/refitself/rslibs/rscrypto"
)

func getPublicID() string {
	eid := rscrypto.AesEncryptECB([]byte(vID), []byte(vID))
	eName := rscrypto.AesEncryptECB([]byte(selfMeta.Name), []byte(selfMeta.Name))
	eAuthID := rscrypto.AesEncryptECB([]byte(selfMeta.AuthID), []byte(selfMeta.AuthID))
	return fmt.Sprintf("%s-%s-%s", eid, eName, eAuthID)
}

func getSwapID(id string) string {
	data, err := rscrypto.EcdsaEncrypt([]byte(vID), []byte(selfMeta.NetPublic))
	if err != nil {
		return ""
	}
	return string(data)
}

func getRelayID(id string) string {
	dataKey := selfMeta.DataPublic
	if id != vID {
		tsdPub, _ := selfMeta.SyncAuthList[id]
		if tsdPub != nil {
			dataKey = string(tsdPub.DataKey)
		} else {
			return ""
		}
	}

	data, err := rscrypto.RsaEncrypt([]byte(id), []byte(dataKey))
	if err != nil {
		return ""
	}
	return string(data)
}

func encodeParams(params ...string) []byte {
	paramsBuf, _ := json.Marshal(params)
	eparamsBuf, err := rscrypto.EcdsaEncrypt(paramsBuf, []byte(selfMeta.RelayPub))
	if err != nil {
		common.ErrorLog("initForRelay failed at RsaEncrypt, detail: %s", err.Error())
		return []byte{}
	}
	return eparamsBuf
}

func decodeParams(data []byte) (token, sig string, datas [][]byte) {
	respDatas := make([][]byte, 0)
	if err := json.Unmarshal(data, &respDatas); err != nil || len(respDatas) < 2 {
		common.ErrorLog("decodeParams failed, detail: %s", err.Error())
		return "", "", nil
	}
	if len(respDatas) > 2 {
		return string(respDatas[0]), string(respDatas[1]), respDatas[2:]
	}
	return string(respDatas[0]), string(respDatas[1]), [][]byte{}
}

func tokenCreate(publicKey, privateKey []byte) ([]byte, []byte) {
	token := rscrypto.AesEncryptECB([]byte(common.GetRandom(16, false)), dhKeyForAES(publicKey, privateKey))
	if signature, err := rscrypto.EcdsaSign(token, privateKey); err != nil {
		return nil, nil
	} else {
		return token, []byte(signature)
	}
}

func tokenVerify(token, signature string, publicKey, privateKey []byte) error {
	return rscrypto.EcdsaVerify(string(rscrypto.AesDecryptECB([]byte(token), dhKeyForAES(publicKey, privateKey))), signature, publicKey)
}

func dhKeyForAES(publicKey, privateKey []byte) []byte {
	dhKey, err := rscrypto.GetSharedKey(privateKey, publicKey)
	if err != nil {
		return nil
	}
	dhKeyRunes := []rune(string(dhKey))
	if len(dhKeyRunes) > 32 {
		dhKey = []byte(string(dhKeyRunes[:32]))
	} else if len(dhKeyRunes) < 32 {
		strSharedKey := string(dhKeyRunes)
		for i := 0; i < 32 - len(dhKeyRunes); i++ {
			strSharedKey += "0"
		}
		dhKey = []byte(strSharedKey)
	}
	return dhKey
}

