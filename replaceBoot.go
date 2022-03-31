//go:generate go-bindata -o=./asset/bindata.go -pkg=asset tool/...
////go:generate go-bindata -fs -prefix "static/" tool/...
//go:generate go-bindata -version

package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	assetFs "github.com/elazarl/go-bindata-assetfs"
	asset "github.com/zhangzhichaolove/AndroidBootReplace/type"
	//"github.com/zhangzhichaolove/AndroidBootReplace/asset"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var imgName = "new.img"

//go get -u github.com/go-bindata/go-bindata/...
//go install github.com/go-bindata/go-bindata/...@latest
//go install -a -v github.com/go-bindata/go-bindata/...@latest
func main() {
	restore()
	if len(os.Args) == 2 {
		imgName = os.Args[1]
	} else {
		fmt.Println("默认会将同级目录下new.img替换到手机，你也可以手动指定该镜像名称，例如：./replaceBoot magisk_patched.img")
	}
	isConnect := make(chan bool)
	go func() {
		fs := assetFs.AssetFS{
			Asset:     asset.Asset,
			AssetDir:  asset.AssetDir,
			AssetInfo: asset.AssetInfo,
		}
		http.Handle("/", http.FileServer(&fs))
		http.ListenAndServe(":168", nil)
	}()
	go func() {
		for {
			select {
			case <-isConnect:
				return
			default:
				fmt.Println("请让手机进入fastboot模式，并连接数据线...")
			}
			time.Sleep(2 * time.Second)
		}
	}()
	execRealTimeCommand("tool/fastboot --version")
	execCommand("tool/fastboot wait-for-device", isConnect)
	execRealTimeCommand(fmt.Sprintf("tool/fastboot flash boot %s", imgName))
	execRealTimeCommand("tool/fastboot reboot")
	//execRealTimeCommand("ping www.baidu.com")
}

func restore() {
	if err := asset.RestoreAssets(".", ""); err != nil {
		fmt.Println("文件释放失败：", err.Error())
	}
}

func execCommand(strCommand string, connect chan bool) string {
	cmd := exec.Command("/bin/bash", "-c", strCommand)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "fastboot: usage: unknown command wait-for-device") {
			fmt.Println("手机连接成功，进入刷机模式...")
			connect <- true
		} else {
			fmt.Println("命令运行错误-->", string(output), err)
		}
		return ""
	}
	fmt.Println(string(output))
	return string(output)
}

//func execCommand(strCommand string) string {
//	cmd := exec.Command("/bin/bash", "-c", strCommand)
//	fmt.Println("执行命令：", cmd.Args)
//	stdout, err := cmd.StdoutPipe()
//	if err != nil {
//		fmt.Println("命令运行错误-->", err)
//		return ""
//	}
//	cmd.Start()
//	outText := ""
//	reader := bufio.NewReader(stdout)
//	//实时循环读取输出流中的一行内容
//	for {
//		line, err2 := reader.ReadString('\n')
//		if err2 != nil || io.EOF == err2 {
//			break
//		}
//		outText += line
//		fmt.Println(line)
//	}
//	cmd.Wait()
//	return outText
//}

func read(wg *sync.WaitGroup, std io.ReadCloser) {
	defer wg.Done()
	reader := bufio.NewReader(std)
	for {
		readString, err := reader.ReadString('\n')
		if err != nil || err == io.EOF {
			return
		}
		fmt.Print(readString)
	}
}

func execRealTimeCommand(cmd string) error {
	//c := exec.Command("cmd", "/C", cmd) 	// windows
	c := exec.Command("bash", "-c", cmd) // mac or linux
	stdout, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go read(&wg, stdout)
	go read(&wg, stderr)
	err = c.Start()
	wg.Wait()
	return err
}

func deCompress() {
	zipFile, err := zip.OpenReader("os.zip")
	if err != nil {
		fmt.Println("文件解压失败：", err.Error())
		return
	}
	defer zipFile.Close()
	for _, innerFile := range zipFile.File {
		info := innerFile.FileInfo()
		if info.IsDir() {
			continue
		}
		if info.Name() == "boot.img" {
			srcFile, _ := innerFile.Open()
			newFile, _ := os.Create(innerFile.Name)
			io.Copy(newFile, srcFile)
			newFile.Close()
			srcFile.Close()
			fmt.Println("文件解压完成：", innerFile.Name)
			return
		}
	}
}
