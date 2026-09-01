package lattice

const (
	SourceSchema          = "gooo/improvement-dominance-lattice/source/v1"
	IRSchema              = "gooo/improvement-dominance-lattice/semantic-ir/v1"
	FixtureSchema         = "gooo/improvement-dominance-lattice/fixture/v1"
	EvaluatorSchema       = "gooo/improvement-dominance-lattice/generated-evaluator/v1"
	ReceiptSchema         = "gooo/improvement-dominance-lattice/decision-receipt/v1"
	DossierSchema         = "gooo/improvement-dominance-lattice/human-dossier/v1"
	RuntimeSchema         = "gooo/improvement-dominance-lattice/runtime/v1"
	ContractSchema        = "gooo/improvement-dominance-lattice/contract/v1"
	StateClosed   State   = "CLOSED"
	StateUnknown  State   = "UNKNOWN"
	StateRefuted  State   = "REFUTED"
	RelationDominates Relation = "DOMINATES"
	RelationIncomparable Relation = "INCOMPARABLE"
	RelationUnknown Relation = "UNKNOWN"
	ActionSelect Action = "SELECT"
	ActionHold   Action = "HOLD"
	ActionReject Action = "REJECT"
	DirectionMinimize Direction = "minimize"
	DirectionMaximize Direction = "maximize"
	DirectionExact Direction = "exact"
	RoleObjective Role = "objective"
	RoleGuardrail Role = "guardrail"
	RoleObservation Role = "observation"
	GuardrailNone Guardrail = "none"
	GuardrailNonIncrease Guardrail = "non_increase"
	GuardrailNonDecrease Guardrail = "non_decrease"
	GuardrailExact Guardrail = "exact"
	ProofFoundation ProofChoice = "FOUNDATION"
	ProofCoherence ProofChoice = "COHERENCE"
	ProofRegression ProofChoice = "REGRESSION"
	ObservationImproved Observation = "IMPROVED"
	ObservationUnchanged Observation = "UNCHANGED"
	ObservationRegressed Observation = "REGRESSED"
	ObservationUnknown Observation = "UNKNOWN"
	CaseNormal CaseClass = "normal"
	CaseUnknown CaseClass = "unknown"
	CaseRefuted CaseClass = "refuted"
)

type State string
type Relation string
type Action string
type Direction string
type Role string
type Guardrail string
type ProofChoice string
type Observation string
type CaseClass string

type IndicatorDecl struct {
	ID              string      `json:"id"`
	Before          int         `json:"before"`
	After           int         `json:"after"`
	Direction      Direction   `json:"direction"`
	ClaimedRelation string     `json:"claimed_relation"`
	Role            Role        `json:"role"`
	Guardrail       Guardrail  `json:"guardrail"`
	Budget          *int        `json:"budget_kib"`
	ProofChoice     ProofChoice `json:"proof_choice"`
	Dependency      string      `json:"dependency"`
	Authority       string      `json:"authority"`
	Precedence      int         `json:"precedence"`
}

type Program struct {
	Schema      string         `json:"schema"`
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Candidates  []string       `json:"candidates"`
	Indicators  []IndicatorDecl `json:"indicators"`
}

type IRIndicator struct {
	IndicatorDecl
	SourceLine int `json:"source_line"`
}

type LatticeNode struct {
	ID         string      `json:"id"`
	DependsOn  []string    `json:"depends_on"`
	Precedence int         `json:"precedence"`
	ProofChoice ProofChoice `json:"proof_choice"`
}

type LatticeEdge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
}

type PartialOrder struct {
	Nodes    []LatticeNode `json:"nodes"`
	Edges    []LatticeEdge `json:"edges"`
	Ordering string        `json:"ordering"`
}

type SemanticIR struct {
	Schema         string         `json:"schema"`
	SourcePath     string         `json:"source_path"`
	SourceDigest   string         `json:"source_digest"`
	ContractDigest string         `json:"contract_digest"`
	Program        string         `json:"program"`
	Indicators     []IRIndicator  `json:"indicators"`
	Lattice        PartialOrder   `json:"lattice"`
}

type MetricPair struct {
	Before *int `json:"before"`
	After  *int `json:"after"`
}

type AuthorityCounts struct {
	RuntimeInputReads       int `json:"runtime_input_reads"`
	RepositoryWrites        int `json:"repository_writes"`
	Apply                   int `json:"apply"`
	Commit                  int `json:"commit"`
	Merge                   int `json:"merge"`
	Tag                     int `json:"tag"`
	Release                 int `json:"release"`
	CrossProjectRequiredGates int `json:"cross_project_required_gates"`
}

type CandidateObservation struct {
	CandidateID       string                    `json:"candidate_id"`
	Indicators        map[string]MetricPair     `json:"indicators"`
	ProofOverrides    map[string][]string        `json:"proof_overrides,omitempty"`
	Authority         AuthorityCounts            `json:"authority"`
	Counterexample    bool                      `json:"counterexample"`
}

type ExpectedDecision struct {
	State    State    `json:"state"`
	Relation Relation `json:"relation"`
	Action   Action   `json:"action"`
}

type Fixture struct {
	Schema      string               `json:"schema"`
	CaseID      string               `json:"case_id"`
	Description string               `json:"description"`
	Class       CaseClass            `json:"class"`
	Source      string               `json:"source"`
	Candidate   CandidateObservation `json:"candidate"`
	Expected    ExpectedDecision     `json:"expected"`
}

type MetricResult struct {
	IndicatorID     string       `json:"indicator_id"`
	Before          *int         `json:"before"`
	After           *int         `json:"after"`
	Delta           *int         `json:"delta"`
	Direction       Direction    `json:"direction"`
	ClaimedRelation string       `json:"claimed_relation"`
	Role            Role         `json:"role"`
	Guardrail       Guardrail    `json:"guardrail"`
	Budget          *int         `json:"budget_kib"`
	ProofChoice     ProofChoice  `json:"proof_choice"`
	Dependency      string       `json:"dependency"`
	Authority       string       `json:"authority"`
	Precedence      int          `json:"precedence"`
	Observation     Observation  `json:"observation"`
	Reason          string       `json:"reason"`
}

type UnknownDetail struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type Violation struct {
	IndicatorID string `json:"indicator_id,omitempty"`
	Kind        string `json:"kind"`
	Reason      string `json:"reason"`
}

type CandidateDecision struct {
	CandidateID string          `json:"candidate_id"`
	State       State           `json:"state"`
	Relation    Relation        `json:"relation"`
	Action      Action          `json:"action"`
	Reason      string          `json:"reason"`
	Indicators  []MetricResult  `json:"indicators"`
	Unknown     []UnknownDetail `json:"unknown,omitempty"`
	Violations  []Violation     `json:"violations,omitempty"`
}

type ArtifactBinding struct {
	Stage  string `json:"stage"`
	Path   string `json:"path"`
	Digest string `json:"digest"`
	Bytes  int    `json:"bytes"`
}

type DecisionReceipt struct {
	Schema          string            `json:"schema"`
	CaseID          string            `json:"case_id"`
	Class           CaseClass         `json:"class"`
	Description     string            `json:"description"`
	Candidate       CandidateDecision `json:"candidate"`
	Source          ArtifactBinding   `json:"source"`
	SemanticIR      ArtifactBinding   `json:"semantic_ir"`
	GeneratedGo     ArtifactBinding   `json:"generated_go_evaluator"`
	Contract        ArtifactBinding   `json:"contract"`
	Authority       AuthorityCounts   `json:"authority"`
	DecisionPath    []string          `json:"decision_path"`
	HumanDossier    string            `json:"human_dossier"`
}

type Contract struct {
	Schema             string      `json:"schema"`
	ContractID         string      `json:"contract_id"`
	IndicatorIDs       []string    `json:"indicator_ids"`
	ProofChoices       []string    `json:"proof_choices"`
	Directions         []string    `json:"directions"`
	ObservationStates  []string    `json:"observation_states"`
	StatePrecedence    []string    `json:"state_precedence"`
	UnknownFields      []string    `json:"unknown_fields"`
	ForbiddenOutputs   []string    `json:"forbidden_outputs"`
	AuthorityZero      bool        `json:"authority_zero"`
}

type PhaseMetrics struct {
	WallMs       int `json:"wall_ms"`
	PeakRSSKib   int `json:"peak_rss_kib"`
}

type InventoryMetrics struct {
	DescendantDirectories int `json:"descendant_directories"`
	RegularFiles          int `json:"regular_files_root_readme_excluded"`
	GoFiles               int `json:"go_files"`
	GoPhysicalLines       int `json:"go_physical_lines"`
	GoooFiles             int `json:"gooo_files"`
	GoooPhysicalLines     int `json:"gooo_physical_lines"`
}

type RuntimeMetrics struct {
	Schema              string          `json:"schema"`
	Compile             PhaseMetrics    `json:"compile"`
	Build               PhaseMetrics    `json:"build"`
	Test                PhaseMetrics    `json:"test"`
	Conformance         PhaseMetrics    `json:"conformance"`
	Integration         PhaseMetrics    `json:"integration"`
	TestsExecuted       int             `json:"tests_executed"`
	TestsReused         int             `json:"tests_reused"`
	TestsSkipped        int             `json:"tests_skipped"`
	GeneratedArtifacts  int             `json:"generated_artifact_files"`
	GeneratedBytes      int             `json:"generated_artifact_bytes"`
	Inventory           InventoryMetrics `json:"inventory"`
	Authority           AuthorityCounts `json:"authority"`
	LocalValidationCommands []string    `json:"local_validation_commands"`
	OperationalRefuted  string          `json:"operational_refuted"`
}

type Dossier struct {
	Schema          string            `json:"schema"`
	Decision        string            `json:"decision"`
	SourceDigest    string            `json:"source_digest"`
	ContractDigest  string            `json:"contract_digest"`
	CaseCounts      map[string]int    `json:"case_counts"`
	Receipts        []DecisionReceipt `json:"receipts"`
	Runtime         RuntimeMetrics    `json:"runtime"`
	DecisionPath    []string          `json:"decision_path"`
	Forbidden       []string          `json:"forbidden_outputs"`
}

type Summary struct {
	Schema   string            `json:"schema"`
	Receipts []DecisionReceipt `json:"receipts"`
}

var StatePrecedence = []State{StateRefuted, StateUnknown, StateClosed}
var ObservationOrder = []Observation{ObservationImproved, ObservationUnchanged, ObservationRegressed, ObservationUnknown}
