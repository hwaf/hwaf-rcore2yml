package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/hwaf/hwaf/hlib"
)

func path_exists(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// str_split slices s into all (non-empty) substrings separated by sep
func str_split(s, sep string) []string {
	strs := strings.Split(s, sep)
	out := make([]string, 0, len(strs))
	for _, str := range strs {
		str = strings.Trim(str, " \t")
		if len(str) == 0 {
			continue
		}
		out = append(out, str)
	}
	return out
}

// str_is_in_slice returns true if str is in the given slice of strings
func str_is_in_slice(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// re_is_in_slice_suffix returns true if an element in the given slice of strings is a prefix of value.
func re_is_in_slice_suffix(slice []string, macro, pattern string) bool {
	for _, s := range slice {
		pat := regexp.MustCompile(s + pattern)
		if pat.MatchString(macro) {
			return true
		}
	}
	return false
}

func hlib_value_from(value map[string]string) hlib.Value {
	hvalue := hlib.Value{}
	if _, ok := value["default"]; ok {
		vals := strings.Split(value["default"], " ")
		def_value := make([]string, 0, len(vals))
		for _, vv := range vals {
			vv = strings.Trim(vv, " \t")
			if len(vv) > 0 {
				def_value = append(def_value, vv)
			}
		}
		hvalue.Set = append(hvalue.Set,
			hlib.KeyValue{
				Tag:   "default",
				Value: def_value,
			},
		)
	}
	for k, v := range value {
		if k == "default" {
			continue
		}
		kv := hlib.KeyValue{Tag: k}
		vals := strings.Split(v, " ")
		for _, vv := range vals {
			vv = strings.Trim(vv, " \t")
			if len(vv) > 0 {
				kv.Value = append(kv.Value, vv)
			}
		}
		hvalue.Set = append(hvalue.Set, kv)
	}

	return hvalue
}

func hlib_value_from_slice(name string, values []string) hlib.Value {
	hvalue := hlib.Value{Name: name}
	if len(values) > 0 {
		dft := values[0]
		vals := strings.Split(dft, " ")
		kv := hlib.KeyValue{Tag: "default",
			Value: make([]string, 0, len(vals)),
		}
		for _, vv := range vals {
			vv = strings.Trim(vv, " \t")
			if len(vv) > 0 {
				kv.Value = append(kv.Value, vv)
			}
		}
		hvalue.Set = append(hvalue.Set, kv)
	}
	if len(values) > 1 {
		toks := values[1:]
		for i := 0; i+1 < len(toks); i += 2 {
			k := toks[i]
			v := toks[i+1]

			kv := hlib.KeyValue{Tag: k}
			vals := strings.Split(v, " ")
			for _, vv := range vals {
				vv = strings.Trim(vv, " \t")
				if len(vv) > 0 {
					kv.Value = append(kv.Value, vv)
				}
			}
			hvalue.Set = append(hvalue.Set, kv)
		}
	}
	return hvalue
}

func w_py_strlist(str []string) string {
	o := make([]string, 0, len(str))
	for _, v := range str {
		vv, err := strconv.Unquote(v)
		if err != nil {
			vv = v
		}
		if strings.HasPrefix(vv, `"`) && strings.HasSuffix(vv, `"`) {
			if len(vv) > 1 {
				vv = vv[1 : len(vv)-1]
			}
		}
		o = append(o, fmt.Sprintf("%q", vv))
	}
	return strings.Join(o, ", ")
}

// EOF
