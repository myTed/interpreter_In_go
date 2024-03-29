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

func applyFunction(ce *ast.CallExpression, env *object.Environment) object.Object{
	
	
	function := Eval(ce.Function, env)

	switch function.(type) {
	case *object.Function:
		f := function.(*object.Function)
		newEnv := object.NewEnvironment(f.Env)
		for idx, arg := range ce.Arguments {
			val := Eval(arg, env)
			if isError(val) {
				return val
			}
			newEnv.Set(f.Parameters[idx].String(), val)
		}
		evaluated := Eval(f.Body, newEnv)
		if returnValue, ok := evaluated.(*object.ReturnValue); ok {
			return returnValue.Value
		}
		return evaluated
	case *object.Builtin:
		f := function.(*object.Builtin)
		var args = []object.Object{}
		for _, arg := range ce.Arguments {
			val := Eval(arg, env)
			if isError(val) {
				return val
			}
			args = append(args , val)
		}
		return f.Fn(args...)
	default:
		return newError("not a function: %s", function.Type())
	}
}

func Eval(node ast.Node, env *object.Environment) object.Object{
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.FunctionLiteral:
		return &object.Function{Parameters: node.Parameters,
		Body: node.Body, Env: env}
	case *ast.CallExpression:
		return applyFunction(node, env)		
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BooleanExpression:
		return nativeBooleanObject(node.Value) 
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		right := Eval(node.Right, env)
		if isError(left) {
			return left
		}
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node.Statements, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		return evalReturnExpression(node.ReturnValue, env)
	}
	return nil
}


func evalIdentifier(ident *ast.Identifier, env *object.Environment) object.Object{
	val, ok := env.Get(ident.Value)
	if ok {
		return val
	}
	if builtin, ok := Builtins[ident.Value]; ok {
		return builtin
	}
	return newError("identifier not found: %s", ident.Value)
}

func evalReturnExpression(node ast.Node, env *object.Environment) object.Object{
	returnValue := Eval(node, env)
	if isError(returnValue) {
		return returnValue
	}
	return &object.ReturnValue{Value: returnValue}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}
	if (isTruthy(condition)) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
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
	case left.Type() == object.STRING_OBJ &&
			right.Type() == object.STRING_OBJ:
			 return evalStringInfixExpression(operator, left, right)
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

func evalStringInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	leftString := left.(*object.String).Value
	rightString := right.(*object.String).Value

	switch operator {
	case "+":
		return &object.String{Value: leftString + rightString}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
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

func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object{
	var result object.Object
	for _, stmt := range stmts {
		result = Eval(stmt, env)
		
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalBlockStatement(stmts []ast.Statement, env *object.Environment) object.Object{
	var result object.Object
	for _, stmt := range stmts {
		result = Eval(stmt, env)
		
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}