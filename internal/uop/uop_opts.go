// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package uop

// Option defines a function that configures a Reader.
type Option func(*Reader)

// WithExtra sets a flag to indicate if extra data is present in the entries.
func WithExtra() Option {
	return func(r *Reader) {
		r.hasextra = true
	}
}

// WithExtension sets the file extension for the pattern.
func WithExtension(ext string) Option {
	return func(r *Reader) {
		r.ext = ext
	}
}

// WithLength sets the length of the index.
func WithLength(length int) Option {
	return func(r *Reader) {
		r.length = length
	}
}

// WithStrict sets a flag to indicate if the reader should perform strict entry validation.
func WithStrict() Option {
	return func(r *Reader) {
		r.strict = true
	}
}
