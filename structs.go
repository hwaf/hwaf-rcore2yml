package main

import (
	"github.com/hwaf/hwaf/hlib"
)

type ReqFile struct {
	Filename string
	Wscript  hlib.Wscript_t
}

func NewReqFile(name string) ReqFile {
	return ReqFile{
		Wscript: hlib.Wscript_t{
			Package: hlib.Package_t{
				Name:    name,
				Authors: []hlib.Author{"hwaf-rcore2yml"},
			},
		},
	}
}

// EOF
