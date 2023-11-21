package sdweb

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"rshub/product/rsapp/selfdata/prowork/common"
	"rshub/product/rsapp/selfdata/prowork/sdwork"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gorilla/sessions"
	"github.com/refitself/rslibs/libhttp"
	"github.com/spf13/afero"
)

type microFS string

var indexHandler = http.FileServer(microFS(""))

const (
	C_Encrypt_Prefix  = "SFv3rg:"
	c_Encryption_Self = "self"

	c_Kind_Send   = "send"
	c_Kind_Qrcode = "qrcode"
	c_Kind_Verify = "verify"
	c_Kind_Logout = "logout"

	c_Session_ID      = "sdweb-session"
	c_Session_AuthID  = "authID"
	c_Session_Public  = "public"
	c_Session_Private = "private"
)

var (
	v_Session_Store *sessions.CookieStore
	selfMeta        *common.TSDMeta
	encryptLocalDir string
	decryptMemDir   string
	selfAuthID      string
	vBox            *rice.Box
	memfs           afero.Fs
)

func Init(encryptDir, authID string) error {
	selfAuthID = authID
	encryptLocalDir = encryptDir
	memfs = sdwork.SetSDFS(afero.NewMemMapFs())
	decryptMemDir = filepath.Base(encryptLocalDir)
	if err := memfs.MkdirAll(decryptMemDir, os.ModePerm); err != nil {
		return err
	}
	selfMeta = sdwork.CloneSelfMetaForReadonly()
	common.DebugLog("=================+++%s, %s", encryptLocalDir, authID)
	return nil
}

func initAuthEnv(initAuthID string) error {
	if selfAuthID != "" {
		return nil
	}
	metaBuf, _ := json.Marshal(map[string]interface{}{"AuthID": initAuthID, "AutoUpgrade": true})
	if err := sdwork.SDEncryptMeta([]byte(metaBuf)); err != nil {
		return err
	}
	selfAuthID = initAuthID
	return nil
}

func initDecryptEnv(fEncryptName string) error {
	if fEncryptName == selfAuthID {
		fEncryptName = c_Encryption_Self
	}
	bHandle := false
	isSelf := fEncryptName == c_Encryption_Self
	fAbsEncryptDir, _ := filepath.Abs(encryptLocalDir)
	walkErr := filepath.Walk(fAbsEncryptDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil || path == fAbsEncryptDir {
			return err
		}
		if filepath.Base(path) != "meta.sd" {
			// support group by authID
			isVisitSelf := strings.Contains(path, fEncryptName)
			if isSelf || isVisitSelf {
				if info.IsDir() {
					common.DebugLog("folder: %s", path)
					fMemDir := strings.TrimPrefix(path, filepath.Dir(fAbsEncryptDir)+"/")
					if err := memfs.MkdirAll(fMemDir, os.ModePerm); err != nil {
						return err
					}
					bHandle = true
				} else if filepath.Ext(path) == c_ext {
					common.DebugLog("file: %s", path)
					// 初始化状态仅加载私有数据目录, 等到访问时再解密数据
					fDecryptPath := strings.TrimPrefix(path, filepath.Dir(fAbsEncryptDir)+"/")
					if err := memfs.MkdirAll(fDecryptPath, os.ModePerm); err != nil {
						return err
					}
					// 暂时一开始全部加载到内存
					// return doDecryptFileAtMem(encryptLocalDir, decryptMemDir, path)
				}
				return nil
			}
		}
		return nil
	})
	if walkErr != nil {
		return walkErr
	}
	if !bHandle {
		return errors.New(libhttp.C_Error_Denied)
	}
	if fEncryptName == c_Encryption_Self {
		meta, err := sdwork.SDDecryptMeta(memfs, filepath.Join(decryptMemDir, "meta.json"))
		selfMeta = meta
		return err
	}
	// } else {
	// 	selfMeta = sdwork.CloneSelfMetaForReadonly()
	// }
	return nil
}

func canScan(authID, name string) bool {
	if authID != selfAuthID && name != authID {
		return false
	}
	return true
}

func canScanFile(fpath string) bool {
	if fpath == "encryption/meta.json" {
		return false
	}
	return true
}
