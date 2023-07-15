package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/mjwhitta/errors"
	"github.com/mjwhitta/pathname"
)

var fixlen = regexp.MustCompile(`(len\(.+?\))`)
var spaces = regexp.MustCompile(`\s+`)

func format(str string) string {
	var tmp []string

	// If no "_", no fix needed
	if !strings.Contains(str, "_") {
		return str
	}

	// Split on "_"
	tmp = strings.Split(strings.ToLower(str), "_")

	// Capitalize every part
	for i := range tmp {
		if tmp[i] == "" {
			continue
		}

		tmp[i] = strings.ToUpper(tmp[i][0:1]) + tmp[i][1:]
	}

	// Join together for camelcase
	str = strings.Join(tmp, "")

	// Fix some special cases
	str = strings.ReplaceAll(str, "Crlf", "CRLF")
	str = strings.ReplaceAll(str, "Http", "HTTP")

	return str
}

func genFile(
	pkg string,
	cache map[string]string,
	lines [][]string,
) error {
	var e error
	var f *os.File
	var fn string = "generated.go"
	var out []string
	var sorted []string

	if f, e = os.Create(fn); e != nil {
		return errors.Newf("failed to create %s: %w", fn, e)
	}
	defer f.Close()

	f.WriteString(
		"// Code generated by tools/defines.go; DO NOT EDIT.\n",
	)
	f.WriteString("package " + pkg + "\n\n")
	f.WriteString("const (\n")

	for k := range cache {
		sorted = append(sorted, k)
	}

	sort.Slice(
		sorted,
		func(i int, j int) bool {
			return len(sorted[i]) > len(sorted[j])
		},
	)

	for _, l := range lines {
		for _, k := range sorted {
			l[1] = strings.ReplaceAll(l[1], k, cache[k])
		}

		if fixlen.MatchString(l[1]) {
			l[1] = fixlen.ReplaceAllString(l[1], "uintptr($1)")
		}

		if strings.HasPrefix(l[1], "\"") {
			out = append(
				out,
				fmt.Sprintf("\t%s string = %s\n", l[0], l[1]),
			)
		} else if strings.HasPrefix(l[1], "L\"") {
			out = append(
				out,
				fmt.Sprintf(
					"\t%s string = %s\n",
					l[0],
					strings.Replace(l[1], "L", "", 1),
				),
			)
		} else if strings.HasPrefix(l[1], "TEXT(") {
			l[1] = strings.Replace(l[1], "TEXT(", "", 1)
			l[1] = strings.Replace(l[1], ")", "", 1)

			out = append(
				out,
				fmt.Sprintf("\t%s string = %s\n", l[0], l[1]),
			)
		} else {
			out = append(
				out,
				fmt.Sprintf("\t%s uintptr = %s\n", l[0], l[1]),
			)
		}
	}

	sort.Slice(
		out,
		func(i int, j int) bool {
			return strings.ToLower(out[i]) < strings.ToLower(out[j])
		},
	)

	for _, line := range out {
		f.WriteString(line)
	}

	f.WriteString(")\n")

	return nil
}

func init() {
	flag.Parse()
}

func main() {
	var cache = map[string]string{
		"NULL":   "0",
		"sizeof": "len",
	}
	var lines [][]string

	if flag.NArg() == 0 {
		return
	}

	for i, arg := range flag.Args() {
		if i == 0 {
			continue
		}

		arg = "/usr/x86_64-w64-mingw32/include/" + arg

		if ok, e := pathname.DoesExist(arg); e != nil {
			fmt.Println(e.Error())
			os.Exit(1)
		} else if !ok {
			fmt.Printf("file %s not found\n", arg)
			os.Exit(1)
		}

		if e := processFile(&cache, &lines, arg); e != nil {
			panic(e)
		}
	}

	if e := genFile(flag.Arg(0), cache, lines); e != nil {
		panic(e)
	}
}

func processFile(
	cache *map[string]string,
	lines *[][]string,
	fn string,
) error {
	var b []byte
	var e error
	var f *os.File
	var fullLine string
	var tmp []string

	if f, e = os.Open(pathname.ExpandPath(fn)); e != nil {
		return errors.Newf("failed to open %s: %w", fn, e)
	}
	defer f.Close()

	if b, e = io.ReadAll(f); e != nil {
		return errors.Newf("failed to read %s: %w", fn, e)
	}

	for _, l := range strings.Split(string(b), "\n") {
		l = strings.ReplaceAll(l, "| INTERNET_FLAG_BGUPDATE", "")
		l = strings.TrimSpace(spaces.ReplaceAllString(l, " "))

		if strings.HasSuffix(l, "\\") {
			fullLine += l[:len(l)-1]
			continue
		} else {
			fullLine += l
		}

		l = fullLine
		fullLine = ""

		if !strings.HasPrefix(l, "#define") {
			continue
		}

		l = strings.Replace(l, "#define ", "", 1)
		tmp = strings.SplitN(l, " ", 2)

		if skip(tmp) {
			continue
		}

		if _, ok := (*cache)[tmp[0]]; !ok {
			(*cache)[tmp[0]] = format(tmp[0])
			*lines = append(*lines, []string{format(tmp[0]), tmp[1]})
		}
	}

	return nil
}

func skip(tmp []string) bool {
	if (len(tmp) != 2) ||
		strings.Contains(tmp[0], "(") ||
		strings.Contains(tmp[0], ")") ||
		strings.Contains(tmp[0], "BOOLAPI") ||
		strings.Contains(tmp[1], "_(") ||
		strings.Contains(tmp[1], "__MINGW_NAME") ||
		strings.Contains(tmp[1], "DWORD") ||
		strings.Contains(tmp[1], "EXTERN_C") ||
		strings.Contains(tmp[1], "STATUS_CALLBACK") ||
		strings.Contains(tmp[1], "~") ||
		(tmp[0]+"W" == tmp[1]) {
		return true
	}

	return false
}
