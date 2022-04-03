package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	package_     = "package "
	_test        = "_test.go"
	func_        = "func "
	len_func_    = 5
	len_package_ = 8
)

var (
	golang_files      []string
	golang_test_files []string
)

func read_folders(folder_name string) {
	files, err := ioutil.ReadDir(folder_name)
	err_panic(err)
	for _, f := range files {
		dir := folder_name + "/" + f.Name()
		switch {
		case f.IsDir():
			read_folders(dir)
		case filepath.Ext(f.Name()) == ".go":
			if strings.Contains(f.Name(), _test) {
				golang_test_files = append(golang_test_files, dir)
			} else {
				golang_files = append(golang_files, dir)
			}
		}
	}
}

func file_lines(file_name string) []string {
	content, err := ioutil.ReadFile(file_name)
	err_panic(err)
	return strings.Split(string(content), "\n")
}

func err_panic(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func create_func_test_line(func_name string) string {
	return "func Test_" + func_name + "(t *testing.T) {"
}

func return_package_name(path string) string {
	for _, a := range file_lines(path) {
		if len(a) >= len_package_ && a[:len_package_] == package_ {
			return a[len_package_:]
		}
	}
	return ""
}

func write_test_tools(path string) {
	f, _ := os.Create(path + "/test_tools.txt")
	f.WriteString(
		`//this file created automatically
package package_name

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"text/tabwriter"
)

func TEST[t any](should_equal bool, actual, expected t) {
	if !reflect.DeepEqual(actual, expected) == should_equal {
		p := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
		fmt.Fprintln(p, "\033[35m", "should_equal\t:", should_equal) //purple
		fmt.Fprintln(p, "\033[34m", "actual\t:", actual)             //blue
		fmt.Fprintln(p, "\033[33m", "expected\t:", expected)         //yellow
		p.Flush()
		fmt.Println("\033[31m") //red
		log.Panic()
	}
}`)
}

func main() {
	read_folders(".")
	write_test_tools(".")
	package_name := return_package_name(golang_files[0])

	for _, a := range golang_files {
		all_func_name := return_all_func_name(a)
		is_there_func_in_the_file := len(all_func_name) > 0
		test_file_for_the_file := a[:len(a)-3] + _test
		if is_there_func_in_the_file {
			if IS_IN(test_file_for_the_file, golang_test_files) {
				f, _ := os.OpenFile(test_file_for_the_file, os.O_APPEND|os.O_WRONLY, 0644)
				for _, b := range all_func_name {
					if !IS_IN(create_func_test_line(b), file_lines(test_file_for_the_file)) {
						write_test_func(f, b)
					}
				}
				f.Close()
			} else {
				f, _ := os.Create(test_file_for_the_file)
				f.WriteString("//this file created automatically\n")
				f.WriteString("package " + package_name + "\n")

				if is_there_func_in_the_file {
					f.WriteString("\n")
					f.WriteString("import \"testing\"\n")
				}
				for _, b := range all_func_name {
					write_test_func(f, b)
				}
				f.Close()
			}
		}
	}
}

func write_test_func(f *os.File, b string) {
	f.WriteString("\n")
	f.WriteString(create_func_test_line(b) + "\n")
	f.WriteString("\t//a:=" + b + "()\n")
	f.WriteString("\t//TEST(true,a,e)\n")
	f.WriteString("}\n")
}

func IS_IN[t any](element t, elements []t) bool {
	for _, a := range elements {
		if reflect.DeepEqual(a, element) {
			return true
		}
	}
	return false
}

func return_all_func_name(path string) []string {
	var all_func_name []string
	for _, b := range file_lines(path) {
		var func_name string
		if len(b) > len_func_ && b[:len_func_] == func_ {
			for _, c := range b[len_func_:] {
				if c == '(' || c == '[' {
					all_func_name = append(all_func_name, func_name)
					break
				}
				func_name += string(c)
			}
		}
	}
	return all_func_name
}
