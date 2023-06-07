// lexer/lexer.go
package lexer

import (
	"malang/token"
)

type Lexer struct {
	input        string
	position     int  // 输入的字符串中的当前位置(指向当前字符)
	readPosition int  // 输入的字符串中的当前读取位置(指向当前字符串之后的一个字符(ch))
	ch           byte // 当前正在查看的字符
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	// 初始化 l.ch,l.position,l.readPosition
	l.readChar()
	return l
}

// 读取下一个字符
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // NUL的ASSII码(0)
	} else {
		// 读取
		l.ch = l.input[l.readPosition]
	}
	// 前移
	l.position = l.readPosition
	l.readPosition += 1
}

// 创建词法单元的方法
func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(ch),
	}
}

// 判断读取到的字符是不是字母
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// 读取字母(标识符/关键字)
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		// 如果接下来还有字母,就一直移动指针到不是字母
		l.readChar()
	}
	return l.input[position:l.position]
}

// 跳过空格
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// 判断是否是数字
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// 读取数字
func (l *Lexer) readNumber() string {
	// 记录起始位置
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// 向前查看一个字符,但是不移动指针
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

// 读取注释内容
func (l *Lexer) readComment() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '\n' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

// 根据当前的ch创建词法单元
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	// 跳过空格
	l.skipWhitespace()

	switch l.ch {
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case '&':
		if l.peekChar() == '&' {
			// 记录当前ch (&)
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.AND, Literal: literal}
		} else {
			// 未知符合 &
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '|':
		if l.peekChar() == '|' {
			// 记录当前ch (|)
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.OR, Literal: literal}
		} else {
			// 未知符号 |
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '=':
		if l.peekChar() == '=' {
			// 记录当前ch (=)
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '!':
		if l.peekChar() == '=' {
			// 记录当前ch (!)
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '/':
		if l.peekChar() == '/' {
			l.readChar()               // 跳过 /
			literal := l.readComment() // 读取注释内容
			tok = token.Token{Type: token.COMMENT, Literal: literal}
		} else {
			tok = newToken(token.SLASH, l.ch)
		}
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ':':
		tok = newToken(token.COLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			// 因为readIdentifier会调用readChar,所以提前return,不需要后面再readChar
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}
