package pgxstore

import "errors"

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")
var ErrNotEnoughBalance = errors.New("not enough balance")
