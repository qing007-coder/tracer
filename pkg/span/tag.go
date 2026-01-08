package span

type Tag struct {
	Key   string
	Value interface{}
	Type  string
}

func NewTag(key string, value interface{}) *Tag {
	return &Tag{Key: key, Value: value}
}
