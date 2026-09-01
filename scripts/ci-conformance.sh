#!/usr/bin/env bash
set -Eeuo pipefail
trap 'status=$?; echo "CI_CONFORMANCE_FAILURE line=$LINENO status=$status" >&2' ERR

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"

TEMP_ROOT=${RUNNER_TEMP:-/tmp}
WORK=$(mktemp -d "$TEMP_ROOT/gooo-improvement-dominance-lattice.XXXXXX")
EVIDENCE="$WORK/evidence"
BIN="$WORK/gooo-lattice"
CONTRACT="contracts/dominance-lattice-v1.json"
CASES=(
  closed-normal
  closed-budget
  closed-exact
  unknown-incomparable
  unknown-missing-proof
  unknown-top-level
  refuted-guardrail-regression
  refuted-counterexample
  refuted-authority
)
SOURCES=(
  examples/dominance-lattice.gooo
  examples/dominance-lattice-rss-budget.gooo
  examples/dominance-lattice-rss-guardrail.gooo
  examples/dominance-lattice-incomparable.gooo
)

mkdir -p "$WORK/irs" "$EVIDENCE/irs" "$EVIDENCE/cases"
declare -A PHASE_WALL=([compile]=0 [build]=0 [test]=0 [conformance]=0 [integration]=0)
declare -A PHASE_RSS=([compile]=0 [build]=0 [test]=0 [conformance]=0 [integration]=0)

max_value() {
  if [ "$1" -gt "$2" ]; then printf '%s\n' "$1"; else printf '%s\n' "$2"; fi
}

run_timed() {
  local label=$1
  local phase=$2
  shift 2
  local start_ns end_ns wall_ms rss_kib
  start_ns=$(date +%s%N)
  /usr/bin/time -f '%M' -o "$WORK/$label.rss" "$@"
  end_ns=$(date +%s%N)
  wall_ms=$(( (end_ns - start_ns) / 1000000 ))
  rss_kib=$(tr -d '[:space:]' < "$WORK/$label.rss")
  PHASE_WALL[$phase]=$(( PHASE_WALL[$phase] + wall_ms ))
  PHASE_RSS[$phase]=$(max_value "${PHASE_RSS[$phase]}" "$rss_kib")
}

run_timed build build go build -trimpath -o "$BIN" ./cmd/gooo-lattice
run_timed test test bash -c 'go test -json ./... > "$1"; status=$?; tee /dev/stderr < "$1" >/dev/null; exit "$status"' _ "$WORK/test-events.json"
TESTS_EXECUTED=$(jq -s '[.[] | select(.Action == "pass" and .Test != null)] | length' "$WORK/test-events.json")
TESTS_REUSED=$(jq -s '[.[] | select(.Action == "pass" and .Cached == true)] | length' "$WORK/test-events.json")
TESTS_SKIPPED=$(jq -s '[.[] | select(.Action == "skip" and .Test != null)] | length' "$WORK/test-events.json")

declare -A IR_BY_SOURCE=()
for source in "${SOURCES[@]}"; do
  name=${source##*/}
  name=${name%.gooo}
  ir="$WORK/irs/$name.json"
  run_timed "compile-$name" compile "$BIN" compile -source "$source" -contract "$CONTRACT" -out "$ir"
  IR_BY_SOURCE["$source"]="$ir"
  cp "$ir" "$EVIDENCE/irs/$name.json"
done

for case_id in "${CASES[@]}"; do
  fixture="fixtures/cases/$case_id.json"
  source=$(jq -r '.source' "$fixture")
  ir=${IR_BY_SOURCE[$source]}
  case_dir="$WORK/cases/$case_id"
  mkdir -p "$case_dir/generated"
  run_timed "$case_id-generate" conformance "$BIN" generate -ir "$ir" -fixture "$fixture" -out-dir "$case_dir/generated" -manifest "$case_dir/manifest.json"
  run_timed "$case_id-execute" conformance "$BIN" execute -ir "$ir" -fixture "$fixture" -contract "$CONTRACT" -generated-go "$case_dir/generated/evaluator.go" -out "$case_dir/receipt.json"
  expected=$(jq -c '[.expected.state,.expected.relation,.expected.action]' "$fixture")
  actual=$(jq -c '[.candidate.state,.candidate.relation,.candidate.action]' "$case_dir/receipt.json")
  [ "$actual" = "$expected" ]
  mkdir -p "$EVIDENCE/cases/$case_id/generated"
  cp "$case_dir/manifest.json" "$EVIDENCE/cases/$case_id/manifest.json"
  cp "$case_dir/receipt.json" "$EVIDENCE/cases/$case_id/receipt.json"
  cp "$case_dir/generated/evaluator.go" "$EVIDENCE/cases/$case_id/generated/evaluator.go"
done

jq -S -s '{schema:"gooo/improvement-dominance-lattice/summary/v1",receipts:sort_by(.case_id)}' "$EVIDENCE"/cases/*/receipt.json > "$WORK/summary.json"
cp "$WORK/summary.json" "$EVIDENCE/summary.json"
cp "$CONTRACT" "$EVIDENCE/contract.json"
cp "fixtures/input/gooo-proof-aware-test-reuse-v0.1.2.json" "$EVIDENCE/gooo-proof-aware-test-reuse-v0.1.2.json"

artifact_files=0
artifact_bytes=0
while IFS= read -r -d '' file; do
  artifact_files=$((artifact_files + 1))
  bytes=$(wc -c < "$file" | tr -d '[:space:]')
  artifact_bytes=$((artifact_bytes + bytes))
done < <(find "$EVIDENCE" -type f ! -name 'runtime.json' ! -name 'dossier.json' ! -name 'dossier.md' -print0 | sort -z)

descendant_directories=0
regular_files=0
go_files=0
go_lines=0
gooo_files=0
gooo_lines=0
while IFS= read -r -d '' directory; do descendant_directories=$((descendant_directories + 1)); done < <(find . -mindepth 1 -type d ! -path './.git' ! -path './.git/*' -print0 | sort -z)
while IFS= read -r -d '' file; do
  [ "$file" = "./README.md" ] && continue
  regular_files=$((regular_files + 1))
  lines=$(awk 'END {print NR+0}' "$file")
  case "$file" in
    *.go) go_files=$((go_files + 1)); go_lines=$((go_lines + lines)) ;;
    *.gooo) gooo_files=$((gooo_files + 1)); gooo_lines=$((gooo_lines + lines)) ;;
  esac
done < <(find . -type f ! -path './.git' ! -path './.git/*' -print0 | sort -z)

test -z "$(git status --porcelain)"
jq -S -n \
  --arg schema "gooo/improvement-dominance-lattice/runtime/v1" \
  --argjson compile_wall_ms "${PHASE_WALL[compile]}" --argjson compile_rss_kib "${PHASE_RSS[compile]}" \
  --argjson build_wall_ms "${PHASE_WALL[build]}" --argjson build_rss_kib "${PHASE_RSS[build]}" \
  --argjson test_wall_ms "${PHASE_WALL[test]}" --argjson test_rss_kib "${PHASE_RSS[test]}" \
  --argjson conformance_wall_ms "${PHASE_WALL[conformance]}" --argjson conformance_rss_kib "${PHASE_RSS[conformance]}" \
  --argjson integration_wall_ms "${PHASE_WALL[integration]}" --argjson integration_rss_kib "${PHASE_RSS[integration]}" \
  --argjson tests_executed "$TESTS_EXECUTED" --argjson tests_reused "$TESTS_REUSED" --argjson tests_skipped "$TESTS_SKIPPED" \
  --argjson artifact_files "$artifact_files" --argjson artifact_bytes "$artifact_bytes" \
  --argjson descendant_directories "$descendant_directories" --argjson regular_files "$regular_files" \
  --argjson go_files "$go_files" --argjson go_lines "$go_lines" --argjson gooo_files "$gooo_files" --argjson gooo_lines "$gooo_lines" \
  '{schema:$schema,compile:{wall_ms:$compile_wall_ms,peak_rss_kib:$compile_rss_kib},build:{wall_ms:$build_wall_ms,peak_rss_kib:$build_rss_kib},test:{wall_ms:$test_wall_ms,peak_rss_kib:$test_rss_kib},conformance:{wall_ms:$conformance_wall_ms,peak_rss_kib:$conformance_rss_kib},integration:{wall_ms:$integration_wall_ms,peak_rss_kib:$integration_rss_kib},tests_executed:$tests_executed,tests_reused:$tests_reused,tests_skipped:$tests_skipped,generated_artifact_files:$artifact_files,generated_artifact_bytes:$artifact_bytes,inventory:{descendant_directories:$descendant_directories,regular_files_root_readme_excluded:$regular_files,go_files:$go_files,go_physical_lines:$go_lines,gooo_files:$gooo_files,gooo_physical_lines:$gooo_lines},authority:{runtime_input_reads:0,repository_writes:0,apply:0,commit:0,merge:0,tag:0,release:0,cross_project_required_gates:0},local_validation_commands:["go test ./... (PROHIBITED locally)","go build ./... (PROHIBITED locally)","go vet ./... (PROHIBITED locally)","gofmt (PROHIBITED locally)","lint (PROHIBITED locally)","check (PROHIBITED locally)","conformance (PROHIBITED locally)"],operational_refuted:"OPERATIONAL_REFUTED: runtime input/repository writes/apply/commit/merge/tag/release authority=0; caller-owned output only"}' > "$EVIDENCE/runtime.json"

run_timed integration integration "$BIN" dossier -contract "$CONTRACT" -summary "$WORK/summary.json" -runtime "$EVIDENCE/runtime.json" -out-json "$EVIDENCE/dossier.json" -out-md "$EVIDENCE/dossier.md"
jq -S --argjson wall_ms "${PHASE_WALL[integration]}" --argjson peak_rss_kib "${PHASE_RSS[integration]}" '.integration={wall_ms:$wall_ms,peak_rss_kib:$peak_rss_kib}' "$EVIDENCE/runtime.json" > "$EVIDENCE/runtime.updated.json"
mv "$EVIDENCE/runtime.updated.json" "$EVIDENCE/runtime.json"
"$BIN" dossier -contract "$CONTRACT" -summary "$WORK/summary.json" -runtime "$EVIDENCE/runtime.json" -out-json "$EVIDENCE/dossier.json" -out-md "$EVIDENCE/dossier.md"
jq -e '.decision == "CONFORMANCE_CLOSED" and .case_counts.normal == 3 and .case_counts.unknown == 3 and .case_counts.refuted == 3 and (.runtime.authority | to_entries | all(.value == 0)) and (.receipts | length == 9)' "$EVIDENCE/dossier.json" >/dev/null
test -z "$(git status --porcelain)"

printf 'evidence=%s\n' "$EVIDENCE"
printf 'generated_artifact_files=%s\n' "$artifact_files"
printf 'generated_artifact_bytes=%s\n' "$artifact_bytes"
printf 'tests_executed=%s\n' "$TESTS_EXECUTED"
