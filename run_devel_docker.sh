#!/usr/bin/env sh

docker run --rm -p 2012:2012 -p 2013:2013 -p 2080:2080 -p 27017:27017 -itv `pwd`:/root/code/src/github.com/accurateproject/accurate --name cc accurate
