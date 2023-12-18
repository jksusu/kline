package utils

import "fmt"

type JSONData struct {
	data map[string]interface{}
}

func NewJSONData(data map[string]interface{}) *JSONData {
	json := &JSONData{data: make(map[string]interface{})}
	json.data = data
	return json
}

func (j *JSONData) Get(key string) interface{} {
	v, ok := j.data[key]
	if !ok {
		return ""
	}
	return v
}

func (j *JSONData) GetInt(key string) int {
	return ConvertStringToInt(fmt.Sprintf("%v", j.data[key]))
}
func (j *JSONData) GetInt64(key string) int64 {
	return ConvertStringToInt64(fmt.Sprintf("%v", j.data[key]))
}

func (j *JSONData) GetFloat64(key string) float64 {
	v := j.GetString(key)
	return ConvertStringToFloat64(v)
}

func (j *JSONData) GetString(key string) string {
	v := j.Get(key)
	return fmt.Sprintf("%v", v)
}
