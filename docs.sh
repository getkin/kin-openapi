#!/bin/bash -eux
set -o pipefail

outdir=.github/docs
mkdir -p "$outdir"
for pkgpath in $(git ls-files | grep  / | while read -r path; do dirname "$path"; done | sort -u | grep -vE '[.]git|testdata|cmd/'); do
	go doc -short "./$pkgpath" | tee "$outdir/${pkgpath////_}.txt"
done
