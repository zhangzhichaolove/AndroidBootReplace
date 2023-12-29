package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func(cancelFunc context.CancelFunc) {
		time.Sleep(10 * time.Second)
		fmt.Println("执行终止命令")
		cancelFunc()
	}(cancel)
	ExecCommand(ctx, "pwd", "cd tool", "ls", "java -version", "ls -al", "ping www.baidu.com")
}

func ExecCommand(ctx context.Context, command ...string) (*os.ProcessState, error) {
	//cmd := exec.Command("sh")
	cmd := exec.CommandContext(ctx, "sh")
	// 定义一对输入输出流
	inReader, inWriter := io.Pipe()
	// 把输入流的给到命令行
	cmd.Stdin = inReader
	// 获取标准输入流和错误信息流
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	sizeIndex := len(command) - 1
	// 指定用户执行
	osUser, err := user.Lookup(command[sizeIndex])
	if err == nil {
		uid, _ := strconv.Atoi(osUser.Uid)
		gid, _ := strconv.Atoi(osUser.Gid)
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}
	// 正常日志
	go func() {
		reader := bufio.NewReader(stdout)
		for {
			select {
			case <-ctx.Done():
				fmt.Println("程序终止")
				return
			default:
				readString, _, err := reader.ReadLine()
				if err != nil || err == io.EOF {
					return
				}
				fmt.Println("正常日志：", string(readString))
			}
		}
	}()
	// 错误日志
	go func() {
		reader := bufio.NewReader(stderr)
		for {
			select {
			case <-ctx.Done():
				fmt.Println("程序终止")
				return
			default:
				readString, _, err := reader.ReadLine()
				if err != nil || err == io.EOF {
					return
				}
				fmt.Println("错误日志: ", string(readString))
			}
		}
	}()
	// 写指令
	go func() {
		lines := command[:]
		go func() {
			time.Sleep(5 * time.Second)
			_, err = inWriter.Write([]byte("ls /Users\n"))
			if err != nil {
				fmt.Println("---->", err)
			}
			inWriter.Close()
		}()
		for i, str := range lines {
			_, err := inWriter.Write([]byte(str))
			if err != nil {
				fmt.Println(err)
			}
			_, err = inWriter.Write([]byte("\n"))
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("正在执行命令：", i, str)
			//if i == len(lines)-1 {
			//	_ = inWriter.Close()
			//}
		}
	}()

	err = cmd.Wait()
	state := cmd.ProcessState
	//执行失败，返回错误信息
	if !state.Success() {
		return state, err
	}
	fmt.Println("---->", "执行流程结束")
	return state, err
}

func ExecCommand2(ctx context.Context, command ...string) (*os.ProcessState, error) {
	var wg sync.WaitGroup
	wg.Add(2)

	cmd := exec.Command("sh")
	// 定义一对输入输出流
	inReader, inWriter := io.Pipe()
	// 把输入流的给到命令行
	cmd.Stdin = inReader
	// 获取标准输入流和错误信息流
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	sizeIndex := len(command) - 1
	// 指定用户执行
	osUser, err := user.Lookup(command[sizeIndex])
	if err == nil {
		//log.Printf("uid=%s,gid=%s", osUser.Uid, osUser.Gid)
		uid, _ := strconv.Atoi(osUser.Uid)
		gid, _ := strconv.Atoi(osUser.Gid)
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}
	// 正常日志
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("程序终止")
				return
			default:
				logScan := bufio.NewScanner(stdout)
				for logScan.Scan() {
					text := logScan.Text()
					fmt.Println("正常日志：", text)
				}
			}
		}
	}()
	// 错误日志
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("程序终止")
				return
			default:
				scan := bufio.NewScanner(stderr)
				for scan.Scan() {
					s := scan.Text()
					fmt.Println("错误日志: ", s)
				}
			}
		}
	}()
	// 写指令
	go func() {
		lines := command[:]
		go func() {
			time.Sleep(5 * time.Second)
			_, err = inWriter.Write([]byte("ls /user\n"))
			if err != nil {
				fmt.Println("---->", err)
			}
		}()
		for i, str := range lines {
			_, err := inWriter.Write([]byte(str))
			if err != nil {
				fmt.Println(err)
			}
			_, err = inWriter.Write([]byte("\n"))
			if err != nil {
				fmt.Println(err)
			}
			if i == len(lines)-1 {
				_ = inWriter.Close()
			}
		}
	}()

	err = cmd.Wait()
	state := cmd.ProcessState
	//执行失败，返回错误信息
	if !state.Success() {
		return state, err
	}

	wg.Wait()
	return state, err
}
