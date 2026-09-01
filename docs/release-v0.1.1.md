# v0.1.1 release boundary

v0.1.1 preserves the v0.1.0 immutable release and adds a correction: every
`UNKNOWN` result, including an objective trade-off labeled
`INCOMPARABLE/UNKNOWN`, now carries all six operational coordinates.

The release is PR-first and CI-bound. Its immutable tag points to the merged
main commit, and its evidence asset is bound to the final main CI run by the
release manifest and SHA-256 digests.
