#!/usr/bin/env sh
./build.sh

docker run -d -p 27017:27017  --name db mongo
docker run --rm -p 2012:2012 -p 2013:2013 -p 2080:2080 --link db:mongo -itv `pwd`:/go/src/github.com/accurateproject/accurate --name cc accurate
#docker-compose run --rm accurate
