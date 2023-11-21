package common

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"rshub/product/rsapp/selfdata/prowork/common/zip"

	"github.com/refitself/rslibs/rscrypto"
	"github.com/refitself/rslibs/libpush"
	"github.com/skip2/go-qrcode"
	//"github.com/yeka/zip"
	"github.com/spf13/afero"
)

var DebugCode string

func IsDebug() bool {
	return DebugCode == "6y4W0V"
}

func InfoPrint(args ...interface{}) {
	if len(args) == 0 {
		return
	}
	if format, ok := args[0].(string); ok {
		//_, file, line, _ := runtime.Caller(1)
		format = fmt.Sprintf("Info===>%s", format)
		log.Printf(format, args[1:]...)
	} else {
		log.Println(args...)
	}
}

func InfoLog(args ...interface{}) {
	if len(args) == 0 {
		return
	}
	if format, ok := args[0].(string); ok {
		//_, file, line, _ := runtime.Caller(1)
		//format = fmt.Sprintf("Info===%s %v===>%s", filepath.Base(file), line, format)
		//log.Printf(format, args...)
		format = format + "\n"
		fmt.Printf(format, args[1:]...)
		//log.Printf(format, args...)
	} else {
		fmt.Println(args...)
	}
}

func DebugLog(args ...interface{}) {
	if IsDebug() {
		if len(args) == 0 {
			return
		}
		if format, ok := args[0].(string); ok {
			_, file, line, _ := runtime.Caller(1)
			format = fmt.Sprintf("Debug===%s %v===>%s", filepath.Base(file), line, format)
			log.Printf(format, args[1:]...)
		} else {
			log.Println(args...)
		}
	}
}

func ErrorLog(args ...interface{}) {
	if len(args) == 0 {
		return
	}
	if format, ok := args[0].(string); ok {
		//_, file, line, _ := runtime.Caller(1)
		//format = fmt.Sprintf("Error===%s %v===>%s", filepath.Base(file), line, format)
		//fmt.Printf(format, args...)
		format = "Error===>" + format + "\n"
		log.Printf(format, args[1:]...)
	} else {
		log.Println(args...)
	}
}

func IDByTimeNow() string {
	return fmt.Sprintf("%v", time.Now().UnixNano())
}

func GetAppExt() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func GetPluginExt() string {
	if runtime.GOOS == "windows" {
		return ".dll"
	}
	return ".so"
}

func IsValidEmail(email string) bool {
	isValid, _ := regexp.MatchString("^([a-z0-9_\\.-]+)@([\\da-z\\.-]+)\\.([a-z\\.]{2,6})$", email)
	return isValid
}

// RenderString as a QR code
func RenderString(s string) error {
	q, err := qrcode.New(s, qrcode.Medium)
	if err != nil {
		return err
	}
	fmt.Println(q.ToSmallString(false))
	return nil
}

// RenderImage returns a QR code as an image.Image
func RenderImage(s string) image.Image {
	q, err := qrcode.New(s, qrcode.Medium)
	if err != nil {
		log.Fatal(err)
	}
	return q.Image(256)
}

// store random code to memory for auth ============================
var v_memdata_map sync.Map

func autoClearByTimer(key string, timeout, timeUnit time.Duration) {
	for {
		select {
		case <-time.After(timeout * timeUnit):
			v_memdata_map.Delete(key)
			return
		}
	}
}

func PushMemDataByTime(key string, val interface{}, timeout, timeUnit time.Duration) error {
	if d, ok := v_memdata_map.Load(key); ok {
		return fmt.Errorf("PushRandomByTime failed, key: %v, val: %+v, exist_val: %v", key, val, d)
	}
	v_memdata_map.Store(key, val)

	// 到期自动失效
	if timeout > 0 {
		go autoClearByTimer(key, timeout, timeUnit)
	}
	return nil
}

func PopMemData(key string, bDelete bool) interface{} {
	if bDelete {
		defer v_memdata_map.Delete(key)
	}
	if d, ok := v_memdata_map.Load(key); ok && d != nil {
		return d
	}
	return ""
}

func GetRandomInt(max *big.Int) (int, error) {
	if max == nil {
		seed := "0123456789"
		alphanum := seed + fmt.Sprintf("%v", time.Now().UnixNano())
		max = big.NewInt(int64(len(alphanum)))
	}
	vrand, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return int(vrand.Int64()), nil
}

func GetRandom(n int, isNO bool) string {
	seed := "0123456789"
	if !isNO {
		seed = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}
	alphanum := seed + fmt.Sprintf("%v", time.Now().UnixNano())
	buffer := make([]byte, n)
	max := big.NewInt(int64(len(alphanum)))

	for i := 0; i < n; i++ {
		index, err := GetRandomInt(max)
		if err != nil {
			return ""
		}

		buffer[i] = alphanum[index]
	}
	return string(buffer)
}

// srcFile could be a single file or a directory
func Zip(srcFS, dstFS afero.Fs, srcFile string, destZip string, passwd string) error {
	zipfile, err := dstFS.Create(destZip)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	err = afero.Walk(srcFS, srcFile, func(path string, info os.FileInfo, werr error) error {
	//err = filepath.Walk(srcFile, func(path string, info os.FileInfo, werr error) error {
		if werr != nil {
			return werr
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, filepath.Dir(srcFile)+"/")
		// header.Name = path
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		var writer io.Writer
		if passwd == "" {
			w, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}
			writer = w
		} else {
			w, err := archive.Encrypt(header.Name, passwd, zip.AES256Encryption)
			if err != nil {
				return err
			}
			writer = w
		}

		if !info.IsDir() {
			file, err := srcFS.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}
		return err
	})
	return err
}

func Unzip(srcFS, dstFS afero.Fs, zipFile string, destDir string, passwd string) error {
	zipReader, err := zip.OpenReader(srcFS, zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			if _, statErr := dstFS.Stat(fpath); statErr != nil {
				if mkErr := dstFS.MkdirAll(fpath, os.ModePerm); mkErr != nil {
					return mkErr
				}
			}
		} else {
			if f.IsEncrypted() {
				f.SetPassword(passwd)
			}

			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()

			outFile, err := dstFS.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CreateFile(srcFS, dstFS afero.Fs, srcPath, dstPath string, data []byte) error {
	if data == nil && srcPath != "" {
		DebugLog("open: %s", srcPath)
		srcf, nerr := srcFS.Open(srcPath)
		if nerr != nil {
			return nerr
		}
		srcBuf, rerr := ioutil.ReadAll(srcf)
		if rerr != nil {
			return rerr
		}
		if err := srcf.Close(); err != nil {
			return err
		}
		data = srcBuf[:]
	}

	InfoLog("create: %s", dstPath)
	dstf, cerr := dstFS.Create(dstPath)
	if cerr != nil {
		return cerr
	}
	if count, werr := dstf.Write(data); werr != nil {
		return werr
	} else {
		DebugLog("write file successed: %v", count)
	}
	return dstf.Close()
}

func CreateFileForEncrypt(srcFS, dstFS afero.Fs, srcPath, dstPath string, fEncrypt func([]byte)[]byte) error {
	DebugLog("open: %s", srcPath)
	bf, nerr := srcFS.Open(srcPath)
	if nerr != nil {
		return nerr
	}

	bfBuf, rerr := ioutil.ReadAll(bf)
	if rerr != nil {
		return rerr
	}
	if fEncrypt != nil {
		bfBuf = fEncrypt(bfBuf)
	}

	InfoLog("create: %s", dstPath)
	//rslog.Debugf("创建文件: %s", filepath.Join(rootDir, path))
	f, aerr := dstFS.Create(dstPath)
	if aerr != nil {
		return aerr
	}
	count, cerr := f.Write(bfBuf)
	if cerr != nil {
		return cerr
	}

	bf.Close()
	ferr := f.Close()
	DebugLog("write file successed: %v, close: %v", count, ferr)
	return ferr
}

func ReadFileForDecrypt(srcFS afero.Fs, srcPath string, fEncrypt func([]byte)[]byte) error {
	DebugLog("open: %s", srcPath)
	bf, nerr := srcFS.Open(srcPath)
	if nerr != nil {
		return nerr
	}

	bfBuf, rerr := ioutil.ReadAll(bf)
	if rerr != nil {
		return rerr
	}
	if fEncrypt != nil {
		fEncrypt(bfBuf)
	}
	return bf.Close()
}

func CreateFileForDecrypt(srcFS, dstFS afero.Fs, srcPath, dstPath string, fEncrypt func([]byte)[]byte) error {
	DebugLog("open: %s", srcPath)
	bf, nerr := srcFS.Open(srcPath)
	if nerr != nil {
		return nerr
	}

	bfBuf, rerr := ioutil.ReadAll(bf)
	if rerr != nil {
		return rerr
	}
	if fEncrypt != nil {
		if bfBuf = fEncrypt(bfBuf); len(bfBuf) == 0 {
			return bf.Close()
		}
	}

	if !strings.HasSuffix(dstPath, ".zip") {
		InfoLog("create: %s", dstPath)
	}
	//rslog.Debugf("创建文件: %s", filepath.Join(rootDir, path))
	f, aerr := dstFS.Create(dstPath)
	if aerr != nil {
		return aerr
	}
	count, cerr := f.Write(bfBuf)
	if cerr != nil {
		return cerr
	}

	bf.Close()
	ferr := f.Close()
	DebugLog("write file successed: %v, close: %v", count, ferr)
	return ferr
}

func EmailSend(email, subject, content string) error {
	if libpush.IsValidEmail(email) {
		_, err := libpush.PushByEmail(email, subject, "", content, nil)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("invalid email: %s", email)
}

// decode by rsa
func DecodeNetParams(param, decodeKey string) ([]string, error) {
	hexParam, err := url.QueryUnescape(param)
	if err != nil {
		return nil, err
	}
	paramBuf, err := hex.DecodeString(hexParam)
	if err != nil {
		return nil, fmt.Errorf("hex decode failed, detail: %s, uparam: %s", err.Error(), hexParam)
	}
	//decodeKey, err := GetSharedKey(selfPrivateKey, publicKey)
	//if err != nil {
	//	return nil, err
	//}

	//dparam := AesDecryptECB(paramBuf, decodeKey)
	//if len(dparam) == 0 {
	//	return nil, fmt.Errorf("aes decode failed, hexParam: %s", hexParam)
	//}
	dparam, err := rscrypto.RsaDecrypt(paramBuf, []byte(decodeKey))
	if err != nil {
		return nil, fmt.Errorf("rsa decode failed, hexParam: %s, detail: %s", hexParam, err.Error())
	}
	dparams := make([]string, 0)
	if err := json.Unmarshal([]byte(dparam), &dparams); err != nil {
		return nil, fmt.Errorf("rsa unmarshal failed, hexParam: %s, detail: %s", hexParam, err.Error())
	}
	return dparams, nil
}

// count: 重试次数, -1为无限循环
// periodRetryTime: 间隔多久进行重试
// timeUnit: 间隔重试时间的单位, time.Second...
func RunByRetry(fRun func() error, count int, periodRetryTime, timeUnit time.Duration) (err error) {
	for i := count; i > 0; i-- {
		err = fRun()
		if err == nil {
			break
		}
		if periodRetryTime != 0 {
			time.Sleep(periodRetryTime * timeUnit)
		}
	}
	return err
}
