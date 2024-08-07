package bobrix

import "errors"

var (
	ErrInappropriateMimeType = errors.New("inappropriate MIME type of audiofile")
	ErrDownloadFile          = errors.New("failed to download audiofile")
	ErrParseMXCURI           = errors.New("failed to parse MXC URI")
)
