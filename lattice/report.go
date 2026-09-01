package lattice

import (
	"fmt"
	"sort"
)

func BuildDossier(summary Summary, runtime RuntimeMetrics, contract Contract) (Dossier, error) {
	if summary.Schema != "gooo/improvement-dominance-lattice/summary/v1" { return Dossier{}, fmt.Errorf("invalid summary schema") }
	if err := ValidateContract(contract); err != nil { return Dossier{}, err }
	if len(summary.Receipts) != 9 { return Dossier{}, fmt.Errorf("canonical denominator requires exactly nine receipts") }
	counts := map[string]int{"normal": 0, "unknown": 0, "refuted": 0}
	seen := map[string]bool{}
	var sourceDigest, contractDigest string
	for index, receipt := range summary.Receipts {
		if receipt.Schema != ReceiptSchema || receipt.CaseID == "" || seen[receipt.CaseID] { return Dossier{}, fmt.Errorf("receipt %d is missing or repeated", index) }
		seen[receipt.CaseID] = true
		counts[string(receipt.Class)]++
		if receipt.Class != CaseNormal && receipt.Class != CaseUnknown && receipt.Class != CaseRefuted { return Dossier{}, fmt.Errorf("receipt %q has an invalid class", receipt.CaseID) }
		if index == 0 { sourceDigest, contractDigest = receipt.Source.Digest, receipt.Contract.Digest } else if receipt.Contract.Digest != contractDigest { return Dossier{}, fmt.Errorf("contract digest differs across receipts") }
		if receipt.Source.Digest == "" || receipt.SemanticIR.Digest == "" || receipt.GeneratedGo.Digest == "" || receipt.Contract.Digest == "" { return Dossier{}, fmt.Errorf("receipt %q is not artifact-bound", receipt.CaseID) }
		for _, unknown := range receipt.Candidate.Unknown { if err := ValidateUnknown(unknown); err != nil { return Dossier{}, fmt.Errorf("receipt %q: %w", receipt.CaseID, err) } }
		switch receipt.Class {
		case CaseNormal:
			if receipt.Candidate.State != StateClosed || receipt.Candidate.Relation != RelationDominates || receipt.Candidate.Action != ActionSelect { return Dossier{}, fmt.Errorf("normal case %q is not DOMINATES/CLOSED", receipt.CaseID) }
		case CaseUnknown:
			if receipt.Candidate.State != StateUnknown || receipt.Candidate.Action != ActionHold || (receipt.Candidate.Relation != RelationIncomparable && receipt.Candidate.Relation != RelationUnknown) { return Dossier{}, fmt.Errorf("unknown case %q is not an UNKNOWN hold", receipt.CaseID) }
		case CaseRefuted:
			if receipt.Candidate.State != StateRefuted || receipt.Candidate.Action != ActionReject { return Dossier{}, fmt.Errorf("refuted case %q is not rejected") }
		}
	}
	if counts[string(CaseNormal)] != 3 || counts[string(CaseUnknown)] != 3 || counts[string(CaseRefuted)] != 3 { return Dossier{}, fmt.Errorf("case denominator must be normal=3 unknown=3 refuted=3") }
	if authorityExceeded(runtime.Authority) { return Dossier{}, fmt.Errorf("runtime authority boundary is open") }
	if runtime.OperationalRefuted != "OPERATIONAL_REFUTED: runtime input/repository writes/apply/commit/merge/tag/release authority=0; caller-owned output only" { return Dossier{}, fmt.Errorf("operational refuted record is not exact") }
	if len(runtime.LocalValidationCommands) == 0 { return Dossier{}, fmt.Errorf("local validation commands are not recorded") }
	ordered := append([]DecisionReceipt(nil), summary.Receipts...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].CaseID < ordered[j].CaseID })
	return Dossier{
		Schema: DossierSchema, Decision: "CONFORMANCE_CLOSED", SourceDigest: sourceDigest, ContractDigest: contractDigest,
		CaseCounts: counts, Receipts: ordered, Runtime: runtime,
		DecisionPath: []string{"source .gooo", "semantic IR", "generated Go evaluator", "executed per-indicator vector", "decision receipt", "human dossier"},
		Forbidden: contract.ForbiddenOutputs,
	}, nil
}
