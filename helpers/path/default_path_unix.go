// +build -windows

package path_helpers

func NewDefaultPath() Path {
	return NewUnixPath()
}
