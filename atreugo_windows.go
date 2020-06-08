// +build windows

package atreugo

// ListenAndServe serves requests from the given network and address in the atreugo configuration.
//
// Pass custom listener to Serve/ServeGracefully if you want to use it.
func (s *Atreugo) ListenAndServe() error {
	ln, err := s.getListener()
	if err != nil {
		return err
	}

	return s.Serve(ln)
}
