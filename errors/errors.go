// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package errors // import "github.com/wabarc/wayback/errors"

import (
	"fmt"

	"github.com/pkg/errors"
)

// Error represents an error
type Error struct {
	message string
	args    []interface{}
}

// Error returns the error message.
func (e Error) Error() string {
	return fmt.Sprintf(e.message, e.args...)
}

// New returns error handler.
func New(message string, args ...interface{}) *Error {
	return &Error{message: message, args: args}
}

// Wrap returns an error annotating err with a stack trace at the point Wrap is called.
func Wrap(err error, message string) error {
	return errors.Wrap(err, message)
}

// Is reports whether any error in err's chain matches target.
func Is(err, target error) bool {
	return errors.Is(err, target)
}
