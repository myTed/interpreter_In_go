package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
)

var (
	TRUE = &object.Boolean{Value:true}
	FALSE = &object.Boolean{Value:false}
	NULL = &object.NULL{}
)

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool{
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func Eval(node ast.Node) object.Object{
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.BooleanExpression:
		return nativeBooleanObject(node.Value) 
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		right := Eval(node.Right)
		if isError(left) {
			return left
		}
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node.Statements)
	case *ast.IfExpression:
		return evalIfExpression(node)
	case *ast.ReturnStatement:
		return evalReturnExpression(node.ReturnValue)
	}
	return nil
}


func evalReturnExpression(node ast.Node) object.Object{
	returnValue := Eval(node)
	if isError(returnValue) {
		return returnValue
	}
	return &object.ReturnValue{Value: returnValue}
}

func evalIfExpression(ie *ast.IfExpression) object.Object {
	condition := Eval(ie.Condition)
	if isError(condition) {
		return condition
	}
	if (isTruthy(condition)) {
		return Eval(ie.Consequence)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative)
	} else {
		return NULL
	}
}

func isTruthy(condition object.Object) bool {
	switch condition{
	case TRUE:
		return true
	case FALSE:
		return false
	case NULL:
		return false
	default:
		return true
	}
}

func evalInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ &&
			 right.Type() == object.INTEGER_OBJ:
			 return evalIntegerInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBooleanObject(left == right)
	case operator == "!=":
		return nativeBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
		left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(),
		operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value
	switch operator {
	case "+":
		return &object.Integer{Value: leftValue + rightValue}
	case "-":
		return &object.Integer{Value: leftValue - rightValue}
	case "*":
		return &object.Integer{Value: leftValue * rightValue}
	case "/":
		return &object.Integer{Value: leftValue / rightValue}
	case ">":
		return nativeBooleanObject(leftValue > rightValue)
	case "<":
		return nativeBooleanObject(leftValue < rightValue)
	case "==":
		return nativeBooleanObject(leftValue == rightValue)
	case "!=":
		return nativeBooleanObject(leftValue != rightValue)
	default:
		return newError("unknown operator: %s %s %s", left.Type(),
	operator, right.Type())
	}
}


func evalPrefixExpression(operator string, right object.Object) object.Object{
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalBangOperatorExpression(right object.Object) object.Object{
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func nativeBooleanObject(input bool) *object.Boolean{
	if input {
		return TRUE
	}
	return FALSE
}

func evalProgram(stmts []ast.Statement) object.Object{
	var result object.Object
	for _, stmt := range stmts {
		result = Eval(stmt)
		
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalBlockStatement(stmts []ast.Statement) object.Object{
	var result object.Object
	for _, stmt := range stmts {
		result = Eval(stmt)
		
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}