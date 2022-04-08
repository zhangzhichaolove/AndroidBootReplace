package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("请指定系统镜像路径，例如：./extractBoot miui.zip")
		return
	}
	deCompress(os.Args[1])
}

func deCompress(file string) {
	zipFile, err := zip.OpenReader(file)
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
