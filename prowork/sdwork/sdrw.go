package sdwork

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"time"

	"github.com/refitself/rslibs/rscrypto"
	"github.com/spf13/afero"
	"rshub/product/rsapp/selfdata/prowork/common"
)

const cExt = ".sd"
var tcore *common.TResp

func initForSDCore(loadPath string) {
	tcore = loadRWCore(loadPath)
	if tcore == nil || tcore.SDAuth == nil {
		tcore = &common.TResp{
			SDAuth:    func() func(data []byte) error{return nil},
			SDPush:    func(selfID, fileName string, data []byte) (bool, error) {return false, nil},
			SDPull:    func(selfID string) (bool, []string, error) {return false, nil, nil},
			SDEncrypt: func(data []byte) []byte {return nil},
			SDDecrypt: func(data []byte) []byte {return nil},
		}
	}
}

func loadRWCore(loadPath string) (rwcore *common.TResp) {
	rwcore = &common.TResp{}
	if loadPath != "" {
		if p, err := plugin.Open(loadPath); err == nil && p != nil {
			if s, err := p.Lookup("SDDecrypt"); err == nil && s != nil {
				rwcore.SDDecrypt = s.(common.T_SDDecrypt)
			}
			if s, err := p.Lookup("SDPull"); err == nil && s != nil {
				rwcore.SDPull = s.(common.T_SDPull)
			}
			if s, err := p.Lookup("SDEncrypt"); err == nil && s != nil {
				rwcore.SDEncrypt = s.(common.T_SDEncrypt)
			}
			if s, err := p.Lookup("SDPush"); err == nil && s != nil {
				rwcore.SDPush = s.(common.T_SDPush)
			}
			if s, err := p.Lookup("sdAuth"); err == nil && s != nil {
				rwcore.SDAuth = s.(common.T_SDAuth)
			}
		} else {
			//fmt.Printf("load plugin for sdcore failed, detail: %v", err)
		}
	}
	return
}

// ==============================encrypt=========================================================
// metaKey: only for update meta.sd
// 加密数据只会存储在本地文件系统
func SDEncryptData(outputFileName, metaKey string, data, dataPrivate, dataPublic, netPrivate, netPublic []byte) string {
	passwd, encryptPasswd := pwdForEncrypt(dataPublic, netPrivate, netPublic)

	// work
	dataBytesArr := make([][]byte, 0)
	dataBytes := rscrypto.AesEncryptECB(data, []byte(passwd))
	dataBytesArr = append(dataBytesArr, []byte(encryptPasswd))
	dataBytesArr = append(dataBytesArr, dataBytes)
	if metaKey != "" {
		dataBytesArr = append(dataBytesArr, []byte(selfMeta.AuthID))
		dataBytesArr = append(dataBytesArr, rscrypto.AesEncryptECB([]byte(metaKey), []byte(common.UpgradeCode)))
	}
	dataBytesBuf, err := json.Marshal(dataBytesArr)
	PanicShutdown(err)
	dataBytesBuf = []byte(hex.EncodeToString(dataBytesBuf))
	if strings.HasSuffix(outputFileName, "meta.sd") {
		PanicShutdown(common.CreateFile(vSysFS, vSysFS, "", outputFileName, dataBytesBuf))
	} else if outputFileName != "" {
		PanicShutdown(common.CreateFile(vSDFS, vSDFS, "", initEncryptEnv(outputFileName), dataBytesBuf))
	}
	return string(dataBytes)
}

// 加密数据只会存储在本地文件系统
func SDEncryptFile(outputFileName string, dataPrivate, dataPublic, netPrivate, netPublic []byte, resourcePath string) (string, []byte) {
	if _, err := vSDFS.Stat(resourcePath); err != nil {
		PanicShutdown(err)
	}
	fileName := initEncryptEnv(outputFileName)
	passwd, encryptPasswd := pwdForEncrypt(dataPublic, netPrivate, netPublic)

	// encrypt for selfdata
	fileBytes := make([]byte, 0)
	fZipPath := strings.Replace(fileName, cExt, ".zip", -1)
	PanicShutdown(common.Zip(vSDFS, vSDFS, resourcePath, fZipPath, ""))
	defaultHandle := func(data []byte) []byte {
		fileBytesArr := make([][]byte, 0)
		fileBytesArr = append(fileBytesArr, json.RawMessage(encryptPasswd))
		fileBytesArr = append(fileBytesArr, rscrypto.AesEncryptECB(data, []byte(passwd)))
		fileBytesBuf, err := json.Marshal(fileBytesArr)
		PanicShutdown(err)
		return []byte(hex.EncodeToString(fileBytesBuf))
	}
	PanicShutdown(common.CreateFileForEncrypt(vSDFS, vSysFS, fZipPath, fileName, func(data []byte) []byte {
		if tcore != nil && tcore.SDEncrypt != nil {
			if respBuf := tcore.SDEncrypt(data); respBuf != nil {
				return respBuf
			}
		}
		return defaultHandle(data)
	}))

	// clear
	PanicShutdown(vSDFS.RemoveAll(fZipPath))

	return fileName, fileBytes
}

func initEncryptEnv(outputFileName string) string {
	if outputFileName == "" {
		outputFileName = "output" + common.IDByTimeNow() + cExt
	} else {
		if filepath.Ext(outputFileName) != cExt {
			outputFileName = outputFileName + cExt
		}
	}

	outputPath := filepath.Join(selfMeta.StoreDir, C_Prefix_En, C_Prefix_Self)
	if _, err := vSDFS.Stat(outputPath); err != nil {
		PanicShutdown(vSDFS.MkdirAll(outputPath, os.ModePerm))
	}
	if _, err := vSDFS.Stat(filepath.Join(outputPath, outputFileName)); err == nil {
		time.Sleep(1)
		outputFileName = "output" + common.IDByTimeNow() + cExt
	}
	return filepath.Join(outputPath, outputFileName)
}

func pwdForEncrypt(dataPublic, netPrivate, netPublic []byte) (string, string) {
	random := common.GetRandom(11, false)
	if netPrivate == nil || netPublic == nil {
		random = common.GetRandom(32, false)
	}
	passwd := getRandomKey([]byte(random), netPrivate, netPublic)

	var err error
	var outputPasswd []byte
	//if isRsaKey(publicKey) {
	//	outputPasswd, err = rscrypto.RsaEncrypt([]byte(passwd), []byte(publicKey))
	//} else {
	//	outputPasswd, err = rscrypto.EcdsaEncrypt([]byte(passwd), []byte(publicKey))
	//}
	outputPasswd, err = rscrypto.RsaEncrypt([]byte(random), dataPublic)
	PanicShutdown(err)
	return passwd, string(outputPasswd)
}

func copyDatas(srcFS, dstFS afero.Fs, srcDir, dstDir string) {
	workSrcDir := filepath.ToSlash(srcDir)
	if !filepath.IsAbs(workSrcDir) {
		if !strings.HasPrefix(workSrcDir, "./") {
			workSrcDir = "./" + workSrcDir
		}
	}
	if _, err := dstFS.Stat(dstDir); err != nil {
		common.InfoLog("create directory: %s", dstDir)
		PanicShutdown(dstFS.MkdirAll(dstDir, os.ModePerm))
	}

	if finfo, err := srcFS.Stat(workSrcDir); err != nil {
		PanicShutdown(fmt.Errorf("invalid workSrcDir: %s", workSrcDir))
	} else {
		if !finfo.IsDir() {
			PanicShutdown(common.CreateFile(srcFS, dstFS, workSrcDir, filepath.Join(dstDir, srcDir), nil))
			return
		}
	}

	PanicShutdown(afero.Walk(srcFS, workSrcDir, func(path string, info fs.FileInfo, err error) error {
		if path == workSrcDir {
			return err
		}

		workPath := path
		if srcDir == "datas" || srcDir == "resource" {
			workPath = filepath.Join(dstDir, strings.TrimPrefix(path, srcDir))
		} else {
			workPath = filepath.Join(dstDir, path)
		}
		if _, err := dstFS.Stat(workPath); err == nil {
			common.InfoLog("resource already exists: %s", workPath)
			return nil
		}
		common.DebugLog("src: %s, dst: %s, path: %s, workPath: %s", workSrcDir, dstDir, path, workPath)

		if info.IsDir() {
			common.InfoLog("create directory: %s", workPath)
			return dstFS.MkdirAll(workPath, os.ModePerm)
		} else {
			return common.CreateFile(srcFS, dstFS, path, workPath, nil)
		}
	}))
}
// ==============================encrypt=========================================================

// ==============================decrypt=========================================================
func SdDecryptData(inputFilePath string, data, dataPrivate, dataPublic, netPrivate, netPublic []byte) ([]byte, error) {
	if inputFilePath != "" {
		workFS := vSDFS
		if strings.HasSuffix(inputFilePath, "meta.sd") {
			workFS = vSysFS
		}
		if err := common.ReadFileForDecrypt(workFS, inputFilePath, func(buf []byte) []byte {
			data = buf[:]
			return buf
		}); err != nil {
			return nil, err
		}
	}

	fileBytes, pwdBytes := pwdForDecrypt(data, dataPrivate, netPrivate, netPublic)
	respBuf := rscrypto.AesDecryptECB(fileBytes, pwdBytes)
	return respBuf, nil
}

func SDDecryptFile(inputFilePath, outputPath string, dataPrivate, dataPublic, netPrivate, netPublic []byte) string {
	if _, err := vSysFS.Stat(inputFilePath); err != nil {
		PanicShutdown(err)
	}
	var fZipPath string
	dirOutput := initDecryptEnv(outputPath)

	// default handle
	defaultHandle := func(data []byte) []byte {
		fileBytes, pwdBytes := pwdForDecrypt(data, dataPrivate, netPrivate, netPublic)
		respBuf := rscrypto.AesDecryptECB(fileBytes, pwdBytes)
		return respBuf
	}

	// decrypt to rs.fs
	fZipPath = strings.Replace(inputFilePath, cExt, ".zip", -1)
	decryptErr := common.CreateFileForDecrypt(vSysFS, vSDFS, inputFilePath, fZipPath, func(data []byte) []byte {
		if tcore != nil && tcore.SDDecrypt != nil {
			if respBuf := tcore.SDDecrypt(data); respBuf != nil {
				return respBuf
			}
		}
		return defaultHandle(data)
	})
	if decryptErr != nil {
		PanicShutdown(vSDFS.RemoveAll(fZipPath))
		PanicShutdown(decryptErr)
	}

	// unzip to os.fs
	if err := common.Unzip(vSDFS, vSDFS, fZipPath, dirOutput, ""); err != nil {
		PanicShutdown(vSDFS.RemoveAll(fZipPath))
		PanicShutdown(err)
	} else {
		PanicShutdown(vSDFS.RemoveAll(fZipPath))
	}
	return dirOutput
}

func initDecryptEnv(outputPath string) string {
	dirOutput := outputPath
	if dirOutput == "" || dirOutput == "memory" {
		dirOutput = filepath.Join(selfMeta.StoreDir, C_Prefix_De, C_Prefix_Self)
	}
	if _, err := vSDFS.Stat(dirOutput); dirOutput != "" && err != nil {
		common.InfoLog("create directory: %s", dirOutput)
		PanicShutdown(vSDFS.MkdirAll(dirOutput, os.ModePerm))
	}
	return dirOutput
}

func pwdForDecrypt(data, dataPrivate, netPrivate, netPublic []byte) (outBytes []byte, pwdBytes []byte) {
	bytesArr := make([][]byte, 0)
	dataBuf, _ := hex.DecodeString(string(data))
	json.Unmarshal(dataBuf, &bytesArr)
	if len(bytesArr) >= 2 {
		outBytes = bytesArr[1]
		pwdBuf, err := rscrypto.RsaDecrypt(bytesArr[0], dataPrivate)
		if err != nil {
			PanicShutdown(err)
		}
		pwdBytes = []byte(getRandomKey(pwdBuf, netPrivate, netPublic))
	}
	return
}
// ==============================decrypt=========================================================

func getRandomKey(pwdRandom, privateKey, publicKey []byte) (ret string) {
	if pwdRandom == nil {
		pwdRandom = []byte(common.GetRandom(32, false))
	}
	if privateKey == nil || publicKey == nil {
		ret = string(pwdRandom)
	} else {
		sharedKey, err := rscrypto.GetSharedKey(privateKey, publicKey)
		if err != nil {
			common.ErrorLog(common.Error(fmt.Errorf("generate random key failed, detail: %s", err.Error())).Error())
			return ""
		}
		ret = string(pwdRandom) + base64.StdEncoding.EncodeToString(sharedKey)
	}
	runesRandom := []rune(ret)
	if len(runesRandom) < 32 {
		for i := 0; i < 32; i++ {
			ret += "0"
		}
	}
	return string([]rune(ret)[:31])
}