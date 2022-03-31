package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func main() {
	//< waiting for any device >
	//fastboot: usage: unknown command wait-for-device
	isConnect := make(chan bool)
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
	//execCommand("fastboot --version")
	execCommand("fastboot wait-for-device", isConnect)
	execRealTimeCommand("fastboot flash boot new.img")
	execRealTimeCommand("fastboot reboot")
	//execRealTimeCommand("ping www.baidu.com")
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
