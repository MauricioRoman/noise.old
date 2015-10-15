// Copyright (c) 2015, Chao Wang <hit9@ele.me>

package config

import "errors"

var (
	ErrDetectorFactor = errors.New("invalid detector.factor in config")
	ErrWebAuth        = errors.New("invalid web.auth in config")
)
