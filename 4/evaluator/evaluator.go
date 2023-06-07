// evaluate/evaluate.go
package evaluator

import (
	"fmt"
	"malang/ast"
	"malang/object"
	"malang/util"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

// 求布尔型的值
func nativeBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

// -操作符求值(前缀)
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

// !操作符求值
func evalBangOperatorExpression(right object.Object) object.Object {
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

// 解析前缀 ! -
func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s"+operator, right.Type())
	}
}

// 解析两侧是数字的正则表达式(eg. 2*2)
func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// 解析字符串中缀操作
func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

// 解析中缀表达式
func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	// 这里可以直接对比是因为布尔型都是复用true和false两个对象的指针(地址),可以直接比对地址来看是否相等
	case operator == "==":
		return nativeBooleanObject(left == right)
	case operator == "!=":
		return nativeBooleanObject(left != right)
	// todo: && 和 ||
	case operator == "&&":
		return nativeBooleanObject(left.(*object.Boolean).Value && right.(*object.Boolean).Value)
	case operator == "||":
		return nativeBooleanObject(left.(*object.Boolean).Value || right.(*object.Boolean).Value)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// if判真策略(真值[true,!false,!null]为成立)
func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

// 解析If表达式
func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

// 解析program
func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

// 对块语句进行求值
func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

// 对标识符求值
func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: " + node.Value)
}

// 对表达式求值
func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

// 解包函数返回值(如果不接包,会冒泡,然后停止后续的语句求值)
func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

// 拓展环境
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

// 函数体求值
func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

// 数组索引求值
func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObj := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObj.Elements) - 1)
	if idx < 0 || idx > max {
		return NULL
	}
	return arrayObj.Elements[idx]
}

// 索引表达式求值
func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	// 语句 -> 继续遍历
	// 根节点
	case *ast.Program:
		return evalProgram(node, env)
	// use导入语句
	case *ast.UseExpression:
		// 解析
		return Eval(util.LoadMalFile(node.FileName), env)
	// 块语句
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	// Let语句
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		// 关联标识符和值
		env.Set(node.Name.Value, val)
	// 标识符
	case *ast.Identifier:
		return evalIdentifier(node, env)
	// 函数语句
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Body: body, Env: env}
	// 调用函数
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)
	// Return语句
	case *ast.ReturnStatement:
		// 对返回值进行求值
		val := Eval(node.ReturnValue, env)
		// 判断该中断是不是Error引发的
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	// IF语句
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	// 索引表达式
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	// 表达式语句
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	// 表达式 -> 求值
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	// 布尔型
	case *ast.Boolean:
		return nativeBooleanObject(node.Value)
	// 字符串
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	// 数组
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	// 前缀表达式
	case *ast.PrefixExpression:
		// 如果一直是!或-,就会一直Eval,直到遇到其他类型,多次执行!或-
		right := Eval(node.Right, env)
		// 判断该中断是不是Error引发的
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	// 中缀表达式
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		// 判断该中断是不是Error引发的
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		// 判断该中断是不是Error引发的
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	}
	return nil
}

func newError(format string, args ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, args...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
