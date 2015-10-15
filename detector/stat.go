package detector

import (
	"fmt"
	"strconv"
	"strings"
)

type Stat struct {
	Name   string
	Stamp  int
	Value  float64
	Anoma  float64
	AvgOld float64
	AvgNew float64
}

func NewStatWithDefaults() *Stat {
	stat := new(Stat)
	stat.Stamp = 0
	stat.Anoma = 0
	return stat
}

func NewStatWithString(s string) (*Stat, error) {
	stat := NewStatWithDefaults()
	words := strings.Fields(s)
	if len(words) != 3 {
		return nil, ErrInvalidStatInput
	}
	name := words[0]
	stamp, err := strconv.Atoi(words[1])
	if err != nil {
		return nil, err
	}
	value, err := strconv.ParseFloat(words[2], 64)
	if err != nil {
		return nil, err
	}
	stat.Name = name
	stat.Stamp = stamp
	stat.Value = value
	return stat, nil
}

func (stat *Stat) String() string {
	return fmt.Sprintf("%s %d %.3f %.3f %.3f %.3f",
		stat.Name, stat.Stamp, stat.Value, stat.Anoma, stat.AvgOld, stat.AvgNew)
}
