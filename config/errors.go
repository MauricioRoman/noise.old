// Copyright (c) 2015, Chao Wang <hit9@icloud.com>

package config

import "errors"

var (
	ErrDetectorFactor = errors.New("invalid detector.factor in config")
	ErrWebAuth        = errors.New("invalid web.auth in config")
)
