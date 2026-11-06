#!/usr/bin/env bash
# fetch-fuzz-crashers.sh — pull the failing inputs that the `fuzz` workflow
# uploaded (artifacts named crashers-<FuzzTarget>) for runs on master, drop
# each one into <pkg>/testdata/fuzz/<FuzzTarget>/, and keep only those that
# still fail on the current checkout, one per distinct failure.
#
# No gh and no auth needed: artifacts are listed through the public REST API
# (60 requests/hour anonymously; one request per 100 artifacts) and, since
# GitHub refuses anonymous artifact downloads even on public repos, fetched
# through https://nightly.link. If GITHUB_TOKEN is set it is used instead,
# for both listing and downloading, straight from api.github.com.
#
# Inputs that fail the same way are duplicates; only the smallest is kept.
# "The same way" is: a panic or fatal error raised from the same kin-openapi
# function and line, a stack overflow cycling through the same kin-openapi
# functions, a timeout stuck in the same place, or a test failure reported
# from the same line with the same message (numbers and quoted text ignored).
#
# Usage: ./fetch-fuzz-crashers.sh [--days N]
#   --days N    look back N days (default 30; GitHub keeps artifacts 90 days)
#
# Env: REPO (getkin/kin-openapi), TEST_TIMEOUT (60s), GITHUB_TOKEN (optional)
# Requires: bash 4+, curl, jq, unzip, go. Run from the root of a kin-openapi
# checkout. Prints the kept files on stdout, e.g. for `| xargs git add`.

set -euo pipefail

REPO=${REPO:-getkin/kin-openapi}
TEST_TIMEOUT=${TEST_TIMEOUT:-60s}
DAYS=30
MODULE=github.com/getkin/kin-openapi/

log()   { printf '\033[1;34m==>\033[0m %s\n' "$*" >&2; }
warn()  { printf '\033[1;33m!!\033[0m %s\n' "$*" >&2; }
die()   { printf '\033[1;31mxx\033[0m %s\n' "$*" >&2; exit 1; }
cmd() { # echo a command, single-quoting the args that need it
  local a q=
  for a; do
    if [[ $a =~ ^[A-Za-z0-9_./=:,+-]+$ ]]; then q+=" $a"; else q+=" '${a//\'/\'\\\'\'}'"; fi
  done
  printf '\033[2m  $%s\033[0m\n' "$q" >&2
}
crash() { printf '\033[1;31m  CRASH\033[0m \033[1m%s\033[0m\n' "$*" >&2; }
usage() { sed -n '2,/^$/p' "$0" | sed 's/^# \{0,1\}//'; exit "${1:-0}"; }

while (($#)); do
  case $1 in
    --days) [[ ${2:-} =~ ^[0-9]+$ ]] || die "--days needs a number"; DAYS=$2; shift ;;
    -h|--help) usage ;;
    *) warn "unknown argument: $1"; usage 1 ;;
  esac
  shift
done

for bin in curl jq unzip go; do
  command -v "$bin" >/dev/null || die "missing dependency: $bin"
done
grep -q "^module ${MODULE%/}\$" go.mod 2>/dev/null \
  || die "run this from the root of a kin-openapi checkout"

# GNU date, then BSD/macOS date
since=$(date -u -d "-$DAYS days" +%Y-%m-%dT%H:%M:%SZ 2>/dev/null \
     || date -u -v "-${DAYS}d" +%Y-%m-%dT%H:%M:%SZ)

auth=()
[[ -n ${GITHUB_TOKEN:-} ]] && auth=(-H "Authorization: Bearer $GITHUB_TOKEN")
api() { curl -fsSL "${auth[@]}" -H 'Accept: application/vnd.github+json' "https://api.github.com/$1"; }

tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT

# ------------------------------------------------------------ list ----
# Repo-wide artifact list, newest first: stop paging once past $since.
# Keep crashers-* from runs on master of this repo itself (fork PRs from a
# branch that happens to be called master have a different head repo).
log "listing crashers-* artifacts from master since $since"
: >"$tmp/artifacts.tsv"
page=1
while :; do
  json=$(api "repos/$REPO/actions/artifacts?per_page=100&page=$page") \
    || die "GitHub API request failed (anonymous rate limit is 60/hour; set GITHUB_TOKEN?)"
  jq -r --arg since "$since" '.artifacts[]
    | select(.created_at >= $since and (.expired | not)
             and (.name | startswith("crashers-"))
             and .workflow_run.head_branch == "master"
             and .workflow_run.head_repository_id == .workflow_run.repository_id)
    | [.id, .name, .workflow_run.id, .created_at] | @tsv' <<<"$json" >>"$tmp/artifacts.tsv"
  # The order is only roughly by date: page on until a whole page is too old.
  n=$(jq '.artifacts | length' <<<"$json")
  newest=$(jq -r '[.artifacts[].created_at] | max // ""' <<<"$json")
  ((n == 100)) && [[ $newest > $since ]] || break
  ((page++))
done
total=$(wc -l <"$tmp/artifacts.tsv" | tr -d ' ')
((total)) || { log "no crasher artifacts in the last $DAYS days"; exit 0; }
log "$total artifacts"

# ------------------------------------------------------- download ----
download() { # $1 artifact id  $2 output zip
  if [[ -n ${GITHUB_TOKEN:-} ]]; then
    curl -fsSL "${auth[@]}" -o "$2" "https://api.github.com/repos/$REPO/actions/artifacts/$1/zip"
  else
    curl -fsSL -o "$2" "https://nightly.link/$REPO/actions/artifacts/$1.zip"
  fi
}

# Fuzz target -> package dir, from where `func FuzzX(f *testing.F)` lives.
declare -A pkg_of=()
pkg_for() { # $1 target
  if [[ -z ${pkg_of[$1]+x} ]]; then
    local f; f=$(grep -rlE --include='*_test.go' "^func $1\(f \*testing\.F\)" . || true)
    f=${f%%$'\n'*}
    pkg_of[$1]=${f:+$(dirname "${f#./}")}
  fi
  printf '%s' "${pkg_of[$1]}"
}

# Returns 0 = input fails (reproduces), 1 = passes, 2 = could not run.
replay() { # $1 pkg dir  $2 target  $3 corpus file name  $4 logfile
  local rc=0 args=(go test -count=1 -timeout "$TEST_TIMEOUT" -run "^$2\$/^$3\$" -v "./$1")
  cmd "${args[@]}"
  "${args[@]}" >"$4" 2>&1 || rc=$?
  if grep -qE '\[build failed\]|\[setup failed\]|^# ' "$4"; then return 2; fi
  if ((rc == 0)); then
    grep -qE -- "--- PASS: $2/$3( |$)" "$4" && return 1
    return 2 # exit 0 but the entry never ran
  fi
  grep -qE -- '--- FAIL:|^panic:|test timed out|fatal error:' "$4" && return 0
  return 2
}

# kin-openapi frames of goroutine dumps in a log, as "<goroutine#> <func> <file:line>",
# with call arguments and absolute paths dropped.
module_frames() { # $1 logfile
  awk -v mod="$MODULE" '
    /^goroutine [0-9]+ .*\[.*\]:$/ { g++; next }   # "goroutine 7 [running]:", or with gp=… m=… after fatal errors
    g && /^[^\t]/ && index($0, mod) == 1 {
      fn = substr($0, length(mod) + 1); sub(/\([^()]*\)$/, "", fn)
      if ((getline loc) <= 0) exit
      sub(/^\t/, "", loc); sub(/ \+0x.*$/, "", loc); n = split(loc, p, "/")  # also drops fp= sp= pc=
      print g, fn, p[n]
    }' "$1"
}

# One line naming how an input fails, so that the same bug found through
# different inputs (or different targets) collapses to one signature.
signature() { # $1 logfile  $2 target
  local log=$1 s
  if grep -qE '^(runtime: goroutine stack exceeds|fatal error: stack overflow)' "$log"; then
    # The overflow can hit anywhere in the cycle: name the set of functions.
    s=$(module_frames "$log" | awk '$1 == 1 && n++ < 200 { print $2 }' | sort -u | paste -sd, -)
    printf 'stack overflow through %s' "${s:-?}"
  elif grep -qE '^panic: test timed out' "$log"; then
    # Goroutine dumps: first non-test kin-openapi frame found anywhere.
    s=$(module_frames "$log" | awk '$3 !~ /_test\.go:/ { print $2, $3; exit }')
    printf 'timeout in %s' "${s:-?}"
  elif grep -qE '^(panic|fatal error):' "$log"; then
    s=$(module_frames "$log" | awk '$1 == 1 { print $2, $3; exit }')
    printf 'panic in %s' "${s:-?}"
  else
    # t.Error/t.Fatal (or testify): where it was reported, and what it said
    # with the input-dependent bits (numbers, quoted text) blanked out.
    s=$(awk '
      !loc && match($0, /^ +[A-Za-z0-9_]+\.go:[0-9]+: ?/) {
        loc = substr($0, 1, RLENGTH); gsub(/^ +|: ?$/, "", loc)
        msg = substr($0, RLENGTH + 1); next
      }
      loc && msg == "" && match($0, /^[ \t]*Error:[ \t]*/) { msg = substr($0, RLENGTH + 1) }
      END { print loc; print msg }' "$log")
    local loc=${s%%$'\n'*} msg=${s#*$'\n'}
    # shellcheck disable=SC2016 # literal backticks
    msg=$(sed -E 's/"([^"\\]|\\.)*"/"…"/g; s/`[^`]*`/`…`/g; s/[0-9]+/N/g' <<<"$msg")
    printf '%s fails at %s: %s' "$2" "${loc:-?}" "${msg:0:160}"
  fi
}

in_git=0; git rev-parse --verify -q HEAD >/dev/null 2>&1 && in_git=1
declare -A seen=() keep_of=() size_of=()
dropped=0 dupes=0 skipped=0
while IFS=$'\t' read -r id name run created; do
  log "$name (run $run, $created)"
  d=$tmp/$id; mkdir -p "$d/x"
  if ! download "$id" "$d/a.zip"; then
    warn "  download failed: https://github.com/$REPO/actions/runs/$run"; continue
  fi
  unzip -qo "$d/a.zip" -d "$d/x" || { warn "  bad zip"; continue; }

  # The artifact is every testdata/fuzz/** file in the CI workspace, so it
  # also carries the committed seed corpus: only files new to us matter.
  while IFS= read -r -d '' f; do
    base=${f##*/}; target=${f%/*}; target=${target##*/}
    [[ $base =~ ^[A-Za-z0-9._-]+$ && $target =~ ^Fuzz[A-Za-z0-9_]*$ ]] || continue
    IFS= read -r first <"$f" || true
    [[ $first == 'go test fuzz v1' ]] || continue
    [[ -z ${seen[$target/$base]+x} ]] || continue
    seen[$target/$base]=1
    pkg=$(pkg_for "$target")
    [[ -n $pkg ]] || { warn "  $target/$base: no such fuzz target in this checkout"; ((++skipped)); continue; }
    dest=$pkg/testdata/fuzz/$target/$base
    # Already here: committed (or no git) means seed corpus, skip it. Present
    # but not in HEAD means an earlier run of this script kept it: replay it
    # too, so that new inputs are deduplicated against it.
    existed=0
    if [[ -e $dest ]]; then
      if ((!in_git)) || git cat-file -e "HEAD:$dest" 2>/dev/null; then continue; fi
      existed=1
    else
      mkdir -p "${dest%/*}"; cp "$f" "$dest"
    fi
    rc=0; replay "$pkg" "$target" "$base" "$d/test.log" || rc=$?
    # never delete what was there before this run unless it is a duplicate
    if ((existed && rc != 0)); then
      log "  already here, left alone: $dest"; continue
    fi
    case $rc in
      0) sig=$(signature "$d/test.log" "$target")
         size=$(wc -c <"$dest" | tr -d ' ')
         if [[ -z ${keep_of[$sig]+x} ]]; then
           crash "$dest  ($size bytes)"; printf '        %s\n' "$sig" >&2
           keep_of[$sig]=$dest; size_of[$sig]=$size
         elif ((size < size_of[$sig])); then
           log "  same failure as ${keep_of[$sig]}, but smaller: keeping $dest instead"
           crash "$dest  ($size bytes)"; printf '        %s\n' "$sig" >&2
           rm -f "${keep_of[$sig]}"; rmdir -p "${keep_of[$sig]%/*}" 2>/dev/null || true
           keep_of[$sig]=$dest; size_of[$sig]=$size; ((++dupes))
         else
           log "  same failure as ${keep_of[$sig]}, dropped: $dest"
           rm -f "$dest"; ((++dupes))
         fi ;;
      1) log "  passes now, dropped: $dest"; rm -f "$dest"; ((++dropped)) ;;
      *) warn "  could not replay $dest:"; tail -n 15 "$d/test.log" | sed 's/^/     /' >&2
         rm -f "$dest"; ((++skipped)) ;;
    esac
    rmdir -p "${dest%/*}" 2>/dev/null || true # if we just created it empty
  done < <(find "$d/x" -type f -path '*testdata/fuzz/*' -print0)
done <"$tmp/artifacts.tsv"

echo >&2
log "${#keep_of[@]} distinct reproducing crashers ($dupes duplicates, $dropped no longer failing, $skipped skipped)"
for sig in "${!keep_of[@]}"; do
  printf '\033[1;31m%s\033[0m\t%s\n' "${keep_of[$sig]}" "$sig" >&2
done
((${#keep_of[@]})) && printf '%s\n' "${keep_of[@]}"
exit 0
