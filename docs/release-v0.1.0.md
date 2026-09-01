# v0.1.0 release boundary

This release contains the proof-aware dominance lattice contract, the fixed
six-indicator real-shaped fixture, the nine-case canonical corpus, and the
CI-only evidence pipeline.

Release identity is the merged `main` commit plus one immutable `v0.1.0` tag.
The release asset is the CI evidence bundle. Its SHA-256 digest is recorded in
the release manifest and in the post-release audit. Failed runs, failed tags,
and failed release attempts are retained rather than deleted or rewritten.

The v0.1.0 policy examples make the RSS distinction explicit:

- `examples/dominance-lattice-rss-guardrail.gooo` → `REFUTED` for +2076 KiB.
- `examples/dominance-lattice-rss-budget.gooo` → `DOMINATES/CLOSED` for the
  same vector under the explicit +4096 KiB budget.

No selection is a claim about the evaluator's own performance. It is a
deterministic judgment of the supplied before/after vector.
