package lattice

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func ParseProgram(source string) (Program, error) {
	program := Program{Schema: SourceSchema}
	seenIndicators := map[string]bool{}
	seenCandidates := map[string]bool{}
	programSeen := false
	for lineNumber, raw := range strings.Split(source, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") { continue }
		fields := strings.Fields(line)
		if len(fields) == 0 { continue }
		switch fields[0] {
		case "program":
			if programSeen || len(fields) != 3 || fields[1] != "gooo-improvement-dominance-lattice" || fields[2] != "v1" {
				return Program{}, fmt.Errorf("line %d: invalid program declaration", lineNumber+1)
			}
			programSeen = true
			program.Name, program.Version = fields[1], fields[2]
		case "candidate":
			if len(fields) != 2 || fields[1] == "" { return Program{}, fmt.Errorf("line %d: candidate requires an id", lineNumber+1) }
			if seenCandidates[fields[1]] { return Program{}, fmt.Errorf("line %d: duplicate candidate %q", lineNumber+1, fields[1]) }
			seenCandidates[fields[1]] = true
			program.Candidates = append(program.Candidates, fields[1])
		case "indicator":
			decl, err := parseIndicator(fields[1:], lineNumber+1)
			if err != nil { return Program{}, err }
			if seenIndicators[decl.ID] { return Program{}, fmt.Errorf("line %d: duplicate indicator %q", lineNumber+1, decl.ID) }
			seenIndicators[decl.ID] = true
			program.Indicators = append(program.Indicators, decl)
		default:
			return Program{}, fmt.Errorf("line %d: unknown declaration %q", lineNumber+1, fields[0])
		}
	}
	if !programSeen { return Program{}, fmt.Errorf("missing program declaration") }
	if len(program.Candidates) == 0 { return Program{}, fmt.Errorf("source must declare at least one candidate") }
	if len(program.Indicators) == 0 { return Program{}, fmt.Errorf("source must declare at least one indicator") }
	return program, nil
}

func parseIndicator(fields []string, line int) (IndicatorDecl, error) {
	if len(fields) < 1 { return IndicatorDecl{}, fmt.Errorf("line %d: indicator requires an id", line) }
	decl := IndicatorDecl{ID: fields[0], Guardrail: GuardrailNone, Role: RoleObservation, Dependency: "none", Authority: "observe"}
	seen := map[string]bool{}
	for _, token := range fields[1:] {
		parts := strings.SplitN(token, "=", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" { return IndicatorDecl{}, fmt.Errorf("line %d: malformed indicator attribute %q", line, token) }
		key, value := parts[0], parts[1]
		if seen[key] { return IndicatorDecl{}, fmt.Errorf("line %d: duplicate indicator attribute %q", line, key) }
		seen[key] = true
		switch key {
		case "before":
			valueInt, err := strconv.Atoi(value); if err != nil || valueInt < 0 { return IndicatorDecl{}, fmt.Errorf("line %d: before must be a non-negative integer", line) }; decl.Before = valueInt
		case "after":
			valueInt, err := strconv.Atoi(value); if err != nil || valueInt < 0 { return IndicatorDecl{}, fmt.Errorf("line %d: after must be a non-negative integer", line) }; decl.After = valueInt
		case "direction":
			decl.Direction = Direction(value)
		case "claimed":
			decl.ClaimedRelation = value
		case "role":
			decl.Role = Role(value)
		case "guardrail":
			decl.Guardrail = Guardrail(value)
		case "budget":
			if value == "none" { decl.Budget = nil; continue }
			valueInt, err := strconv.Atoi(value); if err != nil || valueInt < 0 { return IndicatorDecl{}, fmt.Errorf("line %d: budget must be none or a non-negative integer", line) }; decl.Budget = &valueInt
		case "proof":
			decl.ProofChoice = ProofChoice(value)
		case "dependency":
			decl.Dependency = value
		case "authority":
			decl.Authority = value
		case "precedence":
			valueInt, err := strconv.Atoi(value); if err != nil || valueInt < 0 { return IndicatorDecl{}, fmt.Errorf("line %d: precedence must be a non-negative integer", line) }; decl.Precedence = valueInt
		default:
			return IndicatorDecl{}, fmt.Errorf("line %d: unknown indicator attribute %q", line, key)
		}
	}
	for _, required := range []string{"before", "after", "direction", "claimed", "role", "guardrail", "budget", "proof", "dependency", "authority", "precedence"} {
		if !seen[required] { return IndicatorDecl{}, fmt.Errorf("line %d: missing indicator attribute %q", line, required) }
	}
	if !validDirection(decl.Direction) { return IndicatorDecl{}, fmt.Errorf("line %d: unknown direction %q", line, decl.Direction) }
	if !validRole(decl.Role) { return IndicatorDecl{}, fmt.Errorf("line %d: unknown role %q", line, decl.Role) }
	if !validGuardrail(decl.Guardrail) { return IndicatorDecl{}, fmt.Errorf("line %d: unknown guardrail %q", line, decl.Guardrail) }
	if !validProofChoice(decl.ProofChoice) { return IndicatorDecl{}, fmt.Errorf("line %d: proof choice must be exactly one FOUNDATION, COHERENCE, or REGRESSION", line) }
	if decl.Role == RoleGuardrail && decl.Guardrail == GuardrailNone && decl.Budget == nil { return IndicatorDecl{}, fmt.Errorf("line %d: guardrail role requires a hard guardrail or budget", line) }
	if decl.Role == RoleObjective && decl.Guardrail != GuardrailNone { return IndicatorDecl{}, fmt.Errorf("line %d: objective cannot also declare a hard guardrail", line) }
	return decl, nil
}

func Compile(source, sourcePath string, contract Contract, contractDigest string) (SemanticIR, error) {
	if err := ValidateContract(contract); err != nil { return SemanticIR{}, err }
	program, err := ParseProgram(source)
	if err != nil { return SemanticIR{}, err }
	if len(program.Indicators) != len(contract.IndicatorIDs) { return SemanticIR{}, fmt.Errorf("source indicator count %d does not match contract %d", len(program.Indicators), len(contract.IndicatorIDs)) }
	byID := map[string]bool{}
	for _, indicator := range program.Indicators { byID[indicator.ID] = true }
	for index, expected := range contract.IndicatorIDs {
		if program.Indicators[index].ID != expected { return SemanticIR{}, fmt.Errorf("indicator %q is not in contract position %d", program.Indicators[index].ID, index) }
	}
	for _, indicator := range program.Indicators {
		if indicator.Dependency != "none" && !byID[indicator.Dependency] { return SemanticIR{}, fmt.Errorf("indicator %q depends on unknown indicator %q", indicator.ID, indicator.Dependency) }
	}
	sourceDigest := DigestBytes([]byte(source))
	ir := SemanticIR{Schema: IRScheme, SourcePath: sourcePath, SourceDigest: sourceDigest, ContractDigest: contractDigest, Program: program.Name}
	for _, indicator := range program.Indicators {
		ir.Indicators = append(ir.Indicators, IRIndicator{IndicatorDecl: indicator, SourceLine: sourceLineForIndicator(source, indicator.ID)})
		ir.Lattice.Nodes = append(ir.Lattice.Nodes, LatticeNode{ID: indicator.ID, DependsOn: []string{indicator.Dependency}, Precedence: indicator.Precedence, ProofChoice: indicator.ProofChoice})
	}
	for _, indicator := range program.Indicators {
		if indicator.Dependency != "none" { ir.Lattice.Edges = append(ir.Lattice.Edges, LatticeEdge{From: indicator.Dependency, To: indicator.ID, Reason: "explicit dependency"}) }
	}
	for _, left := range program.Indicators {
		for _, right := range program.Indicators {
			if left.ID != right.ID && left.Precedence < right.Precedence { ir.Lattice.Edges = append(ir.Lattice.Edges, LatticeEdge{From: left.ID, To: right.ID, Reason: "explicit precedence; never a score"}) }
		}
	}
	sort.Slice(ir.Lattice.Edges, func(i, j int) bool { if ir.Lattice.Edges[i].From != ir.Lattice.Edges[j].From { return ir.Lattice.Edges[i].From < ir.Lattice.Edges[j].From }; return ir.Lattice.Edges[i].To < ir.Lattice.Edges[j].To })
	ir.Lattice.Ordering = "partial-order: dependency and explicit precedence relations only; no scalar ranking"
	return ir, nil
}

func sourceLineForIndicator(source, indicatorID string) int {
	for lineNumber, raw := range strings.Split(source, "\n") {
		fields := strings.Fields(strings.TrimSpace(raw))
		if len(fields) >= 2 && fields[0] == "indicator" && fields[1] == indicatorID { return lineNumber + 1 }
	}
	return 0
}

func validDirection(direction Direction) bool { return direction == DirectionMinimize || direction == DirectionMaximize || direction == DirectionExact }
func validRole(role Role) bool { return role == RoleObjective || role == RoleGuardrail || role == RoleObservation }
func validGuardrail(guardrail Guardrail) bool { return guardrail == GuardrailNone || guardrail == GuardrailNonIncrease || guardrail == GuardrailNonDecrease || guardrail == GuardrailExact }
func validProofChoice(choice ProofChoice) bool { return choice == ProofFoundation || choice == ProofCoherence || choice == ProofRegression }

func ValidateContract(contract Contract) error {
	if contract.Schema != ContractSchema || contract.ContractID == "" { return fmt.Errorf("invalid contract schema") }
	if len(contract.IndicatorIDs) != 6 { return fmt.Errorf("contract must contain exactly six indicators") }
	if len(contract.ProofChoices) != 3 || len(contract.Directions) != 3 || len(contract.ObservationStates) != 4 { return fmt.Errorf("contract categories are incomplete") }
	if len(contract.StatePrecedence) != 3 || contract.StatePrecedence[0] != string(StateRefuted) || contract.StatePrecedence[1] != string(StateUnknown) || contract.StatePrecedence[2] != string(StateClosed) { return fmt.Errorf("state precedence is not REFUTED > UNKNOWN > CLOSED") }
	for _, field := range []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"} {
		found := false; for _, contracted := range contract.UnknownFields { if contracted == field { found = true; break } }
		if !found { return fmt.Errorf("unknown field %q is not contracted", field) }
	}
	if !contract.AuthorityZero { return fmt.Errorf("authority-zero boundary is not contracted") }
	for _, forbidden := range []string{"overall_score", "average", "weighted_sum", "percentage", "scalar_state"} {
		found := false; for _, item := range contract.ForbiddenOutputs { if item == forbidden { found = true; break } }
		if !found { return fmt.Errorf("forbidden output %q is not contracted", forbidden) }
	}
	return nil
}
