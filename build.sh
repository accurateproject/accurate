#! /usr/bin/env sh

echo "Building accuRate..."

go install github.com/accurateproject/accurate/cmd/cc-engine
cr=$?
go install github.com/accurateproject/accurate/cmd/cc-loader
cl=$?
go install github.com/accurateproject/accurate/cmd/cc-console
cc=$?
go install github.com/accurateproject/accurate/cmd/cc-tester
ct=$?

exit $cr || $cl || $cc || $ct
