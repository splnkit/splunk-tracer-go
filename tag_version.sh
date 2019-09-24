#!/bin/sh

VERSION=$(cat ./VERSION)

cat > version.go <<EOF
package splunktracing

// TracerVersionValue provides the current version of the splunk-tracer-go release
const TracerVersionValue = "$VERSION"
EOF
