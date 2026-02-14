package othelp

import "go.opentelemetry.io/otel/attribute"

// Str creates a string attribute.
func Str(key, value string) attribute.KeyValue {
	return attribute.String(key, value)
}

// Int creates an int attribute.
func Int(key string, value int) attribute.KeyValue {
	return attribute.Int(key, value)
}

// Int64 creates an int64 attribute.
func Int64(key string, value int64) attribute.KeyValue {
	return attribute.Int64(key, value)
}

// Float64 creates a float64 attribute.
func Float64(key string, value float64) attribute.KeyValue {
	return attribute.Float64(key, value)
}

// Bool creates a bool attribute.
func Bool(key string, value bool) attribute.KeyValue {
	return attribute.Bool(key, value)
}

// Strs creates a string slice attribute.
func Strs(key string, values []string) attribute.KeyValue {
	return attribute.StringSlice(key, values)
}

// Ints creates an int slice attribute.
func Ints(key string, values []int) attribute.KeyValue {
	return attribute.IntSlice(key, values)
}
