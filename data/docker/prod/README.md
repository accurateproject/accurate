Docker
=======

From the project root:

``` bash
# build the image
docker build -t accurate data/docker/prod
# create the container
docker run --rm -itv `pwd`:/root/code --name cgr accurate
```
