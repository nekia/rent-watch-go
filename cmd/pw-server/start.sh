#!/bin/bash

cat << EOF > pw-config.json
{
  "port": $1
}
EOF

npx playwright launch-server --browser chromium --config ./pw-config.json