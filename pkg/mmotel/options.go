// pkg/mmotel/options.go
package mmotel

// ExporterType 定義導出器類型
type ExporterType string

const (
	ExporterTypeJaeger ExporterType = "jaeger"
	ExporterTypeStdout ExporterType = "stdout"
)

// Options 包含追蹤器的所有配置選項
type Options struct {
	// 服務信息
	ServiceName    string
	ServiceVersion string
	Environment    string

	// 採樣配置
	SamplingRatio float64

	// 導出器配置
	ExporterType   ExporterType
	JaegerEndpoint string // Jaeger OTLP 端點，例如 "localhost:4317"

	// 批次處理配置
	BatchTimeout  int // 批次超時，單位毫秒
	BatchSize     int // 批次大小
	MaxExportSize int // 最大導出大小
}

// Option 是一個函數類型，用於修改 Options
type Option func(*Options)

// 設置默認選項
func defaultOptions() *Options {
	return &Options{
		ServiceVersion: "0.1.0",
		Environment:    "development",
		SamplingRatio:  1.0, // 默認全採樣
		ExporterType:   ExporterTypeJaeger,
		JaegerEndpoint: "localhost:4317",
		BatchTimeout:   5000, // 5秒
		BatchSize:      512,  // 默認批次大小
		MaxExportSize:  1024, // 最大導出大小
	}
}

// WithServiceVersion 設置服務版本
func WithServiceVersion(version string) Option {
	return func(o *Options) {
		o.ServiceVersion = version
	}
}

// WithEnvironment 設置環境
func WithEnvironment(env string) Option {
	return func(o *Options) {
		o.Environment = env
	}
}

// WithSamplingRatio 設置採樣比率 (0.0-1.0)
func WithSamplingRatio(ratio float64) Option {
	return func(o *Options) {
		if ratio < 0.0 {
			ratio = 0.0
		}
		if ratio > 1.0 {
			ratio = 1.0
		}
		o.SamplingRatio = ratio
	}
}

// WithJaegerExporter 配置 Jaeger 導出器
func WithJaegerExporter(endpoint string) Option {
	return func(o *Options) {
		o.ExporterType = ExporterTypeJaeger
		o.JaegerEndpoint = endpoint
	}
}

// WithStdoutExporter 使用標準輸出導出器 (通常用於開發和測試)
func WithStdoutExporter() Option {
	return func(o *Options) {
		o.ExporterType = ExporterTypeStdout
	}
}

// WithBatchConfig 設置批次處理配置
func WithBatchConfig(timeout, size, maxExport int) Option {
	return func(o *Options) {
		if timeout > 0 {
			o.BatchTimeout = timeout
		}
		if size > 0 {
			o.BatchSize = size
		}
		if maxExport > 0 {
			o.MaxExportSize = maxExport
		}
	}
}
