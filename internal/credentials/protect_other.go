//go:build !windows

package credentials

import "errors"

func protect([]byte) ([]byte, error)   { return nil, errors.New("credenciales soportadas en Windows") }
func unprotect([]byte) ([]byte, error) { return nil, errors.New("credenciales soportadas en Windows") }
