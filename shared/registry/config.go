package registry

// Config 保存各个服务共用的 Consul 相关配置。
//
// 这里刻意保持“只放配置字段”：不写业务逻辑。
// 具体 Consul HTTP 调用在 consul.go。
type Config struct {
	// Enable 控制是否把自己注册到 Consul。
	Enable bool `mapstructure:"enable"`

	// ConsulAddr 是 Consul HTTP API 的地址，例如 "http://127.0.0.1:8500"。
	ConsulAddr string `mapstructure:"consul_addr"`

	// ServiceName 用于覆盖注册到 Consul 的服务名；为空时通常回退到 Service.Name。
	ServiceName string `mapstructure:"service_name"`

	// ServiceID 用于覆盖注册到 Consul 的实例 ID；为空时程序会自动生成一个。
	ServiceID string `mapstructure:"service_id"`

	// AdvertiseHost 是“对外暴露”的 host（其它服务连你时用这个地址）。
	// 这个地址需要能被其它服务访问到（比如局域网 IP、容器 hostname 等）。
	AdvertiseHost string `mapstructure:"advertise_host"`

	// AdvertisePort 是“对外暴露”的端口（其它服务连你时用这个端口）。
	AdvertisePort int `mapstructure:"advertise_port"`
}
