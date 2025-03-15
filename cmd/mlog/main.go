package main

import (
	"fmt"

	"github.com/lwmacct/250300-go-template-pkgs/pkg/mlog"
)

func main() {

	// 使用默认配置（默认启用颜色）输出日志
	mlog.Info(mlog.H{"msg": "默认日志配置（启用颜色）"})
	// 创建一个禁用颜色的日志实例，并写入文件
	noColorLogger := mlog.New(
		mlog.WithColor(false),
		mlog.WithFile("../../.local/log/no_color.log"),
	)

	fmt.Printf("Color setting for noColorLogger: %v\n", noColorLogger.GetColor())
	noColorLogger.Info(mlog.H{"msg": "禁用颜色的日志"})

	// 创建一个启用颜色的日志实例，并写入文件
	colorLogger := mlog.New(
		mlog.WithColor(true),
		mlog.WithFile("../../.local/log/with_color.log"),
	)

	fmt.Printf("Color setting for colorLogger: %v\n", colorLogger.GetColor())
	colorLogger.Info(mlog.H{"msg": "明确启用颜色的日志"})

	mlog.SetNew(
		mlog.WithFile("../../.local/log/setNew.log"),
		mlog.WithColor(false),
	)
	mlog.Info(mlog.H{"msg": "使用setNew配置"})

	// 确保在程序退出前等待所有日志处理完成
	// 首先关闭自定义日志实例
	noColorLogger.Close()
	colorLogger.Close()
	// 最后关闭全局实例
	mlog.Close()
}
