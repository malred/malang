package evaluator

import (
	"malang/lexer"
	"malang/object"
	"malang/parser"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
	ts := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}
	for _, tt := range ts {
		eval := testEval(tt.input)
		testIntegerObject(t, eval, tt.expected)
	}
}
func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	pro := p.ParseProgram()
	env := object.NewEnvironment()
	return Eval(pro, env)
}
func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	res, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("obj is not int")
		return false
	}
	if res.Value != expected {
		t.Errorf("obj has wrong val`")
		return false
	}
	return true
}
func TestEvalBooleanExpression(t *testing.T) {
	ts := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"(1 < 2) == ture", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == ture", false},
		{"(1 > 2) == false", true},
	}
	for _, tt := range ts {
		eval := testEval(tt.input)
		testBooleanObject(t, eval, tt.expected)
	}
}
func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	res, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("obj is not boolean")
		return false
	}
	if res.Value != expected {
		t.Errorf("obj has wrong val")
		return false
	}
	return true
}
func TestBangOperator(t *testing.T) {
	ts := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}
	for _, tt := range ts {
		eval := testEval(tt.input)
		testBooleanObject(t, eval, tt.expected)
	}
}
func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("obj is not null. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}
func TestIfElseExpression(t *testing.T) {
	ts := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) {10}", nil},
		{"if (1){10}", 10},
		{"if (1<2){10}", 10},
		{"if (1>2){10}", nil},
		{"if (1>2){10} else {20}", 20},
		{"if (1<2){10} else {20}", 10},
	}
	for _, tt := range ts {
		eval := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, eval, int64(integer))
		} else {
			testNullObject(t, eval)
		}
	}
}
func TestReturnStatements(t *testing.T) {
	ts := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2*5; 9;", 10},
		{"9; return 2*5; 9;", 10},
		{
			`
			if(10>1){
				if(10>1){
					return 10;
				}
				return 1;
			}
		`, 10},
	}
	for _, tt := range ts {
		eval := testEval(tt.input)
		testIntegerObject(t, eval, tt.expected)
	}
}
func TestErrorHandling(t *testing.T) {
	ts := []struct {
		input           string
		expectedMessage string
	}{
		{"5 + true;", "type mismatch: INTEGER + BOOLEAN"},
		{"5 + true; 5;", "type mismatch: INTEGER + BOOLEAN"},
		{"-true", "unknown operator: -BOOLEAN"},
		{"true + false;", "unknown operator: BOOLEAN + BOOLEAN"},
		{"5; true + false; 5;", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) {true+false;}", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) {{false+false;} return 1;}", "unknown operator: BOOLEAN + BOOLEAN"},
		{"foobar;", "identifier not found: foobar"},
	}
	for _, tt := range ts {
		eval := testEval(tt.input)
		errobj, ok := eval.(*object.Error)
		if !ok {
			t.Errorf("no error obj returned. got=%T(%+v)", eval, eval)
			return
		}
		if errobj.Message != tt.expectedMessage {
			t.Errorf("wrong error msg: expected=%q, got=%q", tt.expectedMessage, errobj.Message)
			return
		}
	}
}
func TestLetStatement(t *testing.T) {
	ts := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a=5;let b = a; let c = a + b + 5;c;", 15},
	}
	for _, tt := range ts {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}
func TestFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2;};"
	eval := testEval(input)
	fn, ok := eval.(*object.Function)
	if !ok {
		t.Fatalf("object is not function")
		return
	}
	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters")
		return
	}
	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameters is not 'x'")
		return
	}
	expectedBody := "(x + 2)"
	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not '(x + 2)'")
		return
	}
	return
}
func TestFunctionApplication(t *testing.T) {
	ts := []struct {
		input    string
		expected int64
	}{
		{"let identity=fn(x){x;} identity(5);", 5},
		{"let identity=fn(x){return x;}; identity(5);", 5},
		{"let double = fn(x){x*2;};double(5);", 10},
		{"let add = fn(x,y){x+y;}; add(5,5);", 10},
		{"let add = fn(x,y){x+y;}; add(5+5,add(5,5));", 20},
		{"fn(x){x;}(5)", 5},
	}
	for _, tt := range ts {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}
