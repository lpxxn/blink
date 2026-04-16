package sensitiveword

import "errors"

var ErrNotFound = errors.New("sensitiveword: not found")
var ErrDuplicateWord = errors.New("sensitiveword: duplicate word")
