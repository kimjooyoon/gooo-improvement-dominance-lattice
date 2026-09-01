# gooo-improvement-dominance-lattice

`gooo-improvement-dominance-lattice` is a deterministic, proof-aware evaluator
for self-improvement candidate vectors. It keeps every integer indicator and
its relation visible. A candidate is `DOMINATES/CLOSED` only when every hard
guardrail is satisfied, at least one explicit objective improves, and no
explicit objective regresses.

The pipeline is:

`.gooo` → semantic IR / partial-order lattice → generated Go evaluator → executed vector → decision receipt → human dossier

The evaluator fixes `delta = after - before` and classifies each metric as
`IMPROVED`, `UNCHANGED`, `REGRESSED`, or `UNKNOWN`. It never creates an
overall score, average, weighted sum, percentage, or scalar state. Trade-offs
without an explicit resolving budget or priority remain `INCOMPARABLE/UNKNOWN`
with all six operational coordinates: `stage`, `step`, `reason`,
`unknown_class`, `next_operation`, and `blocked_by`.

## Fixed vector and policies

The real-shaped fixture for `gooo-proof-aware-test-reuse v0.1.2` is fixed at:

| indicator | before | after | direction |
|---|---:|---:|---|
| build wall | 127 | 119 | minimize |
| build RSS (KiB) | 65500 | 67576 | minimize |
| test wall | 5 | 5 | exact |
| test RSS (KiB) | 3480 | 3428 | minimize |
| selected/executed (proof-gated) | 1 | 0 | minimize |
| reused | 0 | 1 | maximize |

Under a build-RSS hard `non_increase` policy, the +2076 KiB delta is
`REFUTED`. Under an explicit +4096 KiB budget, the same vector is
`DOMINATES/CLOSED` because all other conditions hold. These are separate
policy decisions; they are never combined into a score.

## Canonical conformance

The CI corpus has exactly nine cases: three normal/CLOSED, three
UNKNOWN/hold, and three REFUTED/reject. It covers normal selection,
incomparability, missing proof, unknown top-level propagation, guardrail
regression, counterexample, and authority excess.

The root README is excluded from physical-file inventory counts. Runtime
receipts record compile/build/test/conformance/integration wall time and peak
RSS, test counts, Go and Gooo files plus physical lines, directory/file counts,
generated artifact counts/bytes, exact local-validation prohibitions, and the
`OPERATIONAL_REFUTED` zero-authority record.

## Verification and release

Build, test, and conformance validation is intentionally prohibited locally.
GitHub Actions runs the full workflow with Go 1.27, preserves failed runs and
uploaded evidence, and uploads the decision dossier and receipts as a
SHA-bound artifact. The release process is PR-first: merge to `main`, create a
fresh immutable tag, publish the evidence bundle and digest manifest, and never
rewrite the tag or release.

Release boundary: [docs/release-v0.1.0.md](docs/release-v0.1.0.md).

## Inventory

```text
cmd/gooo-lattice/       CLI for compile, generate, execute, and dossier
lattice/                parser, semantic IR, partial-order evaluator, reports
examples/               .gooo policy declarations including RSS variants
fixtures/               real vector and nine canonical cases
contracts/              fixed semantic and authority contract
scripts/                CI-only conformance runner
docs/                   research and contract notes
.gooo/                  ownership and release-boundary receipts
```

## License

MIT.
