package handler

type FieldInfo struct {
	Name        string
	Type        string
	JSONName    string
	Location    string 
	Required    bool
	Description string
	Example     string
	Format      string
}

type EnhancedParamInfo struct {
	ParamInfo
	Fields    []FieldInfo 
	SubParams []ParamInfo 
}

var TypeMapping = map[string]struct {
	Type   string
	Format string
}{
	"string":  {"string", ""},
	"int":     {"integer", "int64"},
	"int8":    {"integer", "int32"},
	"int16":   {"integer", "int32"},
	"int32":   {"integer", "int32"},
	"int64":   {"integer", "int64"},
	"uint":    {"integer", "int64"},
	"uint8":   {"integer", "int32"},
	"uint16":  {"integer", "int32"},
	"uint32":  {"integer", "int32"},
	"uint64":  {"integer", "int64"},
	"float32": {"number", "float"},
	"float64": {"number", "double"},
	"bool":    {"boolean", ""},
	"byte":    {"string", "byte"},
	"time.Time": {"string", "date-time"},
}