#!/bin/bash

aws rds describe-db-snapshots --snapshot-type public --include-public --query 'DBSnapshots[?AllocatedStorage==`200`]' --output json | jq '.[].DBSnapshotIdentifier' | tr -d '"'
