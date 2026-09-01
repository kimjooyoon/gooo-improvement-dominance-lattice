# Contribution boundary

This repository is PR-first and CI-only for build, test, vet, lint, format,
check, conformance, and integration evidence. Do not run those validations
locally. GitHub Actions is the verification authority.

The evaluator is read-only with respect to its inputs. All runtime output must
be caller-owned temporary output. Repository writes, apply, commit, merge, tag,
release, and cross-project required gates are recorded as zero authority in the
product runtime contract.

Do not introduce a scalar score, average, weighted sum, percentage, or scalar
state. Keep the indicator vector, its exact integer deltas, and partial-order
relations visible in receipts and dossiers.
