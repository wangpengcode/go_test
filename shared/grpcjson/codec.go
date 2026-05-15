package grpcjson

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

// Name 是该 codec 的 gRPC content-subtype 名称（这里用 json）。
const Name = "json"

type codec struct{}

// Name 返回 codec 名称，gRPC 会用它来匹配 content-subtype。
func (codec) Name() string { return Name }

// Marshal 把 v 编码成 JSON。
func (codec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal 把 JSON 字节解码进 v（v 通常应该是一个指针）。
func (codec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// init 会把该 codec 注册到 gRPC 的全局编码器注册表里。
//
// 给刚接触 Go 的同学：
// - init() 会在 main() 之前自动执行。
// - main 里写 `_ "go_test/shared/grpcjson"`（空白导入）就是为了触发这里的 init()。
func init() {
	encoding.RegisterCodec(codec{})
}
