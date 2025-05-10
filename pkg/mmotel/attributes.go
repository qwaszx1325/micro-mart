// pkg/mmotel/attributes.go
package mmotel

import (
	"go.opentelemetry.io/otel/attribute"
)

// Field 是一個用於構建日誌/追蹤屬性的鍵值對
type Field struct {
	Key   string
	Value interface{}
}

// NewField 創建一個新的字段
func NewField(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// StringAttribute 創建字符串屬性
func StringAttribute(key, value string) attribute.KeyValue {
	return attribute.String(key, value)
}

// IntAttribute 創建整數屬性
func IntAttribute(key string, value int) attribute.KeyValue {
	return attribute.Int(key, value)
}

// Int64Attribute 創建 int64 屬性
func Int64Attribute(key string, value int64) attribute.KeyValue {
	return attribute.Int64(key, value)
}

// Float64Attribute 創建浮點數屬性
func Float64Attribute(key string, value float64) attribute.KeyValue {
	return attribute.Float64(key, value)
}

// BoolAttribute 創建布爾屬性
func BoolAttribute(key string, value bool) attribute.KeyValue {
	return attribute.Bool(key, value)
}
