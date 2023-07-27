package util

import (
	"io/ioutil"
	"malang/ast"
	"malang/lexer"
	"malang/parser"
)

// 加载标准库
func LoadStd() string {
	// todo: 改为循环读取std目录
	buf, err := ioutil.ReadFile("./std/std.mal")
	if err != nil {
		panic(err)
	}
	return string(buf)
}

// 加载用户定义的文件(返回加载后的字符串)
func LoadMalFile(filePath string) *ast.Program {
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	l := lexer.New(string(buf))
	p := parser.New(l)
	return p.ParseProgram()
}
