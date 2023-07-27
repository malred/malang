// main.go
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"malang/repl"
	"os"
	"os/user"
)

type Cmd struct {
	helpFlag    bool
	versionFlag bool
	replFlag    bool // 控制台程序
	cpOption    string
	malFile     string // 待编译的文件
	args        []string
}

func printUsage() {
	fmt.Printf("Usage: %s [-options] [args...]\n", os.Args[0])
}
func parseCmd() *Cmd {
	cmd := &Cmd{}
	flag.Usage = printUsage
	flag.BoolVar(&cmd.helpFlag, "help", false, "print help message")
	flag.BoolVar(&cmd.helpFlag, "?", false, "print help message")
	flag.BoolVar(&cmd.replFlag, "repl", false, "repl")
	flag.BoolVar(&cmd.versionFlag, "v", false, "print version and exit")
	flag.StringVar(&cmd.cpOption, "filepath", "", "filepath")
	flag.StringVar(&cmd.cpOption, "f", "", "filepath")
	flag.StringVar(&cmd.cpOption, "fe", "", "macro filepath")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		// ./malang.exe -f 1.mal
		// 此时cpOption == 1.mal
		cmd.args = args
	}
	return cmd
}
func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is the Malang programming language!\n", user.Username)
	fmt.Printf("Feel free to type in commands\n")
	cmd := parseCmd()
	if cmd.versionFlag {
		fmt.Println("version: 0.0.1 by malred 2023.6.6")
	} else if cmd.helpFlag {
		printUsage()
	} else if cmd.replFlag {
		repl.Start(os.Stdin, os.Stdout)
	} else {
		// 读取-f指定的文件
		// fmt.Println("args: ", cmd.args)
		fmt.Println("reading: ", cmd.cpOption)
		// if cmd.cpOption == "-m" {
		// 	// 宏环境
		// 	fmt.Println("macro")
		// }
		buf, err := ioutil.ReadFile(cmd.cpOption)
		if err != nil {
			panic(err)
		}
		input := string(buf)
		repl.ReadAndEval(input)
	}
}
