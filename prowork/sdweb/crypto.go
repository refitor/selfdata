package sdweb

import (
	"encoding/base64"
	"errors"
	"io/fs"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"rshub/product/rsapp/selfdata/prowork/common"
	"rshub/product/rsapp/selfdata/prowork/sdwork"

	"github.com/refitself/rslibs/libhttp"
	"github.com/refitself/rslibs/rscrypto"
)

const (
	c_ext = ".sd"

	c_kind_create_folder = "create-folder"
	c_kind_create_file   = "create-file"
	c_kind_rename        = "rename"
	c_kind_update        = "update"
	c_kind_delete        = "delete"
	c_kind_encrypt       = "encrypt"
	c_kind_decrypt       = "decrypt"
)

var v_decryptsd_list sync.Map
var v_createsd_list sync.Map

//var SDEncryptMeta func(buf []byte) error
//var SDDecryptMeta func(fs afero.Fs, dstPath string) (*common.TSDMeta, error)
//var SDDecryptFile func(inputFilePath, outputPath string) string
//var SDEncryptData func(outputFileName string, data []byte) string
//var SDDecryptData func(inputFilePath string, data []byte) ([]byte, error)
//var SDEncryptFile func(outputFileName, resourcePath string) (string, []byte)

func doDecryptFileAtMem(encryptDirAtLocal, encryptDirAtMemory, fSDFilePath, fSDMemPath string) error {
	if selfMeta == nil {
		return errors.New(libhttp.C_Error_Denied)
	}

	// make decrypt directory
	common.DebugLog("encryptDirAtLocal: %s, encryptDirAtMemory: %s", encryptDirAtLocal, encryptDirAtMemory)

	decryptResult := sdwork.SDDecryptFile(fSDFilePath, fSDMemPath, []byte(selfMeta.DataPrivate), []byte(selfMeta.DataPublic), nil, nil)
	common.DebugLog("decrypt at memory successed, fSDFilePath: %s, fSDMemPath: %s, decryptResult: %s", fSDFilePath, fSDMemPath, decryptResult)

	// 如果解密出的目录与父级目录名称一致, 重新解密一次以消除两个名称一致的.sd目录
	if memInfo, err := memfs.Stat(filepath.Join(fSDMemPath, filepath.Base(fSDMemPath))); err == nil && memInfo.IsDir() {
		common.DebugLog("start decrypt again, fRootWork: %s, fRepeatDir: %s", filepath.Dir(fSDMemPath), filepath.Join(fSDMemPath, filepath.Base(fSDMemPath)))
		if err := memfs.RemoveAll(fSDMemPath); err != nil {
			return err
		}
		fRootWork := filepath.Dir(fSDMemPath)
		decryptResult := sdwork.SDDecryptFile(fSDFilePath, fRootWork, []byte(selfMeta.DataPrivate), []byte(selfMeta.DataPublic), nil, nil)
		common.DebugLog("decrypt again at memory successed, fSDFilePath: %s, fSDMemPath: %s, decryptResult: %s", fSDFilePath, fSDMemPath, decryptResult)
	}
	return nil
}

// input: the path of file aaa.txt or directory aaa at memory
func doEncryptFileAtMem(encryptDirAtLocal, encryptDirAtMemory, fMemoryPath string) error {
	if selfMeta == nil {
		return errors.New(libhttp.C_Error_Denied)
	}

	fEncryptFileName := filepath.Join(filepath.Base(fMemoryPath))
	encryptResult, _ := sdwork.SDEncryptFile(fEncryptFileName, []byte(selfMeta.DataPrivate), []byte(selfMeta.DataPublic), nil, nil, fMemoryPath)
	common.DebugLog("encrypt at memory successed, fEncryptFileName: %s, fMemoryPath: %s, result: %s", fEncryptFileName, fMemoryPath, encryptResult)
	return nil
}

func readMemDir(dirname string) ([]fs.FileInfo, error) {
	f, err := memfs.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name() < list[j].Name() })
	return list, nil
}

func WebDecryptBefore(r *http.Request) {
	runes := []rune(strings.TrimSuffix(C_Encrypt_Prefix, ":"))
	strPrefix := strings.ToUpper(string(runes[0])) + strings.ToLower(string(runes[1:]))
	aesKey := libhttp.WebParams(r).Get(strPrefix)
	//aesKey := libnet.WebParams(r).Get("SFv3rg")
	common.DebugLog("r: %+v, aesKey: %s, prefix: %s", r, aesKey, strPrefix)
	if aesKey != "" {
		if aesKeyBuf, err := base64.StdEncoding.DecodeString(aesKey); err == nil && aesKeyBuf != nil {
			if outBuf, err := rscrypto.RsaDecrypt([]byte(aesKeyBuf), []byte(popFromSession(r))); err == nil {
				r.Header.Set("Aeskey", string(outBuf))
			}
		}
	}
}

func WebEncryptAfter(r *http.Request, data interface{}) []byte {
	aesKey := libhttp.WebParams(r).Get("Aeskey")
	if aesKey == "" {
		common.ErrorLog("encrypt for web failed with invalid aesKey")
	}
	strOut := base64.StdEncoding.EncodeToString(rscrypto.AesEncryptECB(data.([]byte), []byte(aesKey)))
	common.DebugLog("start doEncrypt for response, prefix: %s, strOut: %s, aesKey: %s", C_Encrypt_Prefix, strOut, aesKey)
	return []byte(C_Encrypt_Prefix + strOut)
}

func WebMetaEncrypt(w http.ResponseWriter, r *http.Request) {
	metaBuf := []byte(libhttp.WebParams(r).Get("data"))
	if err := sdwork.SDEncryptMeta(metaBuf); err != nil {
		libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "sync failed").Error())
		return
	}
	if meta, err := sdwork.SDDecryptMeta(memfs, filepath.Join(decryptMemDir, "meta.json")); err != nil {
		libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "sync failed").Error())
		return
	} else {
		selfMeta = meta
	}
	webServeResponse(w, r, libhttp.WebResponse{Data: libhttp.C_Data_Success}, "application/json")
}
