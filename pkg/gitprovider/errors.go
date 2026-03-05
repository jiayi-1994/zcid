package gitprovider

import "errors"

var (
	ErrUnsupportedProvider = errors.New("unsupported git provider")
	ErrAuthFailed          = errors.New("authentication failed")
	ErrAPICall             = errors.New("git API call failed")
)
