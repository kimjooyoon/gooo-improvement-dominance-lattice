package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/gooo-improvement-dominance-lattice/lattice"
)

func main() {
	if len(os.Args) < 2 { fail(fmt.Errorf("usage: gooo-lattice <compile|generate|execute|dossier>")) }
	var err error
	switch os.Args[1] {
	case "compile": err = compileCommand(os.Args[2:])
	case "generate": err = generateCommand(os.Args[2:])
	case "execute": err = executeCommand(os.Args[2:])
	case "dossier": err = dossierCommand(os.Args[2:])
	default: err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil { fail(err) }
}

func compileCommand(args []string) error {
	fs := flag.NewFlagSet("compile", flag.ContinueOnError)
	sourcePath := fs.String("source", "", "source .gooo")
	contractPath := fs.String("contract", "", "semantic contract")
	outPath := fs.String("out", "", "semantic IR output")
	if err := fs.Parse(args); err != nil { return err }
	if *sourcePath == "" || *contractPath == "" || *outPath == "" { return fmt.Errorf("compile requires -source, -contract, and -out") }
	sourceDigest, source, err := lattice.DigestFile(*sourcePath); if err != nil { return err }
	var contract lattice.Contract
	if err := lattice.ReadJSON(*contractPath, &contract); err != nil { return err }
	contractDigest, _, err := lattice.DigestFile(*contractPath); if err != nil { return err }
	ir, err := lattice.Compile(string(source), *sourcePath, contract, contractDigest); if err != nil { return err }
	if ir.SourceDigest != sourceDigest { return fmt.Errorf("source digest changed during compile") }
	return lattice.WriteJSON(*outPath, ir)
}

func generateCommand(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	irPath := fs.String("ir", "", "semantic IR")
	fixturePath := fs.String("fixture", "", "fixture")
	outDir := fs.String("out-dir", "", "absolute caller-owned output directory")
	manifestPath := fs.String("manifest", "", "generated artifact manifest")
	if err := fs.Parse(args); err != nil { return err }
	if *irPath == "" || *fixturePath == "" || *outDir == "" || *manifestPath == "" { return fmt.Errorf("generate requires -ir, -fixture, -out-dir, and -manifest") }
	var ir lattice.SemanticIR; if err := lattice.ReadJSON(*irPath, &ir); err != nil { return err }
	var fixture lattice.Fixture; if err := lattice.ReadJSON(*fixturePath, &fixture); err != nil { return err }
	binding, err := lattice.GenerateEvaluator(ir, fixture, *outDir); if err != nil { return err }
	return lattice.WriteJSON(*manifestPath, struct { Schema string `json:"schema"`; CaseID string `json:"case_id"`; Evaluator lattice.ArtifactBinding `json:"evaluator"` }{"gooo/improvement-dominance-lattice/generated-manifest/v1", fixture.CaseID, binding})
}

func executeCommand(args []string) error {
	fs := flag.NewFlagSet("execute", flag.ContinueOnError)
	irPath := fs.String("ir", "", "semantic IR")
	fixturePath := fs.String("fixture", "", "fixture")
	contractPath := fs.String("contract", "", "semantic contract")
	generatedGoPath := fs.String("generated-go", "", "generated evaluator Go")
	outPath := fs.String("out", "", "decision receipt output")
	if err := fs.Parse(args); err != nil { return err }
	if *irPath == "" || *fixturePath == "" || *contractPath == "" || *generatedGoPath == "" || *outPath == "" { return fmt.Errorf("execute requires -ir, -fixture, -contract, -generated-go, and -out") }
	var ir lattice.SemanticIR; if err := lattice.ReadJSON(*irPath, &ir); err != nil { return err }
	var fixture lattice.Fixture; if err := lattice.ReadJSON(*fixturePath, &fixture); err != nil { return err }
	var contract lattice.Contract; if err := lattice.ReadJSON(*contractPath, &contract); err != nil { return err }
	if err := lattice.ValidateGeneratedGo(*generatedGoPath); err != nil { return err }
	irDigest, irBytes, err := lattice.DigestFile(*irPath); if err != nil { return err }
	goDigest, goBytes, err := lattice.DigestFile(*generatedGoPath); if err != nil { return err }
	contractDigest, contractBytes, err := lattice.DigestFile(*contractPath); if err != nil { return err }
	sourceBytes, sourceErr := os.ReadFile(ir.SourcePath)
	if sourceErr != nil { sourceBytes = nil }
	sourceBinding := lattice.ArtifactBinding{Stage: "GOOO_SOURCE", Path: ir.SourcePath, Digest: ir.SourceDigest, Bytes: len(sourceBytes)}
	irBinding := lattice.ArtifactBinding{Stage: "SEMANTIC_IR", Path: *irPath, Digest: irDigest, Bytes: len(irBytes)}
	generatedBinding := lattice.ArtifactBinding{Stage: "GENERATED_GO_EVALUATOR", Path: filepath.Clean(*generatedGoPath), Digest: goDigest, Bytes: len(goBytes)}
	contractBinding := lattice.ArtifactBinding{Stage: "CONTRACT", Path: *contractPath, Digest: contractDigest, Bytes: len(contractBytes)}
	receipt, err := lattice.EvaluateFixture(ir, fixture, contract, sourceBinding, irBinding, generatedBinding, contractBinding); if err != nil { return err }
	return lattice.WriteJSON(*outPath, receipt)
}

func dossierCommand(args []string) error {
	fs := flag.NewFlagSet("dossier", flag.ContinueOnError)
	contractPath := fs.String("contract", "", "semantic contract")
	summaryPath := fs.String("summary", "", "summary of receipts")
	runtimePath := fs.String("runtime", "", "exact runtime metrics")
	outJSONPath := fs.String("out-json", "", "machine dossier")
	outMarkdownPath := fs.String("out-md", "", "human dossier")
	if err := fs.Parse(args); err != nil { return err }
	if *contractPath == "" || *summaryPath == "" || *runtimePath == "" || *outJSONPath == "" || *outMarkdownPath == "" { return fmt.Errorf("dossier requires -contract, -summary, -runtime, -out-json, and -out-md") }
	var contract lattice.Contract; if err := lattice.ReadJSON(*contractPath, &contract); err != nil { return err }
	var summary lattice.Summary; if err := lattice.ReadJSON(*summaryPath, &summary); err != nil { return err }
	var runtime lattice.RuntimeMetrics; if err := lattice.ReadJSON(*runtimePath, &runtime); err != nil { return err }
	dossier, err := lattice.BuildDossier(summary, runtime, contract); if err != nil { return err }
	if err := lattice.WriteJSON(*outJSONPath, dossier); err != nil { return err }
	return lattice.WriteBytes(*outMarkdownPath, []byte(lattice.RenderDossierMarkdown(dossier)))
}

func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(2) }
