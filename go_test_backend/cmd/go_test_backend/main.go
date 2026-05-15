package main

import (
	"context"
	"flag"
	"os/signal"
	"path/filepath"
	"syscall"

	"go_test/go_test_backend/internal/app"
	_ "go_test/shared/grpcjson"
	"go_test/shared/pathutil"
)

// main 是 go_test_backend 的程序入口。
// 这里尽量只保留：命令行参数解析 + 优雅退出（Ctrl+C）相关的代码。
// 具体的启动流程（配置、DB、gRPC、Consul 等）放到 internal/app 里。
func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "go_test_backend/configs/config.yaml", "config file path")
	flag.Parse()

	if p, ok := pathutil.FirstExistingFile(
		configPath,
		"configs/config.yaml",
		filepath.Join("go_test_backend", "configs", "config.yaml"),
	); ok {
		configPath = p
	}

	// signal.NotifyContext 会把系统信号（SIGINT/SIGTERM）转换成 ctx 取消。
	// 当你按 Ctrl+C 时，ctx.Done() 会被关闭，app.Run 就能优雅退出。
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if root, ok := pathutil.ModuleRootFrom(configPath); ok {
		configPath = pathutil.ResolveMaybeRelativeToRoot(configPath, root)
	}
	if err := app.Run(ctx, app.Options{ConfigPath: configPath}); err != nil {
		// main 包里没有创建 logger，这里直接 panic 让错误明显暴露出来。
		// 真正线上项目可以换成标准库 log 打印后退出。
		panic(err)
	}
}
