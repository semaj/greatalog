package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/participle/v2"
)

const ( // Misc
	QUERY_PREDICATE = "__query"
)

const ( // Term Type
	VAR = "VAR"
	SYM = "SYM"
)

type Term struct {
	Type string
	Name string
}

type Atom struct {
	PredicateSymbol string
	Terms           []Term
}

func (a Atom) Equals(b Atom) bool {
	if a.PredicateSymbol != b.PredicateSymbol {
		return false
	}
	if len(a.Terms) != len(b.Terms) {
		return false
	}
	for i, aTerm := range a.Terms {
		bTerm := b.Terms[i]
		if aTerm.Type != bTerm.Type {
			return false
		}
		if aTerm.Name != bTerm.Name {
			return false
		}
	}
	return true
}

func (a Atom) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s(", a.PredicateSymbol)
	for i, term := range a.Terms {
		fmt.Fprintf(&b, term.Name)
		if i != len(a.Terms)-1 {
			fmt.Fprintf(&b, ",")
		} else {
			fmt.Fprintf(&b, ")")
		}
	}
	return b.String()
}

type Rule struct {
	Head Atom
	Body []Atom
}

func (r Rule) IsFact() bool {
	return len(r.Body) == 0
}

func (r Rule) String() string {
	var b strings.Builder
	b.WriteString(r.Head.String())
	if len(r.Body) == 0 {
		b.WriteString(".")
	} else {
		b.WriteString(" :- ")
		for i, atom := range r.Body {
			b.WriteString(atom.String())
			if i != len(r.Body)-1 {
				b.WriteString(", ")
			} else {
				b.WriteString(".")
			}
		}
	}
	return b.String()
}

type Program []Rule

func (p Program) String() string {
	var b strings.Builder
	for _, rule := range p {
		b.WriteString(rule.String())
		b.WriteString("\n")
	}
	return b.String()
}

type KnowledgeBase []Atom

func (kb KnowledgeBase) String() string {
	var b strings.Builder
	//b.WriteString("-- Knowledge Base --\n")
	for _, atom := range kb {
		b.WriteString(atom.String())
		b.WriteString(".\n")
	}
	//b.WriteString("----")
	return b.String()
}

// The hardest word to type in the English language
type Substitution map[Term]Term

func (s Substitution) String() string {
	var b strings.Builder
	b.WriteString("{\n")
	for k, v := range s {
		b.WriteString(" ")
		b.WriteString(k.Name)
		b.WriteString(fmt.Sprintf("(%s)", k.Type))
		b.WriteString(" => ")
		b.WriteString(v.Name)
		b.WriteString(fmt.Sprintf("(%s)", v.Type))
		b.WriteString(" ")
	}
	b.WriteString("\n}")
	return b.String()
}

// Convenience merge function
func Merge(subs1 Substitution, subs2 Substitution) Substitution {
	merged := make(map[Term]Term)
	for k, v := range subs1 {
		merged[k] = v
	}
	for k, v := range subs2 {
		merged[k] = v
	}
	return merged
}

func MergeKBs(kb1 KnowledgeBase, kb2 KnowledgeBase) KnowledgeBase {
	merged := make([]Atom, 0)
	for _, v := range kb1 {
		exists := false
		for _, existing := range merged {
			//if existing.PredicateSymbol == v.PredicateSymbol {
			if existing.Equals(v) {
				exists = true
			}
		}
		if !exists {
			merged = append(merged, v)
		}
	}
	for _, v := range kb2 {
		exists := false
		for _, existing := range merged {
			//if existing.PredicateSymbol == v.PredicateSymbol {
			if existing.Equals(v) {
				exists = true
			}
		}
		if !exists {
			merged = append(merged, v)
		}
	}
	return merged
}

func (atom Atom) Substitute(subs Substitution) Atom {
	newTerms := make([]Term, len(atom.Terms))
	for i, term := range atom.Terms {
		if term.Type == SYM {
			newTerms[i] = term
		} else { // == VAR
			if subbed, found := subs[term]; found {
				newTerms[i] = subbed
			} else {
				newTerms[i] = term
			}
		}
	}
	return Atom{
		PredicateSymbol: atom.PredicateSymbol,
		Terms:           newTerms,
	}
}

func (bodyAtom Atom) unify(fact Atom) Substitution {
	if bodyAtom.PredicateSymbol != fact.PredicateSymbol {
		//fmt.Println("1")
		return nil
	}
	if len(bodyAtom.Terms) != len(fact.Terms) {
		//fmt.Println("2")
		return nil
	}
	subs := make(map[Term]Term)
	// Walk both term lists
	for i := 0; i < len(bodyAtom.Terms); i++ {
		bodyTerm := bodyAtom.Terms[i]
		factTerm := fact.Terms[i]
		if factTerm.Type == SYM && bodyTerm.Type == SYM {
			if bodyTerm.Name != factTerm.Name {
				//fmt.Println("3")
				return nil
			}
		}
		if factTerm.Type == SYM && bodyTerm.Type == VAR {
			if existingSym, found := subs[bodyTerm]; found {
				if existingSym != factTerm {
					//fmt.Println("4")
					// Contradictory variable assignment
					// e.g., unifying p(X, X) with p("A", "B")
					return nil
				}
			} else {
				subs[bodyTerm] = factTerm
			}
		}
		if factTerm.Type == VAR {
			panic(fmt.Sprintf("Attempting to unify with `%s` atom as a fact, but it contains variables (term index %d, name %s). Unifying with body atom `%s`.", fact.String(), i, factTerm.Name, bodyAtom.String()))
		}
	}

	return subs
}

func (kb KnowledgeBase) EvalAtom(bodyAtom Atom, substitutions []Substitution) []Substitution {
	result := make([]Substitution, 0)
	for _, substitution := range substitutions {
		// TODO: reverse atom and subs
		grounded := bodyAtom.Substitute(substitution)
		for _, fact := range kb {
			//fmt.Printf("Attempting to unify body atom %s with fact %s, subs: %s\n", grounded.String(), fact.String(), substitution.String())
			extension := grounded.unify(fact)
			if extension != nil {
				result = append(result, Merge(substitution, extension))
				//fmt.Println("Successful unification:", result)
			} else {
				//fmt.Println("Failed unification.")
			}
		}
		//result[i] = substitution
	}
	return result
}

func (kb KnowledgeBase) WalkBody(bodyAtoms []Atom) []Substitution {
	substitutions := []Substitution{make(map[Term]Term)}
	for _, bodyAtom := range bodyAtoms {
		substitutions = kb.EvalAtom(bodyAtom, substitutions)
	}
	return substitutions
}

func (kb KnowledgeBase) EvalRule(rule Rule) KnowledgeBase {
	atoms := make([]Atom, 0)
	if len(rule.Body) == 0 {
		return []Atom{rule.Head}
	}
	for _, subs := range kb.WalkBody(rule.Body) {
		atoms = append(atoms, rule.Head.Substitute(subs))
	}
	return atoms
}

func ImmediateConsequence(program Program, kb KnowledgeBase) KnowledgeBase {
	for _, rule := range program {
		//fmt.Printf("Evaluating rule %s KB:\n%s\n", rule.String(), kb.String())
		kb = MergeKBs(kb, kb.EvalRule(rule))
	}
	return kb
}

func (rule Rule) IsRangeRestricted() bool {
	if len(rule.Body) == 0 {
		return true
	}
	for _, headTerm := range rule.Head.Terms {
		if headTerm.Type == VAR {
			exists := false
			for _, bodyAtom := range rule.Body {
				for _, bodyAtomTerm := range bodyAtom.Terms {
					if bodyAtomTerm.Type == VAR {
						if headTerm == bodyAtomTerm {
							exists = true
						}
					}
				}
			}
			if !exists {
				return false
			}
		}
	}
	return true
}

func Solve(program Program) KnowledgeBase {
	for _, rule := range program {
		if !rule.IsRangeRestricted() {
			panic(fmt.Sprintf("Rule %s is not range restricted", rule.String()))
		}
	}
	kb := KnowledgeBase(make([]Atom, 0))
	oldKB := kb
	for i := 0; ; i++ {
		oldKB = kb
		kb = ImmediateConsequence(program, kb)
		if len(kb) == len(oldKB) {
			return kb
		}
	}
}

func Query(program Program, query Rule) []Atom {
	result := make([]Atom, 0)
	kb := Solve(append(program, query))
	actualQueryAtom := query.Body[0]
	//fmt.Println("Query result:")
	for _, atom := range kb {
		if atom.PredicateSymbol == actualQueryAtom.PredicateSymbol &&
			len(atom.Terms) == len(actualQueryAtom.Terms) {
			match := true
			for i, queryTerm := range actualQueryAtom.Terms {
				term := atom.Terms[i]
				if queryTerm.Type == SYM &&
					!(term.Type == SYM && term.Name == queryTerm.Name) {
					match = false
				}
			}
			if match {
				result = append(result, atom)
			}
		}
	}
	return result
}

func Run(fileName string) []Atom {
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	parser, err := participle.Build(&DatalogSyntax{})
	if err != nil {
		panic(err)
	}
	ast := &DatalogSyntax{}
	err = parser.ParseString(fileName, string(bytes), ast)
	if err != nil {
		panic(err)
	}
	program := ConstructProgram(*ast)
	//fmt.Printf("Program:\n%s", program.String())
	//fmt.Println("")
	query := ConstructQuery(*ast)
	//fmt.Println("Query:", query.String())
	//fmt.Println("")
	return Query(program, *query)
}

func main() {
	args := os.Args[1:]
	firstArg := args[0]
	if firstArg == "TEST" {
		Test()
	} else {
		for _, atom := range Run(firstArg) {
			fmt.Printf("%s.\n", atom.String())
		}
	}
}
