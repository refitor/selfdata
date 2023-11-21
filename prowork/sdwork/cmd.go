package sdwork

import (
	"context"
	"fmt"
	rice "github.com/GeertJohan/go.rice"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"rshub/product/rsapp/selfdata/prowork/common"
)

// function for others
var (
	V_Web_Init = func(encryptDir, authID string) error { return nil }
	V_Web_Run  = func(ctx context.Context, box *rice.Box, port string) string { return "" }
)

// var for cmd
var fPlugin, fOutput string
var vID, vSysPriv, vSysPub, vMetaPath string

var vSupport = "read, write, web"
var vSupportList = make([]string, 0)

func Run() {
	common.InfoLog("\nVersion: %s\nNodeID: %s\nSupport: %s", vVersion, selfMeta.Name, strings.Join(vSupportList, ", "))
	var cmdRead = &cobra.Command{
		Use:   "read [path]",
		Short: "Read private data and decrypt to the specified directory",
		Long:  `read is used to read private data to the specified directory, the default is ./sd/nodeID/decryption, demo: selfdata read output.sd.`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runForDAuth(cmd, args)
			runForRead(cmd, args)
		},
	}

	var cmdWrite = &cobra.Command{
		Use:   "write [path]",
		Short: "Write directory or file resources to generate private data",
		Long:  `write is used to write to the specified directory or file resource to generate private data, the default is ./sd/nodeID/encryption/output-timestamp.sd, demo: selfdata write ./datas1 ./datas2, selfdata write ./res1.txt ./res2.txt, selfdata write ./datas1 ./res1.txt.`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runForDAuth(cmd, args)
			runForWrite(cmd, args)
		},
	}

	//var cmdRun = &cobra.Command{
	//	Use:   "run [cmd]",
	//	Short: "Run selfdata as a service for private data management",
	//	Long: `run is used to open private data services, support relay transfer and web visual management and other forms, demo: selfdata run web | relay.`,
	//	Args: cobra.MinimumNArgs(0),
	//	Run:  func(cmd *cobra.Command, args []string) {
	//		//runForDAuth(cmd, args)
	//		runForServe(cmd, args)
	//	},
	//}
	//
	//var cmdUpdate = &cobra.Command{
	//	Use:   "update",
	//	Short: "Update the local private metadata of the current node",
	//	Long: `update is used to update the current node meta information, including node name, dynamic authorization ID, communication public and private key pair, encryption and decryption public and private key pair, etc, demo: selfdata update.`,
	//	Args: cobra.MinimumNArgs(0),
	//	Run: func(cmd *cobra.Command, args []string) {
	//		runForDAuth(cmd, args)
	//		runForUpdate(cmd, args)
	//	},
	//}

	var rootCmd = &cobra.Command{
		Use:     "selfdata",
		Example: "  selfdata | " + strings.Join(vSupportList, " | "),
		Short:   "\nSelfData: Designed for high-strength security reinforcement for private data",
		Run: func(cmd *cobra.Command, args []string) {
			//runForDAuth(cmd, args)
			runForServe(cmd, args)
		},
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if HasRight(C_Right_Strong) {
		rootCmd.Flags().StringVar(&fPlugin, "plugin", "", "Specify a custom plug-in, the default is empty, --plugin=./sdcore.so")
	}
	if strings.Contains(vSupport, C_Cmd_Read) {
		//cmdRead.Flags().StringVar(&fMode, "mode", "", "Specify the current mode, offline supports offline operation, sync supports automatic synchronization, --mode=offline|sync")
		//cmdRead.Flags().StringVar(&fPlugin, "plugin", "", "Specify a custom plug-in, the default is empty, --plugin=./sdcore.so")
		cmdRead.Flags().StringVar(&fOutput, "output", "", "Specify the directory for the decrypted data, --output=output")
		cmdRead.Flags().AddFlagSet(rootCmd.Flags())
		rootCmd.AddCommand(cmdRead)
	}
	if strings.Contains(vSupport, C_Cmd_Write) {
		//cmdWrite.Flags().StringVar(&fMode, "mode", "", "Specify the current mode, offline supports offline operation, sync supports automatic synchronization, --mode=offline|sync")
		//cmdWrite.Flags().StringVar(&fPlugin, "plugin", "", "Specify a custom plug-in, the default is empty, --plugin=./sdcore.so")
		cmdWrite.Flags().StringVar(&fOutput, "output", "", "Specify the name of the private data generation output, --output=output")
		cmdWrite.Flags().AddFlagSet(rootCmd.Flags())
		rootCmd.AddCommand(cmdWrite)
	}
	//if strings.Contains(vSupport, C_Cmd_Run) {
	//	//cmdRun.Flags().StringVar(&fMode, "mode", "relay", "Specify the current mode, relay supports transfer services, web supports visual management, --mode=relay|web")
	//	//cmdRun.Flags().StringVar(&fRelayPort, "port", "5037", "Specify the listening port of the private data transfer service, the default is 5037, 示例: --port=5037")
	//	rootCmd.AddCommand(cmdRun)
	//}
	//if strings.Contains(vSupport, C_Cmd_Update) {
	//	cmdUpdate.Flags().StringVar(&fName, "name", "", "Specify the name of the private data node, the default is the node id, --name=demo")
	//	cmdUpdate.Flags().StringVar(&fSelfAuth, "auth", "", "Specify the private data dynamic authorization ID, the default is empty, --auth=example@163.com")
	//	cmdUpdate.Flags().StringVar(&fStore, "store", "sd", "Specify the private data storage directory, the default is the sd directory of the current folder, --output=sd")
	//	cmdWrite.Flags().StringVar(&fMode, "mode", "standard", "Specify the current mode, offline supports offline operation, sync supports automatic synchronization, --mode=offline|sync")
	//	//// en-clientID
	//	//// de-clientID
	//	//// meta.sd, sdbase.so
	//	//cmdUpdate.Flags().StringVar(&fPaidCode, "paid", "", "Enter the payment unique authorization code to open the corresponding permissions, the default is blank, --paid=demo")
	//	//cmdUpdate.Flags().StringVar(&fUpgrade, "upgrade", "true", "Whether to enable automatic update at startup, the default is true, --upgrade=true")
	//	//cmdUpdate.Flags().StringVar(&fSelfAuth, "auth", "", "Specify the private data dynamic authorization ID, the default is empty, --auth=example@163.com")
	//	//cmdUpdate.Flags().StringVar(&fSelfPriv, "private", "", "Specify the user's private key to read private data, the default is empty, --private=privateKey.pem")
	//	//cmdUpdate.Flags().StringVar(&fSelfPub, "public", "", "Specify the user's public key to generate private data, the default is empty, --public=publicKey.pem")
	//	//cmdUpdate.Flags().StringVar(&fRelayPub, "relayPublic", "", "Specify the public key of the selfdata relay node, the default is empty, --relayPublic=publicKey.pem")
	//	//cmdUpdate.Flags().StringVar(&fRelay, "relay", "https://relay.refitself.cn", "Specify the private data relay node, the default is empty, --relay=https://example.com")
	//	rootCmd.AddCommand(cmdUpdate)
	//}
	PanicShutdown(rootCmd.Execute())
}

func runForDAuth(cmd *cobra.Command, args []string) {
	//if common.IsDebug() {
	//	return
	//}
	isInit := selfMeta.AuthID == ""
	if !isInit && selfMeta != nil {
		common.InfoLog("\n===>Dynamic authorization: %s", selfMeta.AuthID)
		strFailed := "Failed to operate private data, authorization exception: %s"
		strSuccessed := "Dynamic authorization succeeded"
		// if _, authErr := AuthByGoogle(selfMeta.GoogleSecret); authErr != nil {
		if _, authErr := AuthByEmail(C_Url_Auth, selfMeta.AuthID, selfMeta.Mode, []byte(GetRSPublic())); authErr != nil {
			PanicShutdown(fmt.Errorf(strFailed, authErr.Error()))
		}
		common.InfoLog(strSuccessed)

		if strings.Contains(strings.Join(vSupportList, ", "), C_Right_Strong) {
			// 授权成功的情况下加载自定义插件
			initForSDCore(fPlugin)
		}
	}
}

// read: /path/input.sd ===> ./sd/ID/decryption
func runForRead(cmd *cobra.Command, args []string) {
	// read selfdata
	if fOutput == "" {
		fOutput = filepath.Join("sd", C_Prefix_De, C_Prefix_Self)
	}
	common.InfoLog("\n===>Private data decryption")

	sdFilePaths := args[:]
	for _, fpath := range sdFilePaths {
		outputResult := SDDecryptFile(fpath, fOutput, []byte(selfMeta.DataPrivate), []byte(selfMeta.DataPublic), nil, nil)
		common.InfoLog("decrypt: %s", outputResult)
	}
	common.InfoLog("===>Private data decryption, amount: %v", len(sdFilePaths))
}

// write: ./datas ===> ./sd/selfID/encryption/output-时间戳.sd
func runForWrite(cmd *cobra.Command, args []string) {
	// write
	if fOutput == "" {
		fOutput = "output"
	}
	resArr := []string{"datas"}
	if len(args) >= 1 {
		resArr = args[:]
	}
	common.InfoLog("\n===>Private data encryption")

	for _, res := range resArr {
		workOutput := fOutput //fmt.Sprintf("%v-%v", fOutput, common.IDByTimeNow())
		outputResult, _ := SDEncryptFile(workOutput, []byte(selfMeta.DataPrivate), []byte(selfMeta.DataPublic), nil, nil, res)
		common.InfoLog("encrypt: %s", outputResult)
	}
	common.InfoLog("===>Private data encryption, amount: %v", len(resArr))
}

func runForServe(cmd *cobra.Command, args []string) {
	common.InfoLog("\n===>Run selfdata server")
	PanicShutdown(V_Web_Init(fmt.Sprintf("%s/%s", selfMeta.StoreDir, C_Prefix_En), selfMeta.AuthID))

	common.EnableShutdown = false
	visitURL := V_Web_Run(cmd.Context(), vBox, selfMeta.ListenPort)
	common.InfoLog("===>Successfully run selfdata server: %s", visitURL)
}
