package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

type (
	makePrefixFn func() ast.Expression
	makeInfixFn  func(ast.Expression) ast.Expression
)

const (
	_ int = iota
	LOWEST
	EQUALS
	LESS_GREATER
	SUM
	PRODUCT
	PREFIX
	CALL
)

var precedences = map[token.TokenType]int{
	token.EQUAL:     EQUALS,
	token.NOT_EQUAL: EQUALS,
	token.LT:        LESS_GREATER,
	token.GT:        LESS_GREATER,
	token.PLUS:      SUM,
	token.MINUS:     SUM,
	token.SLASH:     PRODUCT,
	token.ASTERISK:  PRODUCT,
	token.LPAREN: 	 CALL,
}

type Parser struct {
	lexer         *lexer.Lexer
	curToken      token.Token
	peekToken     token.Token
	errors        []string
	makePrefixFns map[token.TokenType]makePrefixFn
	makeInfixFns  map[token.TokenType]makeInfixFn
}

//Parser method
func MakeNewParser(lexer *lexer.Lexer) *Parser {
	newParser := &Parser{lexer: lexer}
	newParser.nextToken()
	newParser.nextToken()
	newParser.makePrefixFns = make(map[token.TokenType]makePrefixFn)
	newParser.makePrefixFns[token.INT] = newParser.makeIntegerLiteral
	newParser.makePrefixFns[token.IDENT] = newParser.makeIdentifier
	newParser.makePrefixFns[token.MINUS] = newParser.makePrefix
	newParser.makePrefixFns[token.BANG] = newParser.makePrefix
	newParser.makePrefixFns[token.TRUE] = newParser.makeBoolean
	newParser.makePrefixFns[token.FALSE] = newParser.makeBoolean
	newParser.makePrefixFns[token.LPAREN] = newParser.makeGroupExpression
	newParser.makePrefixFns[token.IF] = newParser.makeIfExpression
	newParser.makePrefixFns[token.FUNCTION] = newParser.makeFuncExpression
	newParser.makeInfixFns = make(map[token.TokenType]makeInfixFn)
	newParser.makeInfixFns[token.PLUS] = newParser.makeInfix
	newParser.makeInfixFns[token.MINUS] = newParser.makeInfix
	newParser.makeInfixFns[token.SLASH] = newParser.makeInfix
	newParser.makeInfixFns[token.ASTERISK] = newParser.makeInfix
	newParser.makeInfixFns[token.GT] = newParser.makeInfix
	newParser.makeInfixFns[token.LT] = newParser.makeInfix
	newParser.makeInfixFns[token.EQUAL] = newParser.makeInfix
	newParser.makeInfixFns[token.NOT_EQUAL] = newParser.makeInfix
	newParser.makeInfixFns[token.LPAREN] = newParser.makeCallExpression
	return newParser
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) checkNextToken(t token.TokenType) bool {
	if p.peekToken.Type != t {
		msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Literal)
		p.errors = append(p.errors, msg)
		return false
	}
	p.nextToken()
	return true
}

func (p *Parser) makeLetStatement() *ast.LetStatement {
	statement := &ast.LetStatement{Token: p.curToken}

	if !p.checkNextToken(token.IDENT) {
		return nil
	}
	statement.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.checkNextToken(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	
	statement.Value = p.makeExpression(LOWEST)

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}
	return statement
}

func (p *Parser) makeReturnStatement() *ast.ReturnStatement {
	statement := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	statement.ReturnValue = p.makeExpression(LOWEST)
	if p.curToken.Type == token.SEMICOLON {
		p.nextToken()
	}
	return statement
}

func (p *Parser) makePrefix() ast.Expression {
	pe := &ast.PrefixExpression{Token: p.curToken, Operator: p.curToken.Literal}
	p.nextToken()
	pe.Right = p.makeExpression(PREFIX)
	if p.curToken.Type == token.SEMICOLON {
		p.nextToken()
	}
	return pe
}

func (p *Parser) makeInfix(left ast.Expression) ast.Expression {
	ie := &ast.InfixExpression{
		Token: p.curToken,
		Operator: p.curToken.Literal,
		Left: left,
	}
	curPrecedence := p.curPrecedence()
	p.nextToken()
	ie.Right = p.makeExpression(curPrecedence)

	return ie
}


func (p *Parser) makeCallExpression(function ast.Expression) ast.Expression {
	call := &ast.CallExpression{Token: p.curToken, Function: function}
	p.nextToken()
	for p.curToken.Type != token.RPAREN {
		if p.curToken.Type == token.COMMA {
			p.nextToken()
		}
		call.Arguments = append(call.Arguments, p.makeExpression(LOWEST))
		p.nextToken()
	}
	return call
}

func (p *Parser) makeFuncExpression() ast.Expression {
	function := &ast.FunctionLiteral{Token: p.curToken}

	if p.peekToken.Type != token.LPAREN {
		return nil
	}
	p.nextToken()
	p.nextToken()
	if p.curToken.Type != token.RPAREN {
		ident := &ast.Identifier{Token:p.curToken, Value: p.curToken.Literal}
		function.Parameters = append(function.Parameters, ident)
		p.nextToken()
		for p.curToken.Type == token.COMMA {
			p.nextToken()
			ident = &ast.Identifier{Token:p.curToken, Value: p.curToken.Literal}
			function.Parameters = append(function.Parameters, ident)
			p.nextToken() 
		}
	}
	if p.curToken.Type != token.RPAREN {
		return nil
	}
	if p.peekToken.Type != token.LBRACE {
		return nil
	}
	p.nextToken()
	function.Body = p.parseBlockStatement() 
	return function
}

func (p *Parser) makeIfExpression() ast.Expression {
	ie := &ast.IfExpression{Token: p.curToken}
	if p.peekToken.Type != token.LPAREN {
		return nil
	}
	p.nextToken()
	p.nextToken()
	ie.Condition = p.makeExpression(LOWEST)
	if p.peekToken.Type != token.RPAREN {
		return nil
	}
	p.nextToken()
	if p.peekToken.Type != token.LBRACE {
		return nil
	}
	p.nextToken()
	ie.Consequence = p.parseBlockStatement()
	if p.peekToken.Type == token.ELSE {
		p.nextToken()
		if p.peekToken.Type != token.LBRACE {
			return nil
		}
		p.nextToken()
		ie.Alternative = p.parseBlockStatement()
	}
	return ie
}

func (p *Parser) makeIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}


func (p *Parser) makeBoolean() ast.Expression {
	return &ast.BooleanExpression{Token: p.curToken, Value: p.curToken.Type == token.TRUE}
}

func (p *Parser) makeGroupExpression() ast.Expression {
	
	p.nextToken()
	exp := p.makeExpression(LOWEST)
	if p.peekToken.Type != token.RPAREN {
		return nil
	}
	p.nextToken()
	return exp
}


func (p *Parser) makeIntegerLiteral() ast.Expression {
	il := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	il.Value = value
	return il
}

func (p *Parser) makeExpression(precedence int) ast.Expression {
	makePrefixFn := p.makePrefixFns[p.curToken.Type]
	if makePrefixFn == nil {
		return nil
	}
	newExpression := makePrefixFn()

	for precedence < p.peekPrecedence() {
		makeInfixFn:= p.makeInfixFns[p.peekToken.Type]
		if makeInfixFn == nil {
			return newExpression
		}
		p.nextToken()
		newExpression = makeInfixFn(newExpression)
	}
	
	return newExpression
}



func (p *Parser) makeExpressionStatement() *ast.ExpressionStatement {
	statement := &ast.ExpressionStatement{Token: p.curToken}
	statement.Expression = p.makeExpression(LOWEST)

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}
	return statement
}

func (p *Parser) makeStatement() ast.Statement {
	var statement ast.Statement

	switch p.curToken.Type {
	case token.LET:
		statement = p.makeLetStatement()
	case token.RETURN:
		statement = p.makeReturnStatement()
	default:
		statement = p.makeExpressionStatement()
	}
	return statement
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		statement := p.makeStatement()
		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	blockStatement := &ast.BlockStatement{Token: p.curToken}
	blockStatement.Statements = []ast.Statement{}

	p.nextToken()
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		statement := p.makeStatement()
		if statement != nil {
			blockStatement.Statements = append(blockStatement.Statements, statement)
		}
		p.nextToken()
	}
	return blockStatement
}
