package main

import (
	"fmt"

	"github.com/hwaf/hwaf/hlib"
)

func (r *Renderer) render_hscript() error {
	var err error

	_, err = fmt.Fprintf(
		r.w,
		"## automatically generated by rcore2yml\n## do NOT edit\n\n",
	)
	handle_err(err)

	enc := hlib.NewHscriptYmlEncoder(r.w)
	if enc == nil {
		return fmt.Errorf("rcore2yml: got nil hlib.HscriptYmlEncoder")
	}

	err = enc.Encode(&r.pkg)
	return err
}
