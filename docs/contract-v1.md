# Contract v1

The fixed contract contains six named indicator cells and three proof choices.
It requires exact integer pairs, direction-aware per-cell observations, and a
partial order made from dependency and explicit precedence relations. The
precedence field is metadata, never a weighted ranking.

State precedence is `REFUTED > UNKNOWN > CLOSED`. `REFUTED` records a known
contradiction or authority breach. `UNKNOWN` records unresolved evidence and
must include `stage`, `step`, `reason`, `unknown_class`, `next_operation`, and
`blocked_by`. `CLOSED` is emitted only by the strict dominance condition.

The contract forbids `overall_score`, `average`, `weighted_sum`, `percentage`,
and `scalar_state`. The only selection action is the explicit per-candidate
`SELECT`, `HOLD`, or `REJECT` paired with its relation and state.
