#!/bin/bash
set -e

echo "generating mocks next to their source packages..."

failed=0

while read -r file; do
    if ! grep -q "interface {" "$file"; then
        continue
    fi

    dir="$(dirname "$file")"
    pkg="$(basename "$dir")"
    output_file="$dir/$(basename "${file%.go}")_mock_test.go"

    if err="$(mockgen \
        -source="$file" \
        -destination="$output_file" \
        -package="$pkg" 2>&1 < /dev/null)"; then
        echo "SUCCESS: $file → $output_file"
    else
        echo "ERROR: failed to generate mock for $file:"
        echo "$err" | grep -v "Unable to infer -self_package" | sed 's/^/    /'
        failed=1
    fi
done < <(find internal -name "*.go" -not -path "*/vendor/*" -not -name "*_test.go")

if [ "$failed" -ne 0 ]; then
    echo "some mocks failed to generate"
    exit 1
fi

echo "all mocks generated next to their sources"
