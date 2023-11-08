package evaluator

import (
	"monkey/ast"
	"monkey/object"
)

var (
	TRUE = &object.Boolean{Value:true}
	FALSE = &object.Boolean{Value:false}
	NULL = &object.NULL{}
)

func Eval(node ast.Node) object.Object{
	switch node := node.(type) {
	case *ast.Program:
		return loopEval(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.BooleanExpression:
		return nativeBooleanObject(node.Value) 
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	}
	return nil
}

func nativeBooleanObject(input bool) *object.Boolean{
	if input {
		return TRUE
	}
	return FALSE
}

func loopEval(stmts []ast.Statement) object.Object{
	var result object.Object
	for _, stmt := range stmts {
		result = Eval(stmt)
	}
	return result
}