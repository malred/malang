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
		{`"hello" - "world`, "unknown operator: STRING - STRING"},
		{`{"name": "Monkey"}[fn(x) {x}];`, "unusable as hash key: FUNCTION"},
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
func TestStringLiteral(t *testing.T) {
	input := `"hello world"`
	eval := testEval(input)
	str, ok := eval.(*object.String)
	if !ok {
		t.Fatalf("object is not string. got=%T(%+v)", eval, eval)
	}
	if str.Value != "hello world" {
		t.Errorf("string has wrong value. got=%q", str.Value)
	}
}
func TestStringConcatenation(t *testing.T) {
	input := `"hello" +" " +"world"`
	eval := testEval(input)
	str, ok := eval.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T(%+v)", eval, eval)
	}
	if str.Value != "hello world" {
		t.Errorf("string has wrong value. got=%q", str.Value)
	}
}
func TestBuiltinFunctions(t *testing.T) {
	ts := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported. got INTEGER"},
		{`len("one","two")`, "wrong number of arguments. got=2, want=1"},
	}
	for _, tt := range ts {
		eval := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, eval, int64(expected))
		case string:
			errobj, ok := eval.(*object.Error)
			if !ok {
				t.Errorf("obj is not error.")
				continue
			}
			if errobj.Message != expected {
				t.Errorf("wrong error message.")
			}
		}
	}
}
func TestArrayLiterals(t *testing.T) {
	input := `[1, 2 * 2, 3 + 3]`

	eval := testEval(input)
	res, ok := eval.(*object.Array)
	if !ok {
		t.Fatalf("exp not Array. got=%T", eval)
	}
	if len(res.Elements) != 3 {
		t.Fatalf("len(array.Elements) not 3. got=%d", len(res.Elements))
	}
	testIntegerObject(t, res.Elements[0], 1)
	testIntegerObject(t, res.Elements[1], 4)
	testIntegerObject(t, res.Elements[2], 6)
}
func TestArrayIndexExpression(t *testing.T) {
	ts := []struct {
		input    string
		expected interface{}
	}{
		{"[1,2,3][0]", 1},
		{"[1,2,3][1]", 2},
		{"[1,2,3][2]", 3},
		{"let i = 0; [1][i];", 1},
		{"[1,2,3][1+1]", 3},
		{"let myArr=[1,2,3]; myArr[2];", 3},
		{"let myArr=[1,2,3]; myArr[0]+myArr[1]+myArr[2];", 6},
		{"let myArr = [1,2,3]; let i = myArr[0]; myArr[i];", 2},
		{"[1,2,3][3]", nil},
		{"[1,2,3][-1]", nil},
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
func TestHashLiterals(t *testing.T) {
	input := `let two = "two";
	{
		"one": 10-9,
		two: 1+1,
		"thr"+"ee": 6/2,
		4: 4,
		true: 5,
		false: 6
	}
	`
	eval := testEval(input)
	res, ok := eval.(*object.Hash)
	if !ok {
		t.Fatalf("eval didn't return hash. got=%T(%+v)", eval, eval)
	}
	exp := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		TRUE.HashKey():                             5,
		FALSE.HashKey():                            6,
	}
	if len(res.Pairs) != len(exp) {
		t.Fatalf("hash has wrong num of pairs: %d", len(res.Pairs))
	}
	for expKey, expVal := range exp {
		pair, ok := res.Pairs[expKey]
		if !ok {
			t.Errorf("no pair for given key in pairs")
		}
		testIntegerObject(t, pair.Value, expVal)
	}
}
func TestHashIndexExpression(t *testing.T) {
	ts := []struct {
		input string
		exp   interface{}
	}{
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["bar"]`,
			nil,
		},
		{
			`let key = "foo"; {"foo": 5}[key]`,
			5,
		},
		{
			`{}["foo"]`,
			nil,
		},
		{
			`{5: 5}[5]`,
			5,
		},
		{
			`{true: 5}[true]`,
			5,
		},
		{
			`{false: 5}[false]`,
			5,
		},
	}
	for _, tt := range ts {
		eval := testEval(tt.input)
		integer, ok := tt.exp.(int)
		if ok {
			testIntegerObject(t, eval, int64(integer))
		} else {
			testNullObject(t, eval)
		}
	}
}
