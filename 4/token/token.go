// token/token.go
package token

const (
	// 特殊类型
	ILLEGAL = "ILLEGAL" // 未知字符
	EOF     = "EOF"     // 文件结尾
	COMMENT = "//"      // 注释

	// 标识符+字面量
	IDENT  = "IDENT"  // add, foobar, x, y
	INT    = "INT"    // 1343456
	STRING = "STRING" // "hello"

	// 运算符
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	AND      = "&&"
	OR       = "||"

	LT = "<"
	GT = ">"

	EQ     = "=="
	NOT_EQ = "!="

	// 分隔符
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// 关键字
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	USE      = "USE"
	FOR      = "FOR"   // TODO
	RANGE    = "RANGE" // TODO
)

// 关键字map
var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"use":    USE,
	"for":    FOR,
	"range":  RANGE,
}

func LookupIdent(ident string) TokenType {
	// 从关键字map里找,找到了就说明是关键字
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	// 标识符
	return IDENT
}

// 词法单元类型
type TokenType string

// 词法单元
type Token struct {
	Type TokenType
	// 字面量
	Literal string
}
