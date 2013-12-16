package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hwaf/hwaf/hlib"
)

const dbg_parse_line = false

func fmt_line(data []string) string {
	s := bytes.NewBufferString("[")
	for i, v := range data {
		if i == 0 {
			fmt.Fprintf(s, "%q", v)
		} else {
			fmt.Fprintf(s, ", %q", v)
		}
	}
	fmt.Fprintf(s, "]")
	return string(s.Bytes())
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func scan_line(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance, token, err = bufio.ScanLines(data, atEOF)
	return
	// sz := len(token)
	// if sz > 0 && token[sz-1] == '\\' {
	// 	return
	// }
}

type Parser struct {
	req *ReqFile

	table   map[string]ParseFunc
	f       *os.File
	scanner *bufio.Scanner
	tokens  []string
}

func (p *Parser) Close() error {
	if p.f == nil {
		return nil
	}
	err := p.f.Close()
	p.f = nil
	return err
}

func NewParser(fname string) (*Parser, error) {

	var err error
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bufio.NewReader(f))
	if scanner == nil {
		return nil, fmt.Errorf("rcore2yml: nil bufio.Scanner")
	}

	p := &Parser{
		table:   g_dispatch,
		f:       f,
		scanner: scanner,
		req: &ReqFile{
			Filename: fname,
			Wscript:  hlib.Wscript_t{},
		},
		tokens: nil,
	}
	return p, nil
}

func (p *Parser) run() error {
	var err error

	bline := []byte{}
	for p.scanner.Scan() {
		data := p.scanner.Bytes()
		data = bytes.TrimSpace(data)
		data = bytes.Trim(data, " \t\r\n")

		if len(data) == 0 {
			continue
		}

		if data[0] == '#' {
			continue
		}

		idx := len(data) - 1
		if data[idx] == '\\' {
			bline = append(bline, ' ')
			bline = append(bline, data[:idx-1]...)
			continue
		} else {
			bline = append(bline, ' ')
			bline = append(bline, data...)
		}

		var tokens []string
		tokens, err = parse_line(bline)
		if err != nil {
			return err
		}
		p.tokens = tokens

		fct, ok := p.table[p.tokens[0]]
		if !ok {
			fmt.Fprintf(os.Stderr, "**warning** rcore2yml: unknown token [%v]\n", tokens[0])
			bline = nil
			continue
		}
		err = fct(p)
		if err != nil {
			return err
		}
		bline = nil
	}

	return err
}

func parse_line(data []byte) ([]string, error) {
	var err error
	line := []string{}

	worder := bufio.NewScanner(bytes.NewBuffer(data))
	worder.Split(bufio.ScanWords)
	tokens := []string{}
	for worder.Scan() {
		tok := worder.Text()
		if tok != "" {
			tokens = append(tokens, worder.Text())
		}
	}

	if strings.HasSuffix(tokens[0], "=") {
		slice := strings.Split(tokens[0], "=")
		tokens[0] = slice[0]
	}

	if len(tokens) > 1 && tokens[1] == "=" {
		tokens = append(tokens[:1], tokens[2:]...)
	}

	my_printf := func(format string, args ...interface{}) (int, error) {
		return 0, nil
	}
	if dbg_parse_line {
		my_printf = func(format string, args ...interface{}) (int, error) {
			return fmt.Printf(format, args...)
		}
	}

	my_printf("===============\n")
	my_printf("tokens: %v\n", fmt_line(tokens))

	in_dquote := false
	in_squote := false
	for i := 0; i < len(tokens); i++ {
		tok := strings.Trim(tokens[i], " \t")
		my_printf("tok[%d]=%q    (q=%v)\n", i, tok, in_squote || in_dquote)
		if strings.HasPrefix(tok, "#") {
			break
		}
		if in_squote || in_dquote {
			if len(line) > 0 {
				ttok := tok
				if strings.HasPrefix(ttok, `"`) || strings.HasPrefix(ttok, "'") {
					ttok = ttok[1:]
				}
				if strings.HasSuffix(ttok, `"`) || strings.HasSuffix(ttok, "'") {
					if !strings.HasSuffix(ttok, `\"`) {
						ttok = ttok[:len(ttok)-1]
					}
				}
				ttok = strings.Trim(ttok, " \t")
				if len(ttok) > 0 {
					line_val := line[len(line)-1]
					line_sep := ""
					if len(line_val) > 0 {
						line_sep = " "
					}
					ttok = strings.Replace(ttok, `\"`, `"`, -1)
					line[len(line)-1] += line_sep + ttok
				}
			} else {
				panic("logic error")
			}
		} else {
			ttok := tok
			if strings.HasPrefix(ttok, `"`) || strings.HasPrefix(ttok, "'") {
				ttok = ttok[1:]
			}
			if strings.HasSuffix(ttok, `"`) || strings.HasSuffix(ttok, "'") {
				if !strings.HasSuffix(ttok, `\"`) {
					ttok = ttok[:len(ttok)-1]
				}
			}
			ttok = strings.Replace(ttok, `\"`, `"`, -1)
			line = append(line, strings.Trim(ttok, " \t"))
		}
		if len(tok) == 1 && strings.HasPrefix(tok, "\"") {
			in_dquote = !in_dquote
			continue
		}
		if len(tok) == 1 && strings.HasPrefix(tok, "'") {
			in_squote = !in_squote
			continue
		}
		if strings.HasPrefix(tok, "\"") && !strings.HasSuffix(tok, "\"") {
			in_dquote = !in_dquote
			my_printf("--> dquote: %v -> %v\n", !in_dquote, in_dquote)
		}
		if strings.HasPrefix(tok, "'") && !strings.HasSuffix(tok, "'") {
			in_squote = !in_squote
			my_printf("--> squote: %v -> %v\n", !in_squote, in_squote)
		}
		if in_dquote && strings.HasSuffix(tok, "\"") && !strings.HasSuffix(tok, `\""`) {
			in_dquote = !in_dquote
			my_printf("<-- dquote: %v -> %v\n", !in_dquote, in_dquote)
		}
		if in_squote && strings.HasSuffix(tok, "'") {
			in_squote = !in_squote
			my_printf("<-- squote: %v -> %v\n", !in_squote, in_squote)
		}
	}

	return line, err
}

func parse_file(fname string) (*ReqFile, error) {
	fmt.Printf("req=%q\n", fname)
	p, err := NewParser(fname)
	if err != nil {
		return nil, err
	}
	defer p.Close()

	err = p.run()
	if err != nil {
		fmt.Printf("req=%q [ERR]\n", fname)
		return nil, err
	}

	fmt.Printf("req=%q [done]\n", fname)
	return p.req, err
}

type ParseFunc func(p *Parser) error

var g_dispatch = map[string]ParseFunc{
	"PACKAGE":          parsePackage,
	"PACKAGE_BINFLAGS": parseBinFlags,
	"PACKAGE_CLEAN":    parseClean,
	"PACKAGE_CXXFLAGS": parseCxxFlags,
	"PACKAGE_DEP":      parseDep,
	"PACKAGE_LDFLAGS":  parseLdFlags,
	"PACKAGE_NOCC":     parseNoCC,
	"PACKAGE_NOOPT":    parseNoOpt,
	"PACKAGE_OBJFLAGS": parseObjFlags,
	"PACKAGE_PEDANTIC": parsePedantic,
	"PACKAGE_PRELOAD":  parsePreload,
	"PACKAGE_REFLEX":   parseReflex,
	"PACKAGE_TRYDEP":   parseTryDep,
	"include":          parseInclude,
	"ifneq":            parseIfNeq,
	"endif":            parseEndIf,
}

func parsePackage(p *Parser) error {
	var err error
	tokens := p.tokens
	name := tokens[1]

	p.req.Wscript.Package = hlib.Package_t{
		Name: name,
		Authors: []hlib.Author{
			"hwaf-rcore2yml",
		},
		Deps: []hlib.Dep_t{
			{
				Name: "AtlasPolicy",
				Type: hlib.PublicDep,
			},
			{
				Name: "External/AtlasROOT",
				Type: hlib.PublicDep,
			},
		},
	}
	p.req.Wscript.Build.Targets = append(
		p.req.Wscript.Build.Targets,
		hlib.Target_t{
			Name:     name,
			Features: []string{"atlas_library", "atlas_dictionary"},
			Target:   name,
			Source: []hlib.Value{
				hlib.DefaultValue("source",
					[]string{
						"Root/*.cxx",
						fmt.Sprintf("%s/%sDict.h", name, name),
					},
				),
			},
			Use: []hlib.Value{
				hlib.DefaultValue(
					"use",
					[]string{
						"ROOT",
					},
				),
			},
			KwArgs: map[string][]hlib.Value{
				"rootcint_linkdef": []hlib.Value{
					hlib.DefaultValue("linkdef", []string{"Root/LinkDef.h"}),
				},
			},
		},
	)

	progs := []string{}
	dir := filepath.Join(filepath.Dir(filepath.Dir(p.f.Name())), "util")
	err = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		//fmt.Printf("::> [%s]...\n", path)
		fname := filepath.Base(path)
		if !strings.HasSuffix(fname, ".cxx") {
			return nil
		} else {
			progs = append(progs, fname[:len(fname)-len(".cxx")])
			//fmt.Printf("::> [%s]...\n", path)
		}
		return err
	})

	for _, prog := range progs {
		p.req.Wscript.Build.Targets = append(
			p.req.Wscript.Build.Targets,
			hlib.Target_t{
				Name:     prog,
				Features: []string{"atlas_application"},
				Target:   prog,
				Source: []hlib.Value{
					hlib.DefaultValue("source",
						[]string{
							filepath.Join("util", fmt.Sprintf("%s.cxx", prog)),
						},
					),
				},
				Use: []hlib.Value{
					hlib.DefaultValue(
						"use",
						[]string{
							name,
							"ROOT",
						},
					),
				},
			},
		)
	}
	return err
}

func parseObjFlags(p *Parser) error {
	var err error
	return err
}

func parseBinFlags(p *Parser) error {
	var err error
	return err
}

func parseCxxFlags(p *Parser) error {
	var err error
	tokens := p.tokens
	wtgt := &p.req.Wscript.Build.Targets[0]
	cxxflags := []string{}
	defines := []string{}
	includes := []string{}

	for _, flag := range tokens[1:] {
		flag = strings.Trim(flag, " \t")
		if len(flag) == 0 {
			continue
		}
		if strings.HasPrefix(flag, "-D") {
			defines = append(defines, flag[len("-D"):])
		} else if strings.HasPrefix(flag, "-I") {
			includes = append(includes, flag[len("-I"):])
		} else {
			cxxflags = append(cxxflags, flag)
		}
	}
	if len(cxxflags) > 0 {
		wtgt.CxxFlags = []hlib.Value{hlib.DefaultValue("cxxflags", cxxflags)}
	}
	if len(defines) > 0 {
		wtgt.Defines = []hlib.Value{hlib.DefaultValue("defines", defines)}
	}
	if len(includes) > 0 {
		wtgt.Includes = []hlib.Value{hlib.DefaultValue("includes", includes)}
	}
	return err
}

func parseLdFlags(p *Parser) error {
	var err error
	tokens := p.tokens
	wtgt := &p.req.Wscript.Build.Targets[0]
	ldflags := []string{}
	for _, flag := range tokens[1:] {
		flag = strings.Trim(flag, " \t")
		if len(flag) == 0 {
			continue
		}
		if strings.HasPrefix(flag, "-L") {
			// should be handled by hwaf.uses
			continue
		}
		if strings.HasPrefix(flag, "-l") {
			// should be handled by hwaf uses too
			continue
		}
		ldflags = append(ldflags, flag)
	}
	if len(ldflags) > 0 {
		wtgt.LinkFlags = []hlib.Value{hlib.DefaultValue("linkflags", ldflags)}
	}
	return err
}

func parsePreload(p *Parser) error {
	var err error
	return err
}

func parsePedantic(p *Parser) error {
	var err error
	return err
}

func parseReflex(p *Parser) error {
	var err error
	tokens := p.tokens
	wtgt := &p.req.Wscript.Build.Targets[0]
	if len(tokens) > 1 && tokens[1] == "1" {
		pkgname := p.req.Wscript.Package.Name
		delete(wtgt.KwArgs, "rootcint_linkdef")
		wtgt.KwArgs["selection_file"] = []hlib.Value{
			hlib.DefaultValue(
				"selection_file",
				[]string{
					fmt.Sprintf("%s/selection.xml", pkgname),
				},
			),
		}

		features := make([]string, 0, len(wtgt.Features))
		for _, x := range wtgt.Features {
			if x != "atlas_dictionary" {
				features = append(features, x)
			}
		}
		wtgt.Features = features
	}
	return err
}

func parseDep(p *Parser) error {
	var err error
	tokens := p.tokens
	wtgt := &p.req.Wscript.Build.Targets[0]
	uses := []string{}
	for _, tok := range tokens[1:] {
		tok = strings.Trim(tok, " \t")
		if len(tok) == 0 {
			continue
		}
		uses = append(uses, tok)
	}
	if len(uses) > 0 {
		wtgt.Use[0].Set[0].Value = append(
			wtgt.Use[0].Set[0].Value,
			uses...,
		)
	}
	return err
}

func parseTryDep(p *Parser) error {
	var err error
	return err
}

func parseClean(p *Parser) error {
	var err error
	return err
}

func parseNoOpt(p *Parser) error {
	var err error
	tokens := p.tokens
	wtgt := &p.req.Wscript.Build.Targets[0]
	if len(tokens) > 1 && tokens[1] == "1" {
		wtgt.CxxFlags[0].Set[0].Value = append(
			wtgt.CxxFlags[0].Set[0].Value,
			"-g",
		)
	}
	return err
}

func parseNoCC(p *Parser) error {
	var err error
	return err
}
func parseInclude(p *Parser) error {
	var err error
	return err
}

func parseIfNeq(p *Parser) error {
	var err error
	return err
}

func parseEndIf(p *Parser) error {
	var err error
	return err
}

// EOF
