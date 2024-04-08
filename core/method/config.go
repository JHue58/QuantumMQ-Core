package method

// CoreConfig QuantumMQ-Core的配置
type CoreConfig struct {
	// 消息队列初始容量
	QueueInitialCap uint64
	// 消息队列自动扩容容量
	QueueAutoExpandCap uint64
}

var coreConfig = generateCoreConfig()

func generateCoreConfig() CoreConfig {
	// TODO 从配置文件读取CoreConfig
	return CoreConfig{
		QueueInitialCap:    1024,
		QueueAutoExpandCap: 512,
	}
}

func GetCoreConfig() CoreConfig {
	return coreConfig
}
