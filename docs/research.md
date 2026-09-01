# Research note: dominance without scalarization

## Problem

Self-improvement candidates can improve one indicator and regress another.
Reducing that vector to a score hides the exact regression and makes a policy
decision look more certain than its evidence. The useful semantic boundary is
therefore a product order over named indicators, with hard constraints checked
before any dominance claim.

## Chosen semantics

Each source declaration names an indicator, integer `before` and `after`, a
direction (`minimize`, `maximize`, or `exact`), a claimed relation, role,
guardrail or budget, exactly one proof choice, dependency, authority, and
precedence. Go only lowers this declaration to IR and evaluates the supplied
vector; it does not infer missing intent.

The fixed delta is `after - before`. Direction maps that integer to one of the
four independent observations: `IMPROVED`, `UNCHANGED`, `REGRESSED`, or
`UNKNOWN`. Missing values remain JSON `null` and produce `UNKNOWN`.

For candidate `c`, `DOMINATES/CLOSED` is emitted only if:

1. no authority boundary is exceeded;
2. every known hard guardrail and budget is satisfied;
3. every proof choice is exactly one of `FOUNDATION`, `COHERENCE`, or
   `REGRESSION`;
4. at least one explicit objective is `IMPROVED`; and
5. no explicit objective is `REGRESSED` or unresolved.

Any hard guardrail regression, declared-claim counterexample, or authority
excess is `REFUTED`. An unresolved proof, null metric, or objective trade-off
is held as `UNKNOWN`; objective trade-offs are explicitly labeled
`INCOMPARABLE`. No Munchausen-trilemma proof choice is selected automatically.

## Why the real vector matters

The proof-aware test reuse fixture is not a claim about this evaluator's own
performance. It is an input vector used to test the semantics:

`build wall 127→119`, `build RSS 65500→67576`, `test wall 5→5`,
`test RSS 3480→3428`, `selected/executed 1→0` proof-gated, and `reused 0→1`.

The build-RSS delta is +2076 KiB. A hard non-increase policy must reject it.
An explicit +4096 KiB budget policy can close it while retaining that delta in
the receipt. The different outcomes demonstrate policy sensitivity without
scalarization.

## Boundary

Runtime input, repository writes, apply, commit, merge, tag, release, and
cross-project required gates all have authority count zero. Outputs are written
only to caller-owned temporary paths. CI is the sole build/test/conformance
authority; local validation commands are recorded as prohibited.
