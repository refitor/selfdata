package common

import (
	"fmt"
	"runtime"
	"time"

	"github.com/refitself/rslibs/libhttp"
)

const (
	C_time      = "15:04:05"
	C_date      = "2006-01-02"
	C_date_ex   = "20060102"
	C_date_time = "2006-01-02 15:04:05"
	C_date_time_mill = "2006-01-02 15:04:05.999"
)

var (
	UpgradeCode = GetRandom(32, true)
	EnableShutdown = true
)

// ID: *TSDPublic
type TSDPublic struct {
	_ [0][]byte // 取消注释以阻止比较
	ID string
	Info string
	NetKey string
	DataKey string
}

type TSDMeta struct {
	_                  [0][]byte // 取消注释以阻止比较
	Name               string    // 节点名称
	Mode               string    // 当前模式
	AuthID             string    // 动态授权ID
	Relay              string    // relay节点url
	SysRight           string    // 系统权限
	StoreDir           string    // 存储目录
	Language           string    // 系统语言
	RelayPub           string    // relay节点公钥
	NetPublic          string    // 网络传输公钥
	NetPrivate         string    // 网络传输私钥
	DataPublic         string    // 数据加解密公钥
	DataPrivate        string    // 数据加解密私钥
	ListenPort         string    // 作为服务监听端口
	GoogleSecret       string    // google验证密钥
	AutoUpgrade		   bool

	PaidInfoMap		map[string]string     // 支付信息列表, kv ===> (authID + paidCode + rs.privateKey).signature: cmds
	SyncAuthList	map[string]*TSDPublic // 同步授权列表, kv ===> id: *TSDPublic
	VisitAuthList	map[string]string     // 访问授权列表, kv ===> ip: * | demo1.sd, demo2.sd
}

type T_SDDecrypt func(data []byte) []byte
type T_SDEncrypt func(data []byte) []byte
type T_SDAuth func() func(data []byte) error
type T_SDPull func(selfID string) (bool, []string, error)
type T_SDPush func(selfID, fileName string, data []byte) (bool, error)

type TResp struct {
	SDAuth T_SDAuth
	SDPush T_SDPush
	SDPull T_SDPull
	SDEncrypt T_SDEncrypt
	SDDecrypt T_SDDecrypt
}

// 私有数据
type SD struct {
	Name       string

	// 如果由其他selfdata同步的则为对方名称
	From       string

	Size       string
	UpdateTime string
	CreateTime string
}

type ListSD []*SD
func (p ListSD) Len() int           { return len(p) }
func (p ListSD) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ListSD) Less(i, j int) bool {
	it, _ := time.Parse(C_date_time, p[i].CreateTime)
	jt, _ := time.Parse(C_date_time, p[j].CreateTime)
	return it.Sub(jt).Nanoseconds() >= 0
}

func Error(err error) error {
	if err == nil {
		return nil
	}
	pc, _, line, _ := runtime.Caller(0)
	return fmt.Errorf("%s-%v===>%s", runtime.FuncForPC(pc).Name(), line, err.Error())
}

func InitLog() {


	libhttp.LogPrint = func(kind string, params ...interface{}) {
		switch kind {
		case "info":
			InfoLog(params...)
		case "debug":
			DebugLog(params...)
		case "error":
			ErrorLog(params...)
		}
	}
}

func GetPaidRight(service string) string {
	switch service {
	case "selfDataExtraStrong":
		return "strong"
	case "selfDataExtraSync":
		return "sync"
	case "selfDataExtraVisit":
		return "visit"
	case "selfDataExtraRelay":
		return "relay"
	}
	return ""
}
