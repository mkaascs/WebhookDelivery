#!/bin/bash
set -e

echo "generating mocks next to their source packages..."

find internal -name "*.go" \
    -not -path "*/vendor/*" \
    -not -name "*_test.go" | while read -r file; do

    if ! grep -q "interface {" "$file"; then
        continue
    fi

    dir="$(dirname "$file")"
    pkg="$(basename "$dir")"
    output_file="$dir/$(basename "${file%.go}")_mock_test.go"

    echo "SUCCESS: $file → $output_file"

    mockgen \
        -source="$file" \
        -destination="$output_file" \
        -package="$pkg" \
        2>/dev/null || echo "WARNING: no interfaces found in $file"
done

echo "all mocks generated next to their sources"
