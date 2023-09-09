package admin

import "io"

type IndexParameters struct {
	UserName string
}

func (s *service) RenderIndex(wr io.Writer, p IndexParameters) error {
	return s.t.ExecuteTemplate(wr, "index.html", struct {
		IndexParameters
		LoginURL string
	}{
		IndexParameters: p,
		LoginURL:        s.loginURL,
	})
}
