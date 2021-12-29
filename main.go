package main

import "fmt"

const (
	VAR = iota
	SYM = iota
)

type Term struct {
	Type int
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
			if existing.PredicateSymbol != v.PredicateSymbol {
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
			if existing.PredicateSymbol != v.PredicateSymbol {
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
	for i := 0; i < len(bodyAtom.Terms); i++ {
		bodyTerm := bodyAtom.Terms[i]
		factTerm := fact.Terms[i]
		if factTerm.Type == VAR {
			panic(fmt.Sprintf("Fact atom term %d (name: %s) must be ground, not a variable",
				i, factTerm.Name))
		}
		if bodyTerm.Type == SYM { // factTerm.Type is implicitly SYM too
			if bodyTerm != factTerm {
				return nil
			}
		} else { // bodyTerm.Type == VAR
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
	}

	return subs
}

func (kb KnowledgeBase) EvalAtom(bodyAtom Atom, substitutions []Substitution) []Substitution {
	result := make([]Substitution, len(substitutions))
	for i, substitution := range substitutions {
		grounded := bodyAtom.Substitute(substitution)
		extension := bodyAtom.unify(grounded)
		if extension == nil {
			result[i] = substitution
		} else {
			result[i] = Merge(substitution, extension)
		}
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
	for i, subs := range kb.WalkBody(rule.Body) {
		atoms[i] = rule.Head.Substitute(subs)
	}
	return atoms
}

func ImmediateConsequence(program Program, kb KnowledgeBase) KnowledgeBase {
	for _, rule := range program {
		kb = MergeKBs(kb, kb.EvalRule(rule))
	}
	return kb
}

func (rule Rule) IsRangeRestricted() bool {
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

func Solve(program Program) {
	for _, rule := range program {
		if !rule.IsRangeRestricted() {
			panic(fmt.Sprintf("Rule %s is not range restricted", rule.Head.PredicateSymbol))
		}
	}
	kb := make([]Atom, 0)
	oldKB := kb
	for {
		kb = ImmediateConsequence(program, kb)
		if len(kb) == len(oldKB) {
			fmt.Println("done")
			return
		} else {
			return
		}
	}
}

func main() {
	fmt.Println("hello!")
}
