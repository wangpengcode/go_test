package main

import (
	"context"
	"flag"
	"os/signal"
	"path/filepath"
	"syscall"

	"go_test/go_test_api/internal/app"
	_ "go_test/shared/grpcjson"
	"go_test/shared/pathutil"
)

// main 是 go_test_api 的程序入口。
// 这里尽量只保留：命令行参数解析 + 优雅退出（Ctrl+C）。
// 具体的启动流程放到 internal/app 里。
func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "go_test_api/configs/config.yaml", "config file path")
	flag.Parse()

	if p, ok := pathutil.FirstExistingFile(
		configPath,
		"configs/config.yaml",
		filepath.Join("go_test_api", "configs", "config.yaml"),
	); ok {
		configPath = p
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if root, ok := pathutil.ModuleRootFrom(configPath); ok {
		configPath = pathutil.ResolveMaybeRelativeToRoot(configPath, root)
	}
	if err := app.Run(ctx, app.Options{ConfigPath: configPath}); err != nil {
		panic(err)
	}
}
