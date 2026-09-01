# CI and operational boundary

The repository uses GitHub Actions only for build, test, conformance, and
integration evidence. Local execution of those validation commands is
prohibited by the repository contract.

The runtime product has zero authority for runtime input mutation,
repository writes, apply, commit, merge, tag, release, and cross-project
required gates. It writes generated evaluator, receipt, and dossier files only
under a caller-owned temporary output directory. Failed CI runs and their
artifacts remain retained so an unsuccessful validation is evidence rather
than a deleted state.

The release is created only after the PR is green and merged to `main`. A new
tag is created once, the release and evidence asset are published once, and a
digest manifest binds the release to the commit and asset bytes. The tag and
release are never force-updated.
