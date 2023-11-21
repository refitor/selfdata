package sdweb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"rshub/product/rsapp/selfdata/prowork/common"

	"github.com/refitself/rslibs/libhttp"
)

func WebResOperate(w http.ResponseWriter, r *http.Request) {
	path := libhttp.WebParams(r).Get("path")
	data := libhttp.WebParams(r).Get("data")
	kind := libhttp.WebParams(r).Get("kind")
	common.DebugLog("WebResOperate, kind: %s, path: %s, data: %s", kind, path, data)

	switch kind {
	case c_kind_encrypt:
		WebResEncrypt(w, r)
		return
	case c_kind_decrypt:
		if _, ok := v_createsd_list.Load(path); !ok {
			if _, ok := v_decryptsd_list.Load(path); !ok {
				v_decryptsd_list.Store(path, true)
				ApiResDecrypt(w, r)
				return
			}
		}
		webServeResponse(w, r, libhttp.WebResponse{Data: libhttp.C_Data_Pass}, "application/json")
		return
	case c_kind_create_folder:
		if filepath.Base(filepath.Dir(path)) == decryptMemDir {
			data += c_ext
			v_createsd_list.Store(filepath.Join(path, data), true)
			common.DebugLog("start create folder for %s", filepath.Join(path, data))
		}
		if err := memfs.MkdirAll(filepath.Join(path, data), os.ModePerm); err != nil {
			libhttp.ResponseError(w, r, libhttp.Error(err.Error(), kind + " failed").Error())
			return
		}
	case c_kind_create_file:
		if f, err := memfs.Create(filepath.Join(path, data)); err != nil {
			libhttp.ResponseError(w, r, libhttp.Error(err.Error(), kind + " failed").Error())
			return
		} else {
			if err := f.Close(); err != nil {
				libhttp.ResponseError(w, r, libhttp.Error(err.Error(), kind + " failed at close").Error())
				return
			}
		}
	case c_kind_rename:
		if err := memfs.Rename(path, strings.Replace(path, filepath.Base(path), data, -1)); err != nil {
			libhttp.ResponseError(w, r, libhttp.Error(err.Error(), kind + " failed").Error())
			return
		}
	case c_kind_update:
		if path == filepath.Join(decryptMemDir, "meta.json") {
			WebMetaEncrypt(w, r)
			return
		} else {
			if f, err := memfs.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666); err != nil {
				libhttp.ResponseError(w, r, libhttp.Error(err.Error(), kind+" failed").Error())
				return
			} else {
				if _, err := f.Write([]byte(data)); err != nil {
					libhttp.ResponseError(w, r, libhttp.Error(err.Error(), kind+" failed").Error())
					return
				}
				if err := f.Close(); err != nil {
					libhttp.ResponseError(w, r, libhttp.Error(err.Error(), kind+" failed at close").Error())
					return
				}
			}
		}
	case c_kind_delete:
		if info, err := memfs.Stat(path); err == nil {
			if info.IsDir() {
				if err := memfs.RemoveAll(path); err != nil {
					libhttp.ResponseError(w, r, libhttp.Error(err.Error(), kind + " failed").Error())
					return
				}
			} else {
				if err := memfs.Remove(path); err != nil {
					libhttp.ResponseError(w, r, libhttp.Error(err.Error(), kind + " failed").Error())
					return
				}
			}
		} else {
			libhttp.ResponseError(w, r, libhttp.Error(err.Error(), kind + " failed").Error())
			return
		}
	}
	webServeResponse(w, r, libhttp.WebResponse{Data: libhttp.C_Data_Success}, "application/json")
}

func WebResEncrypt(w http.ResponseWriter, r *http.Request) {
	//fpath := libhttp.WebParams(r).Get("path")
	var decryptErr error
	workList := v_decryptsd_list
	v_createsd_list.Range(func(key, value interface{}) bool {
		workList.Store(key, true)
		return true
	})
	workList.Range(func(key, value interface{}) bool {
		if fmt.Sprintf("%v", key) != filepath.Join(decryptMemDir, "meta.json") {
			// encryption/self/demo.sd
			common.DebugLog("start encrypt for %s", fmt.Sprintf("%v", key))
			nodes := strings.Split(fmt.Sprintf("%v", key), "/")
			if len(nodes) >= 3 && decryptMemDir == nodes[0] && strings.HasSuffix(nodes[2], c_ext) {
				if err := doEncryptFileAtMem(encryptLocalDir, decryptMemDir, filepath.Join(decryptMemDir, nodes[1], nodes[2])); err != nil {
					decryptErr = err
					return false
				}
			}
		}
		return true
	})
	if decryptErr != nil {
		libhttp.ResponseError(w, r, libhttp.Error(decryptErr.Error(), "sync failed").Error())
		return
	}
	webServeResponse(w, r, libhttp.WebResponse{Data: libhttp.C_Data_Success}, "application/json")
}

func ApiResEncrypt(w http.ResponseWriter, r *http.Request) {
	// encryption/self/demo.sd
	encryptPath := libhttp.WebParams(r).Get("path")
	common.DebugLog("start encrypt for %s", encryptPath)
	nodes := strings.Split(encryptPath, "/")
	if len(nodes) >= 3 && decryptMemDir == nodes[0] && strings.HasSuffix(nodes[2], c_ext) {
		if err := doEncryptFileAtMem(encryptLocalDir, decryptMemDir, filepath.Join(decryptMemDir, nodes[1], nodes[2])); err != nil {
			libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "encrypt failed").Error())
		} else {
			webServeResponse(w, r, libhttp.WebResponse{Data: libhttp.C_Data_Success}, "application/json")
		}
		return
	}
	webServeResponse(w, r, libhttp.WebResponse{Data: libhttp.C_Data_Pass}, "application/json")
}

func ApiResDecrypt(w http.ResponseWriter, r *http.Request) {
	directory := libhttp.WebParams(r).Get("path")
	fAbsEncryptDir, _ := filepath.Abs(encryptLocalDir)

	// directory: encryption/self/demo.sd
	fSDFilePath := filepath.Join(filepath.Dir(fAbsEncryptDir), directory)
	if fInfo, err := os.Stat(fSDFilePath); err == nil && !fInfo.IsDir() {
		if err := doDecryptFileAtMem(encryptLocalDir, decryptMemDir, fSDFilePath, directory); err != nil {
			libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "decrypt failed").Error())
		} else {
			webServeResponse(w, r, libhttp.WebResponse{Data: libhttp.C_Data_Success}, "application/json")
		}
		return
	}
	webServeResponse(w, r, libhttp.WebResponse{Data: libhttp.C_Data_Pass}, "application/json")
}

func webServeResponse(w http.ResponseWriter, r *http.Request, data interface{}, contentType string, params ...interface{}) {
	//needEncrypt := strings.HasPrefix(r.URL.Path, "/api") && !strings.HasPrefix(r.URL.Path, "/api/user") && !strings.HasPrefix(r.URL.Path, "/api/cert/public")
	if len(params) > 0 {
		dataBuf := data.([]byte)
		fileName := fmt.Sprintf("%v", params[0])
		w.Header().Add("content-disposition", "attachment; filename=\""+fileName+"\"")
		w.Header().Set("Content-Length", strconv.Itoa(len(dataBuf)))
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(dataBuf)
	} else {
		if dbuf, ok := data.([]byte); ok {
			libhttp.ResponseDirect(w, r, dbuf, contentType)
			//if needEncrypt {
			//	libhttp.ResponseDirect(w, r, WebEncryptAfter(r, dbuf), contentType)
			//} else {
			//	libhttp.ResponseDirect(w, r, dbuf, contentType)
			//}
		} else {
			respBuf, err := json.Marshal(data)
			if err != nil {
				libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "system error").Error())
				return
			}
			libhttp.ResponseDirect(w, r, respBuf, contentType)
			//if needEncrypt {
			//	libhttp.ResponseDirect(w, r, WebEncryptAfter(r, respBuf), contentType)
			//} else {
			//	libhttp.ResponseDirect(w, r, respBuf, contentType)
			//}
		}
	}
}