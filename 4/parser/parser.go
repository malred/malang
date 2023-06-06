// parser/parser.go
package parser

import (
	"fmt"
	"malang/ast"
	"malang/lexer"
	"malang/token"
	"strconv"
)

const (
	_ int = iota // 0
	LOWEST
	EQUALS     // ==
	LESSGEATER // > or <
	SUM        // +
	PRODUCT    // *
	PREFIX     // -X or !X
	CALL       // myFunction(X)
)

// 优先级map
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGEATER,
	token.GT:       LESSGEATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l *lexer.Lexer

	errors    []string
	curToken  token.Token
	peekToken token.Token

	// 解析函数
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

// 查询下一个词法单元的优先级
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

// 查询当前词法单元的优先级
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// 解析函数-标识符-前缀
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// 解析函数-整数字面量-前缀
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

// 解析函数-前缀表达式-前缀
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	// 解析表达式,调用解析函数
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// 解析函数-中缀表达式-中缀
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	// 进入下一个词法单元然后填充Right
	p.nextToken()
	// 解析下一个表达式,并传入优先级
	expression.Right = p.parseExpression(precedence)

	return expression
}

// 解析函数-布尔字面量-前缀
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// 解析函数-分组表达式(括号)-前缀
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// 解析块内语句
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	// !}
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// 解析函数-IF表达式-前缀
func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	// if(
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	// 解析if()里的表达式
	expression.Condition = p.parseExpression(LOWEST)

	// if()
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// if(){
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// 解析块内语句
	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

// 解析函数参数列表
func (p *Parser) parseFunctionParameter() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	// fn()
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	// fn(arg1,
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // arg2
		p.nextToken() // ,
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	// fn(arg1,arg2)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return identifiers
}

// 解析函数-函数表达式-前缀
func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{}

	// fn(
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameter()

	// fn (args){
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

// 解析调用参数
func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	// add()
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	// add(2,
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // 3
		p.nextToken() // ) or ,
		args = append(args, p.parseExpression(LOWEST))
	}

	// add(2,3)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

// 解析函数-调用表达式-中缀
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// 创建解析器
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	// 关联解析函数
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	// 读取两个词法单元,设置peekToken和curToken
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

// 记录error
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// 判断parser当前token是否和传入的token类型一致
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// 判断parser下一个token是否和传入的token类型一致
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// 是否是期望的token类型
func (p *Parser) expectPeek(t token.TokenType) bool {
	// 是期望类型就前移指针
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		// 记录error
		p.peekError(t)
		return false
	}
}

// 解析let语句
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	// 如果接下来不是标识符(如果是,指针前移)
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 如果接下来不是=
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// 解析return语句
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// 无法解析的语句
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// 解析表达式
func (p *Parser) parseExpression(preceduece int) ast.Expression {
	// 获取前缀处理函数
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		// 处理未知语句
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	// 调用处理函数,返回左侧表达式
	leftExp := prefix()

	// 循环直到遇到分号,并且下一个词法单元优先级高于当前优先级(高的先执行)
	for !p.peekTokenIs(token.SEMICOLON) && preceduece < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		// 如果没有中缀解析函数->不是中缀表达式或没有解析
		if infix == nil {
			// 前面已经处理了前缀表达式,直接返回
			return leftExp
		}

		p.nextToken()

		// 解析中缀表达式
		leftExp = infix(leftExp)
	}
	return leftExp
}

// 处理表达式语句(前缀/中缀)
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	// 分号是可选的,有没有都没关系
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// 解析语句
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	// 遇到LET开头就解析let语句
	case token.LET:
		return p.parseLetStatement()
	// 遇到return开头就解析return语句
	case token.RETURN:
		return p.parseReturnStatement()
	// 解析表达式
	default:
		return p.parseExpressionStatement()
	}
}

// 解析源码,构建AST
func (p *Parser) ParseProgram() *ast.Program {
	// 构造根节点
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		// 解析语句
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}
