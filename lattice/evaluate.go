package lattice

import (
	"fmt"
	"sort"
	"strings"
)

func ValidateFixture(fixture Fixture, ir SemanticIR) error {
	if fixture.Schema != FixtureSchema || fixture.CaseID == "" || fixture.Description == "" || fixture.Source == "" { return fmt.Errorf("invalid fixture header") }
	if fixture.Class != CaseNormal && fixture.Class != CaseUnknown && fixture.Class != CaseRefuted { return fmt.Errorf("invalid fixture class %q", fixture.Class) }
	if fixture.Candidate.CandidateID == "" { return fmt.Errorf("fixture candidate id is empty") }
	if fixture.Expected.State != StateClosed && fixture.Expected.State != StateUnknown && fixture.Expected.State != StateRefuted { return fmt.Errorf("invalid expected state") }
	if fixture.Expected.Action != ActionSelect && fixture.Expected.Action != ActionHold && fixture.Expected.Action != ActionReject { return fmt.Errorf("invalid expected action") }
	if fixture.Expected.Relation != RelationDominates && fixture.Expected.Relation != RelationIncomparable && fixture.Expected.Relation != RelationUnknown { return fmt.Errorf("invalid expected relation") }
	if len(ir.Indicators) == 0 { return fmt.Errorf("fixture cannot be evaluated without indicators") }
	return nil
}

func EvaluateFixture(ir SemanticIR, fixture Fixture, contract Contract, sourceBinding, irBinding, generatedBinding, contractBinding ArtifactBinding) (DecisionReceipt, error) {
	if err := ValidateContract(contract); err != nil { return DecisionReceipt{}, err }
	if ir.Schema != IRScheme || ir.ContractDigest != contractBinding.Digest { return DecisionReceipt{}, fmt.Errorf("semantic IR is not bound to the contract") }
	if err := ValidateFixture(fixture, ir); err != nil { return DecisionReceipt{}, err }

	decision := evaluateCandidate(ir, fixture.Candidate)
	receipt := DecisionReceipt{
		Schema: ReceiptSchema, CaseID: fixture.CaseID, Class: fixture.Class, Description: fixture.Description,
		Candidate: decision, Source: sourceBinding, SemanticIR: irBinding, GeneratedGo: generatedBinding, Contract: contractBinding,
		Authority: fixture.Candidate.Authority,
		DecisionPath: []string{
			"delta := after - before",
			"classify each indicator independently as IMPROVED, UNCHANGED, REGRESSED, or UNKNOWN",
			"evaluate hard guardrails and authority before dominance",
			"DOMINATES/CLOSED requires an improved explicit objective and zero regressed objectives",
			"state precedence is REFUTED > UNKNOWN > CLOSED",
		},
	}
	receipt.HumanDossier = RenderReceiptDossier(receipt)
	return receipt, nil
}

func evaluateCandidate(ir SemanticIR, candidate CandidateObservation) CandidateDecision {
	decision := CandidateDecision{CandidateID: candidate.CandidateID, State: StateClosed, Relation: RelationDominates, Action: ActionSelect}
	byID := map[string]IRIndicator{}
	for _, indicator := range ir.Indicators { byID[indicator.ID] = indicator }
	ids := make([]string, 0, len(ir.Indicators))
	for _, indicator := range ir.Indicators { ids = append(ids, indicator.ID) }
	sort.SliceStable(ids, func(i, j int) bool { a, b := byID[ids[i]], byID[ids[j]]; if a.Precedence != b.Precedence { return a.Precedence < b.Precedence }; return a.ID < b.ID })
	objectiveImproved, objectiveRegressed, hasUnknown := false, false, false
	for _, id := range ids {
		indicator := byID[id]
		pair := candidate.Indicators[id]
		result := MetricResult{
			IndicatorID: indicator.ID, Before: pair.Before, After: pair.After, Direction: indicator.Direction,
			ClaimedRelation: indicator.ClaimedRelation, Role: indicator.Role, Guardrail: indicator.Guardrail,
			Budget: indicator.Budget, ProofChoice: indicator.ProofChoice, Dependency: indicator.Dependency,
			Authority: indicator.Authority, Precedence: indicator.Precedence, Observation: ObservationUnknown,
		}
		if choices, overridden := candidate.ProofOverrides[id]; overridden {
			if len(choices) != 1 || !validProofChoice(ProofChoice(firstOrEmpty(choices))) || choices[0] != string(indicator.ProofChoice) {
				result.Reason = "proof choice is missing, plural, unknown, or conflicts with the semantic IR"
				decision.Unknown = append(decision.Unknown, UnknownDetail{Stage: "execute", Step: "proof-choice-validation", Reason: result.Reason, UnknownClass: proofUnknownClass(choices), NextOperation: "supply exactly one declared proof choice", BlockedBy: []string{id}})
				hasUnknown = true
				decision.Indicators = append(decision.Indicators, result)
				continue
			}
		}
		if pair.Before == nil || pair.After == nil {
			result.Reason = "before or after is null; the metric remains UNKNOWN"
			decision.Unknown = append(decision.Unknown, UnknownDetail{Stage: "execute", Step: "metric-pair-validation", Reason: result.Reason, UnknownClass: "MISSING_METRIC", NextOperation: "provide both integer observations", BlockedBy: []string{id}})
			hasUnknown = true
			decision.Indicators = append(decision.Indicators, result)
			continue
		}
		delta := *pair.After - *pair.Before
		result.Delta = &delta
		result.Observation = classify(indicator.Direction, delta)
		result.Reason = observationReason(result.Observation, delta)
		if claimViolation(indicator, result.Observation, delta) {
			decision.Violations = append(decision.Violations, Violation{IndicatorID: id, Kind: "COUNTEREXAMPLE", Reason: "observed relation contradicts the declared claimed relation"})
		}
		if guardrailViolation(indicator, delta) {
			decision.Violations = append(decision.Violations, Violation{IndicatorID: id, Kind: "HARD_GUARDRAIL_REGRESSION", Reason: "observed regression exceeds the declared hard guardrail or budget"})
		}
		if indicator.Role == RoleObjective {
			if result.Observation == ObservationImproved { objectiveImproved = true }
			if result.Observation == ObservationRegressed { objectiveRegressed = true }
		}
		decision.Indicators = append(decision.Indicators, result)
	}
	if candidate.Counterexample { decision.Violations = append(decision.Violations, Violation{Kind: "COUNTEREXAMPLE", Reason: "fixture declares a preserved counterexample"}) }
	if authorityExceeded(candidate.Authority) { decision.Violations = append(decision.Violations, Violation{Kind: "AUTHORITY_EXCEEDED", Reason: "runtime input/repository writes/apply/commit/merge/tag/release or external gate authority is non-zero"}) }
	if len(decision.Violations) > 0 {
		decision.State, decision.Relation, decision.Action = StateRefuted, RelationUnknown, ActionReject
		decision.Reason = "REFUTED: a hard guardrail, counterexample, or authority violation takes precedence over unresolved observations."
		return decision
	}
	if hasUnknown {
		decision.State, decision.Relation, decision.Action = StateUnknown, RelationUnknown, ActionHold
		decision.Reason = "UNKNOWN: an unresolved metric or proof choice prevents a closed partial-order decision."
		return decision
	}
	if objectiveRegressed {
		decision.State, decision.Relation, decision.Action = StateUnknown, RelationIncomparable, ActionHold
		decision.Reason = "INCOMPARABLE/UNKNOWN: an explicit objective regresses; no scalar, average, weighted sum, or percentage resolves the trade-off."
		return decision
	}
	if !objectiveImproved {
		decision.State, decision.Relation, decision.Action = StateUnknown, RelationUnknown, ActionHold
		decision.Reason = "UNKNOWN: no explicit objective improved, so dominance is not established."
		return decision
	}
	decision.Reason = "DOMINATES/CLOSED: every hard guardrail is satisfied, at least one explicit objective improves, and no explicit objective regresses."
	return decision
}

func classify(direction Direction, delta int) Observation {
	switch direction {
	case DirectionMinimize:
		if delta < 0 { return ObservationImproved }; if delta > 0 { return ObservationRegressed }
	case DirectionMaximize:
		if delta > 0 { return ObservationImproved }; if delta < 0 { return ObservationRegressed }
	case DirectionExact:
		if delta != 0 { return ObservationRegressed }
	}
	return ObservationUnchanged
}

func observationReason(observation Observation, delta int) string {
	return fmt.Sprintf("delta=%d; after-before is fixed and direction maps this vector cell to %s", delta, observation)
}

func claimViolation(indicator IRIndicator, observation Observation, delta int) bool {
	switch indicator.ClaimedRelation {
	case "OBSERVED": return false
	case "IMPROVED": return observation != ObservationImproved
	case "UNCHANGED": return observation != ObservationUnchanged
	case "REGRESSED": return observation != ObservationRegressed
	case "NON_INCREASE": return delta > 0
	case "NON_DECREASE": return delta < 0
	case "WITHIN_BUDGET": return indicator.Budget == nil || regressionDelta(indicator.Direction, delta) > *indicator.Budget
	default: return true
	}
}

func guardrailViolation(indicator IRIndicator, delta int) bool {
	if indicator.Role != RoleGuardrail { return false }
	switch indicator.Guardrail {
	case GuardrailNonIncrease: return delta > 0 && indicator.Budget == nil
	case GuardrailNonDecrease: return delta < 0 && indicator.Budget == nil
	case GuardrailExact: return delta != 0
	}
	if indicator.Budget != nil { return regressionDelta(indicator.Direction, delta) > *indicator.Budget }
	return false
}

func regressionDelta(direction Direction, delta int) int {
	if direction == DirectionMaximize { return -delta }
	return delta
}

func authorityExceeded(authority AuthorityCounts) bool {
	return authority.RuntimeInputReads != 0 || authority.RepositoryWrites != 0 || authority.Apply != 0 || authority.Commit != 0 || authority.Merge != 0 || authority.Tag != 0 || authority.Release != 0 || authority.CrossProjectRequiredGates != 0
}

func firstOrEmpty(values []string) string { if len(values) == 0 { return "" }; return values[0] }

func proofUnknownClass(choices []string) string {
	switch len(choices) { case 0: return "MISSING_PROOF_CHOICE"; case 1: return "UNKNOWN_PROOF_CHOICE"; default: return "MULTIPLE_PROOF_CHOICES" }
}

func RenderReceiptDossier(receipt DecisionReceipt) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %s/%s/%s\n", receipt.CaseID, receipt.Candidate.State, receipt.Candidate.Relation, receipt.Candidate.Action)
	fmt.Fprintf(&b, "reason: %s\n", receipt.Candidate.Reason)
	for _, indicator := range receipt.Candidate.Indicators {
		fmt.Fprintf(&b, "%s: before=%s after=%s delta=%s observation=%s proof=%s\n", indicator.IndicatorID, pointerString(indicator.Before), pointerString(indicator.After), pointerString(indicator.Delta), indicator.Observation, indicator.ProofChoice)
	}
	for _, unknown := range receipt.Candidate.Unknown { fmt.Fprintf(&b, "unknown: stage=%s step=%s reason=%s unknown_class=%s next_operation=%s blocked_by=%s\n", unknown.Stage, unknown.Step, unknown.Reason, unknown.UnknownClass, unknown.NextOperation, strings.Join(unknown.BlockedBy, ",")) }
	return b.String()
}

func pointerString(value *int) string { if value == nil { return "null" }; return fmt.Sprintf("%d", *value) }

func ValidateUnknown(detail UnknownDetail) error {
	if detail.Stage == "" || detail.Step == "" || detail.Reason == "" || detail.UnknownClass == "" || detail.NextOperation == "" || len(detail.BlockedBy) == 0 { return fmt.Errorf("unknown detail must contain stage, step, reason, unknown_class, next_operation, and blocked_by") }
	return nil
}
