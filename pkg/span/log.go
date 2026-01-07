package span

import "time"

type Log struct {
	Timestamp time.Time
	Fields    []Tag
}

func NewLog(timestamp time.Time, fields ...Tag) *Log {

}
