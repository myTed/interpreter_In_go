package ast

import "monkey/token"

type Node interface {
	TokenLiteral() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Identifier struct {
	Token token.Token
	value string
}

type LetStatement struct {
	Token token.Token
	name  Identifier
	value Expression
}

func (l *LetStatement) statementNode() {}
func (l *LetStatement) TokenLiteral() string {
	return l.Token.Literal
}

func (l *LetStatement) expressionNode() {}

type Program struct {
	Statements []Statement
}
