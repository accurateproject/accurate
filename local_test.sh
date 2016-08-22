#! /usr/bin/env sh
./test.sh
gen=$?
echo 'go test github.com/accurateproject/accurate/apier/v1 -local'
go test github.com/accurateproject/accurate/apier/v1 -local
ap1=$?
echo 'go test github.com/accurateproject/accurate/apier/v2 -local'
go test github.com/accurateproject/accurate/apier/v2 -local
ap2=$?
echo 'go test github.com/accurateproject/accurate/apier/v2 -tp -config_dir=tutmysql'
go test github.com/accurateproject/accurate/apier/v2 -tp -config_dir=tutmysql
tpmysql=$?
echo 'go test github.com/accurateproject/accurate/apier/v2 -tp -config_dir=tutpostgres'
go test github.com/accurateproject/accurate/apier/v2 -tp -config_dir=tutpostgres
tppg=$?
echo 'go test github.com/accurateproject/accurate/apier/v2 -tp -config_dir=tutmongo'
go test github.com/accurateproject/accurate/apier/v2 -tp -config_dir=tutmongo
tpmongo=$?
echo 'go test github.com/accurateproject/accurate/engine -local -integration'
go test github.com/accurateproject/accurate/engine -local -integration
en=$?
echo 'go test github.com/accurateproject/accurate/cdrc -local'
go test github.com/accurateproject/accurate/cdrc -local
cdrc=$?
echo 'go test github.com/accurateproject/accurate/config -local'
go test github.com/accurateproject/accurate/config -local
cfg=$?
echo 'go test github.com/accurateproject/accurate/utils -local'
go test github.com/accurateproject/accurate/utils -local
utl=$?
echo 'go test github.com/accurateproject/accurate/general_tests -local -integration'
go test github.com/accurateproject/accurate/general_tests -local -integration
gnr=$?
echo 'go test github.com/accurateproject/accurate/agents -integration'
go test github.com/accurateproject/accurate/agents -integration
agts=$?
echo 'go test github.com/accurateproject/accurate/sessionmanager -integration'
go test github.com/accurateproject/accurate/sessionmanager -integration
smg=$?

exit $gen && $ap1 && $ap2 && $tpmysql && $tppg && $tpmongo && $en && $cdrc && $cfg && $utl && $gnr && $agts && $smg
