package main

import (
	"fmt"
	"math"
	"os"
	"text/scanner"
)

// 値
type Value float64

// 構文木の型
type Expr interface {
	Eval() Value
}

// 評価
func (e Value) Eval() Value {
	return e
}

// 単項演算子
type Op1 struct {
	code rune
	expr Expr
}

func newOp1(code rune, e Expr) Expr {
	return &Op1{code, e}
}

func (e *Op1) Eval() Value {
	v := e.expr.Eval()
	if e.code == '-' {
		v = -v
	}
	return v
}

// 二項演算子
type Op2 struct {
	code        rune
	left, right Expr
}

func newOp2(code rune, left, right Expr) Expr {
	return &Op2{code, left, right}
}

func (e *Op2) Eval() Value {
	x := e.left.Eval()
	y := e.right.Eval()
	switch e.code {
	case '+':
		return x + y
	case '-':
		return x - y
	case '*':
		return x * y
	case '/':
		return x / y
	default:
		panic(fmt.Errorf("invalid op code"))
	}
}

// 変数
type Variable string

// 大域的な環境
var globalEnv = make(map[Variable]Value)

// 変数の評価
func (v Variable) Eval() Value {
	val, ok := globalEnv[v]
	if !ok {
		panic(fmt.Errorf("unbound variable: %v", v))
	}
	return val
}

// 代入演算子
type Agn struct {
	name Variable
	expr Expr
}

func newAgn(v Variable, e Expr) *Agn {
	return &Agn{v, e}
}

// 代入演算子の評価
func (a *Agn) Eval() Value {
	val := a.expr.Eval()
	globalEnv[a.name] = val
	return val
}

// 組み込み関数
type Func interface {
	Argc() int
}

type Func1 func(float64) float64

func (f Func1) Argc() int {
	return 1
}

type Func2 func(float64, float64) float64

func (f Func2) Argc() int {
	return 2
}

// 組み込み関数の構文
type App struct {
	fn Func
	xs []Expr
}

func newApp(fn Func, xs []Expr) *App {
	return &App{fn, xs}
}

// 組み込み関数の評価
func (a *App) Eval() Value {
	switch f := a.fn.(type) {
	case *Func1:
		x := float64(a.xs[0].Eval())
		return Value(f.body(x))
	case *Func2:
		x := float64(a.xs[0].Eval())
		y := float64(a.xs[1].Eval())
		return Value(f.body(x, y))
	default:
		panic(fmt.Errorf("function Eval error"))
	}
}

// 組み込み関数の初期化
var funcTable = make(map[string]Func)

func initFunc() {
	funcTable["sqrt"] = Func1(math.Sqrt)
	funcTable["sin"] = Func1(math.Sin)
	funcTable["cos"] = Func1(math.Cos)
	funcTable["tan"] = Func1(math.Tan)
	funcTable["sinh"] = Func1(math.Sinh)
	funcTable["cosh"] = Func1(math.Cosh)
	funcTable["tanh"] = Func1(math.Tanh)
	funcTable["asin"] = Func1(math.Asin)
	funcTable["acos"] = Func1(math.Acos)
	funcTable["atan"] = Func1(math.Atan)
	funcTable["atan2"] = Func2(math.Atan2)
	funcTable["exp"] = Func1(math.Exp)
	funcTable["pow"] = Func2(math.Pow)
	funcTable["log"] = Func1(math.Log)
	funcTable["log10"] = Func1(math.Log10)
	funcTable["log2"] = Func1(math.Log2)
}

// 字句解析
type Lex struct {
	scanner.Scanner
	Token rune
}

func (lex *Lex) getToken() {
	lex.Token = lex.Scan()
}

// 引数の取得
func getArgs(lex *Lex) []Expr {
	e := make([]Expr, 0)
	if lex.Token != '(' {
		panic(fmt.Errorf("'(' expected"))
	}
	lex.getToken()
	if lex.Token == ')' {
		lex.getToken()
		return e
	}
	for {
		e = append(e, expression(lex))
		switch lex.Token {
		case ')':
			lex.getToken()
			return e
		case ',':
			lex.getToken()
		default:
			panic(fmt.Errorf("unexpected token in argument list"))
		}
	}
}

// 因子
func factor(lex *Lex) Expr {
	switch lex.Token {
	case '(':
		lex.getToken()
		e := expression(lex)
		if lex.Token != ')' {
			panic(fmt.Errorf("')' expected"))
		}
		lex.getToken()
		return e
	case '+':
		lex.getToken()
		return newOp1('+', factor(lex))
	case '-':
		lex.getToken()
		return newOp1('-', factor(lex))
	case scanner.Int, scanner.Float:
		var n float64
		fmt.Sscan(lex.TokenText(), &n)
		lex.getToken()
		return Value(n)
	case scanner.Ident:
		name := lex.TokenText()
		lex.getToken()
		if name == "quit" {
			panic(name)
		}
		v, ok := funcTable[name]
		if ok {
			xs := getArgs(lex)
			if len(xs) != v.Argc() {
				panic(fmt.Errorf("wrong number of arguments: %v", name))
			}
			return newApp(v, xs)
		} else {
			return Variable(name)
		}
	default:
		panic(fmt.Errorf("unexpected token: %v", lex.TokenText()))
	}
}

// 項
func term(lex *Lex) Expr {
	e := factor(lex)
	for {
		switch lex.Token {
		case '*':
			lex.getToken()
			e = newOp2('*', e, factor(lex))
		case '/':
			lex.getToken()
			e = newOp2('/', e, factor(lex))
		default:
			return e
		}
	}
}

// 式
func expr1(lex *Lex) Expr {
	e := term(lex)
	for {
		switch lex.Token {
		case '+':
			lex.getToken()
			e = newOp2('+', e, term(lex))
		case '-':
			lex.getToken()
			e = newOp2('-', e, term(lex))
		default:
			return e
		}
	}
}

func expression(lex *Lex) Expr {
	e := expr1(lex)
	if lex.Token == '=' {
		v, ok := e.(Variable)
		if ok {
			lex.getToken()
			return newAgn(v, expression(lex))
		} else {
			panic(fmt.Errorf("invalid assign form"))
		}
	}
	return e
}

// 式の入力と評価
func toplevel(lex *Lex) (r bool) {
	r = false
	defer func() {
		err := recover()
		if err != nil {
			mes, ok := err.(string)
			if ok && mes == "quit" {
				r = true
			} else {
				fmt.Fprintln(os.Stderr, err)
				for lex.Token != ';' {
					lex.getToken()
				}
			}
		}
	}()
	for {
		fmt.Print("Calc> ")
		lex.getToken()
		e := expression(lex)
		if lex.Token != ';' {
			panic(fmt.Errorf("invalid expression"))
		} else {
			fmt.Println(e.Eval())
		}
	}
	return r
}

func main() {
	var lex Lex
	lex.Init(os.Stdin)
	initFunc()
	for {
		if toplevel(&lex) {
			break
		}
	}
}
