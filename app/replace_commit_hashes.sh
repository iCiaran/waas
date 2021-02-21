#!/bin/sh

# Get the values from the files created in the first stage of the build
short=$(cat commit_hash_short)
long=$(cat commit_hash_long)

# Replace the placeholders
sed -i ./static/index.html -e "s/__SHORT_COMMIT__/$short/g"
sed -i ./static/index.html -e "s/__LONG_COMMIT__/$long/g"
