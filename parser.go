package main

import (
	"unicode"
)

type DatalogSyntax struct {
	Expressions []ExpressionSyntax `@@*`
	Query       *AtomSyntax        `("-" @@ "?")?`
}

type ExpressionSyntax struct {
	Head AtomSyntax            `@@`
	Body []NegatableAtomSyntax `( ":" "-" (@@ ","?)+ )? "."`
}

type NegatableAtomSyntax struct {
	Negated string     `("\\" "+")?`
	Atom    AtomSyntax `@@`
}

type AtomSyntax struct {
	Predicate string   `@Ident`
	Terms     []string `"(" (@Ident ","?)+ ")"`
}

func ConstructTerm(ts string) Term {
	var termType string
	if unicode.IsUpper(rune(ts[0])) {
		termType = "VAR"
	} else {
		termType = "SYM"
	}
	return Term{
		Type: termType,
		Name: ts,
	}
}

func ConstructAtom(as AtomSyntax) Atom {
	terms := make([]Term, len(as.Terms))
	for i, ts := range as.Terms {
		terms[i] = ConstructTerm(ts)
	}
	return Atom{
		PredicateSymbol: as.Predicate,
		Terms:           terms,
	}
}

func ConstructProgram(ast DatalogSyntax) Program {
	program := make([]Rule, len(ast.Expressions))
	for i, expression := range ast.Expressions {
		body := make([]Atom, len(expression.Body))
		for j, bs := range expression.Body {
			body[j] = ConstructAtom(bs.Atom)
		}
		program[i] = Rule{
			Head: ConstructAtom(expression.Head),
			Body: body,
		}
	}
	return program
}

func ConstructQuery(ast DatalogSyntax) *Rule {
	if ast.Query == nil {
		return nil
	}
	atom := ConstructAtom(*ast.Query)
	queryVars := make([]Term, 0)
	for _, term := range atom.Terms {
		if term.Type == "VAR" {
			queryVars = append(queryVars, term)
		}
	}
	return &Rule{
		Head: Atom{QUERY_PREDICATE, queryVars},
		Body: []Atom{atom},
	}
}
