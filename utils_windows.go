//go:build windows
// +build windows

package atreugo

func chmodFileToSocket(filepath string) error {
	return nil
}

func newPreforkServer(s *Atreugo) preforkServer {
	return newPreforkServerBase(s)
}
