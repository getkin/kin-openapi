#!/bin/bash -eux
set -o pipefail

outdir=.github/docs
mkdir -p "$outdir"
for pkgpath in $(git ls-files | grep  / | while read -r path; do dirname "$path"; done | sort -u | grep -vE '[.]git|testdata|internal|cmd/'); do
	echo $pkgpath
	go doc -all ./"$pkgpath" | tee "$outdir/${pkgpath////_}.txt"
done

git --no-pager diff -- .github/docs/

count_missing_mentions() {
	local errors=0
	for thing in $(git --no-pager diff -- .github/docs/ \
		| grep -vE '[-]{3}' \
		| grep -Eo '^-[^ ]+ ([^ (]+)[ (]' \
		| sed 's%(% %' \
		| cut -d' ' -f2); do
		if ! grep -A999999 '## Sub-v0 breaking API changes' README.md | grep -F "$thing"; then
			((errors++)) || true
		fi
	done
	return $errors
}
count_missing_mentions
