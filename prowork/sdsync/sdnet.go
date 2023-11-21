package sdsync

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/refitself/rslibs/libhttp"
	"rshub/product/rsapp/selfdata/prowork/common"
)

var (
	vIsValidRelay bool

	vID string
	bSync bool
	tcore *common.TResp
	selfMeta *common.TSDMeta
)

func SyncInit(vID string, tcore *common.TResp, selfMeta *common.TSDMeta) (retErr error) {
	vID = vID
	tcore = tcore
	selfMeta = selfMeta

	pubID := getPublicID()
	dataForm := url.Values{}
	token, sig := tokenCreate([]byte(selfMeta.RelayPub), []byte(selfMeta.NetPrivate))
	dataForm.Add("data", hex.EncodeToString(encodeParams(string(selfMeta.NetPublic))))
	dataForm.Add("param", hex.EncodeToString(encodeParams(pubID, string(token), string(sig), pubID)))
	if err := libhttp.HttpPostByFunc(fmt.Sprintf("%s/api/sd/push", selfMeta.Relay), dataForm, func(buf []byte) {
		rtoken, rsig, _ := decodeParams(buf)
		if tokenVerify(rtoken, rsig, []byte(selfMeta.RelayPub), []byte(selfMeta.NetPrivate)) != nil {
			retErr = errors.New("Initialization failed, illegal relay node")
			vIsValidRelay = false
		}
	}); err != nil {
		return err
	}
	vIsValidRelay = true
	return nil
}

func SDPull() (retDatas []string, retErr error) {
	if tcore != nil && tcore.SDPull != nil {
		if bHandle, datas, err := tcore.SDPull(vID); bHandle {
			return datas, err
		}
	}

	if vIsValidRelay {
		pubID := getPublicID()
		dataForm := url.Values{}
		token, sig := tokenCreate([]byte(selfMeta.RelayPub), []byte(selfMeta.NetPrivate))
		dataForm.Add("data", hex.EncodeToString(encodeParams(string(selfMeta.NetPublic))))
		dataForm.Add("param", hex.EncodeToString(encodeParams(pubID, string(token), string(sig), getRelayID(vID))))
		if err := libhttp.HttpPostByFunc(fmt.Sprintf("%s/api/data/pull", selfMeta.Relay), dataForm, func(buf []byte) {
			rtoken, rsig, datas := decodeParams(buf)
			if tokenVerify(rtoken, rsig, []byte(selfMeta.RelayPub), []byte(selfMeta.NetPrivate)) != nil {
				retErr = errors.New("Failed to pull private data, invalid relay node")
				vIsValidRelay = false
			}
			for _, data := range datas {
				retDatas = append(retDatas, string(data))
			}
		}); err != nil {
			return nil, err
		}
	}
	return nil, errors.New("Failed to pull private data, invalid relay node")
}

func SDPushData(recvIds []string, data []byte) (pushErr error) {
	if tcore != nil && tcore.SDPush != nil {
		if bHandle, err := tcore.SDPush(vID, "", data); bHandle {
			return err
		}
	}
	recvRelayIds := make([]string, 0)
	for _, recvID := range recvIds {
		recvRelayIds = append(recvRelayIds, getRelayID(recvID))
	}

	return common.RunByRetry(func() error {
		if vIsValidRelay {
			dataForm := url.Values{}
			dataForm.Add("data", string(data))
			token, sig := tokenCreate([]byte(selfMeta.RelayPub), []byte(selfMeta.NetPrivate))
			workParams := []string{getPublicID(), string(token), string(sig)}
			workParams = append(workParams, recvRelayIds...)
			dataForm.Add("param", hex.EncodeToString(encodeParams(workParams...)))
			if herr := libhttp.HttpPostByFunc(fmt.Sprintf("%s/api/data/push", selfMeta.Relay), dataForm, func(buf []byte) {
				recvRelayIds = make([]string, 0)
				rtoken, rsig, failedIds := decodeParams(buf)
				if tokenVerify(rtoken, rsig, []byte(selfMeta.RelayPub), []byte(selfMeta.NetPrivate)) != nil {
					vIsValidRelay = false
				}
				for _, failedId := range failedIds {
					recvRelayIds = append(recvRelayIds, string(failedId))
				}
			}); herr != nil {
				return fmt.Errorf("SDPushData failed at HttpPostByFunc, detail: %s", herr.Error())
			}
		}
		return errors.New("Failed to push data, invalid relay node")
	}, 3, 1, time.Second)
}

func SDPushFile(fileName string, recvIds []string, data []byte) (pushErr error) {
	if tcore != nil && tcore.SDPush != nil {
		if bHandle, err := tcore.SDPush(vID, fileName, data); bHandle {
			return err
		}
	}
	recvRelayIds := make([]string, 0)
	for _, recvID := range recvIds {
		recvRelayIds = append(recvRelayIds, getRelayID(recvID))
	}

	return common.RunByRetry(func() error {
		if vIsValidRelay {
			dataForm := url.Values{}
			dataForm.Add("data", string(data))
			token, sig := tokenCreate([]byte(selfMeta.RelayPub), []byte(selfMeta.NetPrivate))
			workParams := []string{getPublicID(), string(token), string(sig)}
			workParams = append(workParams, recvRelayIds...)
			dataForm.Add("param", hex.EncodeToString(encodeParams(workParams...)))
			//if herr := common.HttpPostFileByFunc(fmt.Sprintf("%s/api/data/push", vRelayUrl), dataForm, "data", fileName, data, func(buf []byte) {
			//	recvRelayIds = make([]string, 0)
			//	rtoken, rsig, failedIds := decodeParams(buf)
			//	if tokenVerify(rtoken, rsig, selfMeta.RelayPub, selfMeta.NetPrivate) != nil {
			//		vIsValidRelay = false
			//	}
			//	for _, failedId := range failedIds {
			//		recvRelayIds = append(recvRelayIds, string(failedId))
			//	}
			//}); herr != nil {
			//	return fmt.Errorf("SDPushData failed at HttpPostByFunc, detail: %s", herr.Error())
			//}
		}
		return errors.New("Failed to push file, invalid relay node")
	}, 3, 1, time.Second)
}
