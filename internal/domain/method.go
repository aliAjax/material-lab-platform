package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type FieldType string

const (
	FieldNumber  FieldType = "number"
	FieldText    FieldType = "text"
	FieldBoolean FieldType = "boolean"
)

type MethodStatus string

const (
	MethodDraft     MethodStatus = "draft"
	MethodPublished MethodStatus = "published"
	MethodArchived  MethodStatus = "archived"
)

type Method struct {
	ID            string        `json:"id"`
	Code          string        `json:"code"`
	Name          string        `json:"name"`
	MaterialScope string        `json:"materialScope"`
	Version       int           `json:"version"`
	Status        MethodStatus  `json:"status"`
	Fields        []MethodField `json:"fields"`
	Formula       string        `json:"formula"`
	Precision     int32         `json:"precision"`
	Rule          string        `json:"rule"`
	CreatedBy     string        `json:"createdBy"`
	PublishedAt   time.Time     `json:"publishedAt,omitempty"`
	CreatedAt     time.Time     `json:"createdAt"`
}
type MethodField struct {
	Name     string           `json:"name"`
	Label    string           `json:"label"`
	Type     FieldType        `json:"type"`
	Unit     string           `json:"unit"`
	Min      *decimal.Decimal `json:"min,omitempty"`
	Max      *decimal.Decimal `json:"max,omitempty"`
	Required bool             `json:"required"`
}

func (m *Method) Validate() error {
	if err := Require(m.Code, "code"); err != nil {
		return err
	}
	if err := Require(m.Name, "name"); err != nil {
		return err
	}
	if m.Version < 1 {
		return fmt.Errorf("%w: version", ErrValidation)
	}
	if len(m.Fields) == 0 {
		return fmt.Errorf("%w: fields required", ErrValidation)
	}
	known := map[string]bool{}
	for _, f := range m.Fields {
		if strings.TrimSpace(f.Name) == "" || !f.TypeValid() {
			return fmt.Errorf("%w: invalid method field", ErrValidation)
		}
		if known[f.Name] {
			return fmt.Errorf("%w: duplicate field", ErrValidation)
		}
		known[f.Name] = true
		if f.Min != nil && f.Max != nil && f.Min.GreaterThan(*f.Max) {
			return fmt.Errorf("%w: field range", ErrValidation)
		}
	}
	if strings.TrimSpace(m.Formula) == "" {
		return fmt.Errorf("%w: formula required", ErrValidation)
	}
	if _, err := ParseExpression(m.Formula, known); err != nil {
		return err
	}
	return nil
}
func (f MethodField) TypeValid() bool {
	return f.Type == FieldNumber || f.Type == FieldText || f.Type == FieldBoolean
}
func (m *Method) Publish(now time.Time) error {
	if m.Status != MethodDraft {
		return ErrInvalidTransition
	}
	if err := m.Validate(); err != nil {
		return err
	}
	m.Status, m.PublishedAt = MethodPublished, now
	return nil
}

type Expression interface {
	Eval(values map[string]decimal.Decimal) (decimal.Decimal, error)
	Variables() []string
}
type numberExpr struct{ value decimal.Decimal }

func (e numberExpr) Eval(map[string]decimal.Decimal) (decimal.Decimal, error) { return e.value, nil }
func (e numberExpr) Variables() []string                                      { return nil }

type variableExpr struct{ name string }

func (e variableExpr) Eval(v map[string]decimal.Decimal) (decimal.Decimal, error) {
	n, ok := v[e.name]
	if !ok {
		return decimal.Zero, fmt.Errorf("%w: missing variable %s", ErrValidation, e.name)
	}
	return n, nil
}
func (e variableExpr) Variables() []string { return []string{e.name} }

type binaryExpr struct {
	op          byte
	left, right Expression
}

func (e binaryExpr) Eval(v map[string]decimal.Decimal) (decimal.Decimal, error) {
	a, err := e.left.Eval(v)
	if err != nil {
		return decimal.Zero, err
	}
	b, err := e.right.Eval(v)
	if err != nil {
		return decimal.Zero, err
	}
	switch e.op {
	case '+':
		return a.Add(b), nil
	case '-':
		return a.Sub(b), nil
	case '*':
		return a.Mul(b), nil
	case '/':
		if b.IsZero() {
			return decimal.Zero, fmt.Errorf("%w: division by zero", ErrValidation)
		}
		return a.Div(b), nil
	}
	return decimal.Zero, fmt.Errorf("%w: operator", ErrValidation)
}
func (e binaryExpr) Variables() []string { return append(e.left.Variables(), e.right.Variables()...) }

type callExpr struct {
	name string
	args []Expression
}

func (e callExpr) Eval(v map[string]decimal.Decimal) (decimal.Decimal, error) {
	if len(e.args) == 0 {
		return decimal.Zero, fmt.Errorf("%w: empty function", ErrValidation)
	}
	vals := make([]decimal.Decimal, len(e.args))
	for i, a := range e.args {
		x, err := a.Eval(v)
		if err != nil {
			return decimal.Zero, err
		}
		vals[i] = x
	}
	switch strings.ToLower(e.name) {
	case "min":
		r := vals[0]
		for _, x := range vals[1:] {
			if x.LessThan(r) {
				r = x
			}
		}
		return r, nil
	case "max":
		r := vals[0]
		for _, x := range vals[1:] {
			if x.GreaterThan(r) {
				r = x
			}
		}
		return r, nil
	case "avg":
		sum := decimal.Zero
		for _, x := range vals {
			sum = sum.Add(x)
		}
		return sum.Div(decimal.NewFromInt(int64(len(vals)))), nil
	}
	return decimal.Zero, fmt.Errorf("%w: function %s", ErrValidation, e.name)
}
func (e callExpr) Variables() []string {
	var out []string
	for _, a := range e.args {
		out = append(out, a.Variables()...)
	}
	return out
}

type parser struct {
	input string
	pos   int
	known map[string]bool
}

func ParseExpression(input string, known map[string]bool) (Expression, error) {
	p := &parser{input: strings.TrimSpace(input), known: known}
	e, err := p.parseAdd()
	if err != nil {
		return nil, err
	}
	p.skip()
	if p.pos != len(p.input) {
		return nil, fmt.Errorf("%w: unexpected token", ErrValidation)
	}
	return e, nil
}
func (p *parser) skip() {
	for p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t') {
		p.pos++
	}
}
func (p *parser) parseAdd() (Expression, error) {
	l, err := p.parseMul()
	if err != nil {
		return nil, err
	}
	for {
		p.skip()
		if p.pos >= len(p.input) || (p.input[p.pos] != '+' && p.input[p.pos] != '-') {
			break
		}
		op := p.input[p.pos]
		p.pos++
		r, e := p.parseMul()
		if e != nil {
			return nil, e
		}
		l = binaryExpr{op: op, left: l, right: r}
	}
	return l, nil
}
func (p *parser) parseMul() (Expression, error) {
	l, err := p.parseAtom()
	if err != nil {
		return nil, err
	}
	for {
		p.skip()
		if p.pos >= len(p.input) || (p.input[p.pos] != '*' && p.input[p.pos] != '/') {
			break
		}
		op := p.input[p.pos]
		p.pos++
		r, e := p.parseAtom()
		if e != nil {
			return nil, e
		}
		l = binaryExpr{op: op, left: l, right: r}
	}
	return l, nil
}
func (p *parser) parseAtom() (Expression, error) {
	p.skip()
	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("%w: expression ended", ErrValidation)
	}
	if p.input[p.pos] == '(' {
		p.pos++
		e, err := p.parseAdd()
		p.skip()
		if p.pos >= len(p.input) || p.input[p.pos] != ')' {
			return nil, fmt.Errorf("%w: missing )", ErrValidation)
		}
		p.pos++
		return e, err
	}
	start := p.pos
	for p.pos < len(p.input) && ((p.input[p.pos] >= '0' && p.input[p.pos] <= '9') || p.input[p.pos] == '.') {
		p.pos++
	}
	if p.pos > start {
		d, err := decimal.NewFromString(p.input[start:p.pos])
		if err != nil {
			return nil, fmt.Errorf("%w: number", ErrValidation)
		}
		return numberExpr{d}, nil
	}
	for p.pos < len(p.input) && ((p.input[p.pos] >= 'a' && p.input[p.pos] <= 'z') || (p.input[p.pos] >= 'A' && p.input[p.pos] <= 'Z') || p.input[p.pos] == '_') {
		p.pos++
	}
	if p.pos == start {
		return nil, fmt.Errorf("%w: token", ErrValidation)
	}
	name := p.input[start:p.pos]
	p.skip()
	if p.pos < len(p.input) && p.input[p.pos] == '(' {
		p.pos++
		var args []Expression
		for {
			a, err := p.parseAdd()
			if err != nil {
				return nil, err
			}
			args = append(args, a)
			p.skip()
			if p.pos >= len(p.input) {
				return nil, fmt.Errorf("%w: function", ErrValidation)
			}
			if p.input[p.pos] == ')' {
				p.pos++
				break
			}
			if p.input[p.pos] != ',' {
				return nil, fmt.Errorf("%w: function separator", ErrValidation)
			}
			p.pos++
		}
		return callExpr{name: name, args: args}, nil
	}
	if !p.known[name] {
		return nil, fmt.Errorf("%w: unknown variable %s", ErrValidation, name)
	}
	return variableExpr{name}, nil
}

func Calculate(m Method, values map[string]decimal.Decimal) (decimal.Decimal, error) {
	expr, err := ParseExpression(m.Formula, mapFields(m.Fields))
	if err != nil {
		return decimal.Zero, err
	}
	result, err := expr.Eval(values)
	if err != nil {
		return decimal.Zero, err
	}
	return result.Round(m.Precision), nil
}
func mapFields(fields []MethodField) map[string]bool {
	m := map[string]bool{}
	for _, f := range fields {
		m[f.Name] = true
	}
	return m
}
