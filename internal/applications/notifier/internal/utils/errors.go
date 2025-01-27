package utils

const (
	errPrefix = "internal/applications/messenger/internal/utils = "
)

type SUtilsError struct {
	str string
}

func (err *SUtilsError) Error() string {
	return errPrefix + err.str
}

var (
	ErrMessageSizeGteLimit = &SUtilsError{"message size >= limit"}
	ErrGetSettingsHLS      = &SUtilsError{"get settings hls"}
)
