#! /usr/bin/env sh

echo "Building accuRate..."

go install github.com/accurateproject/accurate/cmd/cgr-engine
cr=$?
go install github.com/accurateproject/accurate/cmd/cgr-loader
cl=$?
go install github.com/accurateproject/accurate/cmd/cgr-console
cc=$?
go install github.com/accurateproject/accurate/cmd/cgr-tester
ct=$?

exit $cr || $cl || $cc || $ct
