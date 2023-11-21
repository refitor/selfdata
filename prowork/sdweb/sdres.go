package sdweb

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"rshub/product/rsapp/selfdata/prowork/common"

	"github.com/refitself/rslibs/libhttp"
)

// /api/file/preview: support file preview
func WebFilePreview(w http.ResponseWriter, r *http.Request) {
	DoWebFilePreview(w, r, func(params ...interface{}) {
		// params[0]: data, params[1]: fileName
		webServeResponse(w, r, params[0], "application/octet-stream", params[1])
	}, false)
}

// /api/file/view: support file view
func WebFileView(w http.ResponseWriter, r *http.Request) {
	DoWebFileView(w, r, func(params ...interface{}) {
		webServeResponse(w, r, params[0], fmt.Sprintf("%v", params[1]))
	}, false)
}

// /api/file/upload: support file upload
func WebFileUpload(w http.ResponseWriter, r *http.Request) {
	DoWebFileUpload(w, r, func(params ...interface{}) {
		file := params[0].(multipart.File)
		fullPath := params[1].(string)

		// save private cert
		dst, err := memfs.Create(fullPath)
		if err != nil {
			libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "system error").Error())
			return
		}
		defer dst.Close()

		//copy the uploaded file to the destination file
		if _, err := io.Copy(dst, file); err != nil {
			libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "system error").Error())
			return
		}
		webServeResponse(w, r, libhttp.WebResponse{Data: libhttp.C_Data_Success}, "application/json")
	}, false)
}

// /api/file/scan: support director scan
func WebFileScan(w http.ResponseWriter, r *http.Request) {
	DoWebFileScan(w, r, func(params ...interface{}) {
		webServeResponse(w, r, params[0].([]byte), params[1].(string))
	}, false)
}

// /api/file/view: support download vile
func DoWebFilePreview(w http.ResponseWriter, r *http.Request, handleFunc func(...interface{}), bForce bool) {
	directory := libhttp.WebParams(r).Get("directory")
	fpath := libhttp.WebParams(r).Get("filepath")
	if directory == "" {
		directory = decryptMemDir
	}

	file, ferr := memfs.Open(fpath)
	if ferr != nil {
		libhttp.ResponseError(w, r, "resource open failed, detail: "+ferr.Error())
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		libhttp.ResponseError(w, r, "resource read failed, detail: "+err.Error())
		return
	}

	fType, err := common.GetFileType(data)
	if err != nil {
		libhttp.ResponseError(w, r, "resource type read failed, detail: "+err.Error())
		return
	}
	fType = strings.Split(fType, ":")[0]
	// fType := common.GetFileTypeEx(data)
	common.DebugLog("file type: %s", fType)
	//bEncrypt := selfdata.EnableEncrypt(common.WebParams(r).Get("from"))

	// support file view
	isImg := strings.HasPrefix(fType, "image")
	isText := strings.HasPrefix(fType, "text")
	isPdf := strings.HasPrefix(fType, "application/pdf")
	isMedia := strings.HasPrefix(fType, "audio") || strings.HasPrefix(fType, "video")
	if !bForce || isText {
		if isImg {
			http.ServeContent(w, r, filepath.Base(fpath), time.Now(), file)
			return
		} else if isText || isPdf || isMedia {
			ruri := r.URL.RequestURI()
			ruri = strings.Replace(r.URL.RequestURI(), "/api/file/preview", "/#/", -1)
			ruri += "&ftype=" + html.EscapeString(fType)
			common.DebugLog("handle preview, %s ---> %s", r.URL.RequestURI(), ruri)
			//if v_GetConfig("remote") == "true" {
			//	common.ResponseOk(w, r, common.WebOk(ruri))
			//} else {
			//	http.Redirect(w, r, ruri, http.StatusFound)
			//}
			http.Redirect(w, r, ruri, http.StatusFound)
			return
		}
	}

	common.DebugLog("---------------%s", filepath.Base(fpath))
	if handleFunc != nil {
		handleFunc(data, filepath.Base(fpath))
	}
}

// /api/file/view: support view vile
func DoWebFileView(w http.ResponseWriter, r *http.Request, handleFunc func(...interface{}), bForce bool) {
	directory := libhttp.WebParams(r).Get("directory")
	//readonly := common.WebParams(r).Get("readonly")
	fpath := libhttp.WebParams(r).Get("filepath")
	ftype := libhttp.WebParams(r).Get("ftype")
	if directory == "" {
		directory = decryptMemDir
	}

	file, ferr := memfs.Open(fpath)
	if ferr != nil {
		libhttp.ResponseError(w, r, "resource open failed, detail: "+ferr.Error())
		return
	}
	defer file.Close()

	// unsupport pdf encrypt
	isPdf := strings.HasPrefix(ftype, "application/pdf")
	isMedia := strings.HasPrefix(ftype, "audio") || strings.HasPrefix(ftype, "video")
	if isPdf || isMedia  {
		http.ServeContent(w, r, filepath.Base(fpath), time.Now(), file)
		return
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		libhttp.ResponseError(w, r, "resource read failed, detail: "+err.Error())
		return
	}
	if handleFunc != nil {
		handleFunc(data, strings.Split(common.GetFileTypeEx(data), ":")[0])
	}
	//common.ResponseDirect(w, r, data, strings.Split(common.GetFileTypeEx(data), ":")[0])
}

func DoWebFileUpload(w http.ResponseWriter, r *http.Request, handleFunc func(...interface{}), bForce bool) {
	directory := libhttp.WebParams(r).Get("path")
	if directory == "" {
		directory = decryptMemDir
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		libhttp.ResponseError(w, r, libhttp.Error(err.Error(), "system error").Error())
		return
	}
	defer file.Close()
	fullPath := filepath.Join(directory, fileHeader.Filename)

	if handleFunc != nil {
		handleFunc(file, fullPath)
	}
}

func DoWebFileScan(w http.ResponseWriter, r *http.Request, handleFunc func(...interface{}), bForce bool) {
	directory := libhttp.WebParams(r).Get("directory")
	//kind := common.WebParams(r).Get("kind")
	if directory == "" {
		directory = decryptMemDir
	}
	authID := popFromSession(r)
	common.DebugLog("DoWebFileScan receive request successed, directory: %s", directory)

	ret := make(map[string]interface{}, 0)
	ret["type"] = "folder"
	ret["name"] = filepath.Base(directory)
	ret["path"] = filepath.Base(directory)
	ret["items"] = make([]interface{}, 0)

	ret, _ = readDir(directory, filepath.Dir(directory)+"/", ret)
	workItems := make([]interface{}, 0)
	for _, v := range ret["items"].([]interface{}) {
		if canScan(authID, fmt.Sprintf("%v", v.(map[string]interface{})["name"])) {
			workItems = append(workItems, v)
		}
	}
	ret["items"] = workItems
	buf, _ := json.Marshal(ret)
	common.DebugLog("buf: %v", string(buf))

	if handleFunc != nil {
		handleFunc(buf, "application/json")
	}
}

func readDir(dirPath, trimPrefix string, parentMap map[string]interface{}) (map[string]interface{}, error) {
	flist, err := readMemDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("readDir failed, dirPath: %s, detail: %s", dirPath, err.Error())
	}
	common.DebugLog("dirPath: %s, trimPrefix: %s, parentMap: %+v, len(flist): %v", dirPath, trimPrefix, parentMap, len(flist))

	if len(flist) == 0 {
		parentMap["items"] = make([]interface{}, 0)
	}
	for _, f := range flist {
		if f.IsDir() {
			var subErr error
			subMap := make(map[string]interface{}, 0)
			if subMap, subErr = readDir(dirPath+"/"+f.Name(), trimPrefix, subMap); subErr != nil {
				return nil, subErr
			} else {
				subMap["type"] = "folder"
				subMap["name"] = f.Name()
				subMap["size"] = f.Size()
				subMap["path"] = strings.TrimPrefix(dirPath+"/"+f.Name(), trimPrefix)

				if _, ok := parentMap["items"]; !ok {
					parentMap["items"] = make([]interface{}, 0)
				}
				items := parentMap["items"].([]interface{})
				items = append(items, subMap)
				parentMap["items"] = items
			}
		} else {
			if canScanFile(strings.TrimPrefix(dirPath+"/"+f.Name(), trimPrefix)) {
				subMap := make(map[string]interface{}, 0)
				subMap["type"] = "file"
				subMap["name"] = f.Name()
				subMap["size"] = f.Size()
				subMap["path"] = strings.TrimPrefix(dirPath+"/"+f.Name(), trimPrefix)
				if _, ok := parentMap["items"]; !ok {
					parentMap["items"] = make([]interface{}, 0)
				}
				items := parentMap["items"].([]interface{})
				items = append(items, subMap)
				parentMap["items"] = items
			}
		}
	}
	return parentMap, nil
}