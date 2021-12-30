package main

import "fmt"

const (
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

type Rule struct {
	Head Atom
	Body []Atom
}

type Program []Rule

type KnowledgeBase []Atom

// The hardest word to type in the English language
type Substitution map[Term]Term

var emptySubstition Substitution = make(map[Term]Term)

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
			if existing.PredicateSymbol == v.PredicateSymbol {
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
			if existing.PredicateSymbol == v.PredicateSymbol {
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
		return nil
	}
	if len(bodyAtom.Terms) != len(fact.Terms) {
		return nil
	}
	subs := make(map[Term]Term)
	// Walk both term lists
	fmt.Println("Unifying")
	for i := 0; i < len(bodyAtom.Terms); i++ {
		bodyTerm := bodyAtom.Terms[i]
		factTerm := fact.Terms[i]
		if factTerm.Type == SYM && bodyTerm.Type == SYM {
			if bodyTerm.Name != factTerm.Name {
				return nil
			}
		}
		if factTerm.Type == SYM && bodyTerm.Type == VAR {
			if existingSym, found := subs[bodyTerm]; found {
				if existingSym != factTerm {
					// Contradictory variable assignment
					// e.g., unifying p(X, X) with p("A", "B")
					return nil
				}
			} else {
				subs[bodyTerm] = factTerm
			}
		}
		if factTerm.Type == VAR {
			panic(fmt.Sprintf("Fact atom term %d (name: %s) must be ground, not a variable",
				i, factTerm.Name))
		}
	}

	return subs
}

func (kb KnowledgeBase) EvalAtom(bodyAtom Atom, substitutions []Substitution) []Substitution {
	result := make([]Substitution, len(substitutions))
	for i, substitution := range substitutions {
		grounded := bodyAtom.Substitute(substitution)
		fmt.Println("Grounded:", grounded)
		for _, fact := range kb {
			extension := grounded.unify(fact)
			fmt.Println("Unified extension:", extension)
			if extension != nil {
				substitution = Merge(substitution, extension)
			}
		}
		result[i] = substitution
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
	atoms := make([]Atom, len(rule.Body))
	fmt.Println("Evaluating rule:", rule.Head.PredicateSymbol)
	fmt.Println("Rule body:", rule.Body)
	fmt.Println("KB:", kb)
	if len(rule.Body) == 0 {
		fmt.Printf("FACT\n\n")
		return []Atom{rule.Head}
	}
	for i, subs := range kb.WalkBody(rule.Body) {
		fmt.Println("Potential subs:", subs)
		atoms[i] = rule.Head.Substitute(subs)
	}
	fmt.Println("AFTER ATOMS:", atoms)
	fmt.Println("")
	return atoms
}

func ImmediateConsequence(program Program, kb KnowledgeBase) KnowledgeBase {
	for _, rule := range program {
		kb = MergeKBs(kb, kb.EvalRule(rule))
	}
	return kb
}

func (rule Rule) IsRangeRestricted() bool {
	if len(rule.Body) == 0 {
		return true
	}
	for _, headTerm := range rule.Head.Terms {
		exists := false
		if headTerm.Type == VAR {
			for _, bodyAtom := range rule.Body {
				for _, bodyAtomTerm := range bodyAtom.Terms {
					if bodyAtomTerm.Type == VAR {
						if headTerm == bodyAtomTerm {
							exists = true
						}
					}
				}
			}
		}
		if !exists {
			return false
		}
	}
	return true
}

func Solve(program Program) KnowledgeBase {
	for _, rule := range program {
		if !rule.IsRangeRestricted() {
			panic(fmt.Sprintf("Rule %s is not range restricted", rule.Head.PredicateSymbol))
		}
	}
	kb := make([]Atom, 0)
	oldKB := kb
	for {
		oldKB = kb
		kb = ImmediateConsequence(program, kb)
		if len(kb) == len(oldKB) {
			fmt.Println("done")
			fmt.Println(kb)
			return kb
		} else {
			//fmt.Println("old:", len(oldKB))
			//fmt.Println("new:", len(kb))
			//return
		}
	}
}

func Query(program Program, query Rule) {
	kb := Solve(append(program, query))
	fmt.Println("Query result:")
	for _, atom := range kb {
		if atom.PredicateSymbol == query.Head.PredicateSymbol {
			fmt.Println(atom)
		} else {
		}
	}
	fmt.Println(".")
}

func main() {
	fact1 := Rule{
		Head: Atom{"first", []Term{{SYM, "a"}}},
		Body: []Atom{},
	}
	rule1 := Rule{
		Head: Atom{"second", []Term{{VAR, "X"}}},
		Body: []Atom{{"first", []Term{{VAR, "X"}}}},
	}
	query1 := Rule{
		Head: Atom{"query1", []Term{{VAR, "Y"}}},
		Body: []Atom{{"second", []Term{{VAR, "Y"}}}},
	}
	program := []Rule{fact1, rule1}
	Query(program, query1)
}
