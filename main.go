package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"wdr/utils"
)

const (
	ROOT_PATH   = "./"
	TIME_FORMAT = "2006-01-02 15:04:05"
)

var (
	command       string //备份的命令
	database_file string //数据库文件名
	begin_time    string //备份开始时间
	interval_time string //备份间隔时间
	save_days     int    //保存天数
)

func main() {
	flag.StringVar(&command, "command", "", "数据库备份命令")
	flag.StringVar(&database_file, "file-name", "test.sql", "数据库文件地址")
	flag.StringVar(&interval_time, "interval-time", "", "备份间隔时间，例：1h、1m2s")
	flag.IntVar(&save_days, "save-days", 15, "保存时间，单位：天")
	flag.StringVar(&begin_time, "begin-time", time.Now().Format(TIME_FORMAT), "备份开始时间，格式：2019-06-02 15:32:00")
	flag.Parse()
	if command == "" {
		log.Fatal("请输入数据库备份命令")
	}
	if database_file == "" {
		log.Fatal("请输入数据库备份文件路径")
	}
	if interval_time == "" {
		log.Fatal("请输入备份间隔时间")
	}
	interval_time, err := time.ParseDuration(interval_time)
	if err != nil {
		log.Fatal("备份间隔时间输入错误")
	}
	begin, err := time.ParseInLocation("2006-01-02 15:04:05", begin_time, time.Local)
	if err != nil {
		log.Fatal("启动时间错误")
	}
	start := begin.Unix() - time.Now().Unix()
	if start < 0 {
		start = 0
	}
	time.Sleep(time.Second * time.Duration(start))
	for {
		backup()
		time.Sleep(interval_time)
	}
}

//备份
func backup() error {
	defer func() {
		if err := recover(); err != nil {
			log.Println("遇到异常：", err)
		}
	}()
	//备份数据库
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		log.Println("备份失败：", err.Error())
		return err
	}
	//压缩数据库文件
	f, err := os.Open(database_file)
	if err != nil {
		log.Println("备份文件打开失败：", err.Error())
		return err
	}
	cmpfile, err := os.Create(fmt.Sprintf("%s.zip", time.Now().Format("2006-01-02-15-04-05")))
	if err != nil {
		log.Println("压缩文件创建失败：", err.Error())
		return err
	}
	if err := utils.ZipCompress([]*os.File{f}, cmpfile); err != nil {
		log.Println("文件压缩失败：", err.Error())
		return err
	}
	//删除备份文件
	if err := os.Remove(database_file); err != nil {
		log.Println("数据库文件删除失败：", err.Error())
		return err
	}
	//删除过期文件
	filepath.Walk(ROOT_PATH, func(file string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if info.ModTime().AddDate(0, 0, save_days).After(time.Now()) {
			return nil
		}
		if err := os.Remove(ROOT_PATH + file); err != nil {
			log.Println("过期文件删除失败：", err.Error())
			return err
		}
		return nil
	})
	log.Println("备份成功")
	return nil
}
