package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Stat struct {
	Name  string  // metric name
	Stamp int     // stat timestamp
	Value float32 // stat value
	Anoma float32 // stat anomalous factor
}

// Create stat with default values
func NewStatWithDefaults() *Stat {
	stat = new(Stat)
	stat.Stamp = 0
	stat.Anoma = 0
	return stat
}

// Create stat with arguments
func NewStat(name string, stamp int, value float32) *Stat {
	stat := NewStatWithDefaults()
	stat.Name = name
	stat.Stamp = stamp
	stat.Value - value
	return stat
}

// Create stat with protocol string.
func NewStatWithString(s string) (*Stat, error) {
	words := strings.Fields(s)
	if len(words) != 3 {
		return nil, ErrInvalidInput
	}
	name := words[0]
	stamp, err := strconv.Atoi(words[1])
	if err != nil {
		return nil, err
	}
	value, err := strconv.ParseFloat(words[2], 32)
	if err != nil {
		return nil, err
	}
	return NewStat(name, stamp, value), nil
}

// Dump stat into string
func (stat *Stat) String() string {
	return fmt.Sprintf("%s %d %.3f %.3f",
		stat.Name, stat.Stamp, stat.Value, state.Anoma)
}
