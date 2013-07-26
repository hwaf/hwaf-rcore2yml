package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func handle_err(err error) {
	if err != nil {
		panic(fmt.Errorf("hwaf-rcore2yml: %v", err))
	}
}

func main() {
	fmt.Printf("::: hwaf-rcore2yml\n")

	dir := "."
	switch len(os.Args) {
	case 1:
		dir = "."
	case 2:
		dir = os.Args[1]
	default:
		panic(fmt.Errorf("rcore2yml takes at most 1 argument (got %d)", len(os.Args)))
	}

	var err error
	//dir, err = filepath.Abs(dir)
	handle_err(err)

	fnames := []string{}
	fmt.Printf(">>> dir=%q\n", dir)
	if !path_exists(dir) {
		fmt.Printf("** no such file or directory [%s]\n", dir)
		os.Exit(1)
	}

	err = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		//fmt.Printf("::> [%s]...\n", path)
		if filepath.Base(path) != "Makefile.RootCore" {
			return nil
		} else {
			fnames = append(fnames, path)
			fmt.Printf("::> [%s]...\n", path)
		}
		return err
	})
	handle_err(err)

	if len(fnames) < 1 {
		fmt.Printf(":: hwaf-rcore2yml: no Makefile.RootCore file under [%s]\n", dir)
		os.Exit(0)
	}

	type Response struct {
		req *ReqFile
		err error
	}

	ch := make(chan Response)
	for _, fname := range fnames {
		go func(fname string) {
			reqfile, err := parse_file(fname)
			if err != nil {
				ch <- Response{
					reqfile,
					fmt.Errorf("err w/ file [%s]: %v", fname, err),
				}
				return
			}

			err = render_script(reqfile)
			if err != nil {
				ch <- Response{
					reqfile,
					fmt.Errorf("err w/ file [%s]: %v", fname, err),
				}
				return
			}
			ch <- Response{reqfile, nil}
		}(fname)
	}

	sum := 0
	allgood := true
loop:
	for {
		select {
		case resp := <-ch:
			sum += 1
			if resp.err != nil {
				fmt.Printf("**err: %v\n", resp.err)
				allgood = false
			}
			if sum == len(fnames) {
				close(ch)
				break loop
			}
		}
	}

	if !allgood {
		os.Exit(1)
	}
}

// EOF
