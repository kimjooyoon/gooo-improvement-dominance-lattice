package lattice

import "testing"

func TestClassifyUsesFixedDeltaDirection(t *testing.T) {
	if got := classify(DirectionMinimize, -8); got != ObservationImproved { t.Fatalf("minimize delta -8: got %s", got) }
	if got := classify(DirectionMaximize, 1); got != ObservationImproved { t.Fatalf("maximize delta 1: got %s", got) }
	if got := classify(DirectionExact, 0); got != ObservationUnchanged { t.Fatalf("exact delta 0: got %s", got) }
	if got := classify(DirectionExact, 1); got != ObservationRegressed { t.Fatalf("exact delta 1: got %s", got) }
}

func TestParseFailsClosedForProofChoiceErrors(t *testing.T) {
	missing := "program gooo-improvement-dominance-lattice v1\ncandidate c\nindicator x before=1 after=0 direction=minimize claimed=IMPROVED role=objective guardrail=none budget=none dependency=none authority=observe precedence=1"
	if _, err := ParseProgram(missing); err == nil { t.Fatal("missing proof choice accepted") }
	multiple := "program gooo-improvement-dominance-lattice v1\ncandidate c\nindicator x before=1 after=0 direction=minimize claimed=IMPROVED role=objective guardrail=none budget=none proof=FOUNDATION,COHERENCE dependency=none authority=observe precedence=1"
	if _, err := ParseProgram(multiple); err == nil { t.Fatal("multiple proof choices accepted") }
	unknown := "program gooo-improvement-dominance-lattice v1\ncandidate c\nindicator x before=1 after=0 direction=minimize claimed=IMPROVED role=objective guardrail=none budget=none proof=OTHER dependency=none authority=observe precedence=1"
	if _, err := ParseProgram(unknown); err == nil { t.Fatal("unknown proof choice accepted") }
}

func TestCandidateDecisionPreservesVectorAndUnknownCoordinates(t *testing.T) {
	ir := SemanticIR{Schema: IRSchema, Indicators: []IRIndicator{{IndicatorDecl: IndicatorDecl{ID: "x", Direction: DirectionMinimize, ClaimedRelation: "IMPROVED", Role: RoleObjective, Guardrail: GuardrailNone, ProofChoice: ProofFoundation, Dependency: "none", Authority: "observe", Precedence: 1}}}}
	before, after := 4, 2
	decision := evaluateCandidate(ir, CandidateObservation{CandidateID: "c", Indicators: map[string]MetricPair{"x": MetricPair{Before: &before, After: &after}}})
	if decision.State != StateClosed || decision.Relation != RelationDominates || decision.Action != ActionSelect { t.Fatalf("unexpected closed decision: %+v", decision) }
	if decision.Indicators[0].Delta == nil || *decision.Indicators[0].Delta != -2 || decision.Indicators[0].Observation != ObservationImproved { t.Fatalf("vector was not preserved: %+v", decision.Indicators[0]) }
	decision = evaluateCandidate(ir, CandidateObservation{CandidateID: "c", Indicators: map[string]MetricPair{"x": MetricPair{Before: &before, After: &after}}, ProofOverrides: map[string][]string{"x": []string{}}})
	if decision.State != StateUnknown || len(decision.Unknown) != 1 { t.Fatalf("missing proof did not fail closed: %+v", decision) }
	if err := ValidateUnknown(decision.Unknown[0]); err != nil { t.Fatal(err) }
}
