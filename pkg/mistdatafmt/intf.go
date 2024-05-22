package mistdatafmt

type mistDataFmtIntf interface {
	GetJsonKeyValueAsStr(key string) (string, error)
	GetJsonKeyValueAsFloat64(key string) (float64, error)
	GetJsonKeyValueAsInt64(key string) (int64, error)
}
