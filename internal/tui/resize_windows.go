//go:build windows

package tui

func watchResize() (<-chan struct{}, func()) {
	return make(chan struct{}), func() {}
}
