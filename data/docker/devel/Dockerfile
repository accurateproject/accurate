FROM golang:alpine
MAINTAINER Radu Fericean, radu@fericean.ro

# install dependencies
RUN apk add --update git wget bash zsh ngrep curl nano

# add accurate user
RUN adduser -h /var/run/accurate -s /bin/false -S accurate

#install glide
RUN go get github.com/Masterminds/glide

#install oh-my-zsh
RUN TERM=xterm sh -c "$(wget https://raw.github.com/robbyrussell/oh-my-zsh/master/tools/install.sh -O -)"; exit 0

# cleanup
RUN rm -f /tmp/* /etc/apk/cache/*

# expose ports
EXPOSE 2012 2013 2080

WORKDIR /go/src/github.com/accurateproject/accurate

# set start command
CMD ./build.sh; cc-engine -cfgdir data/conf/samples/phoenix; sleep 365d
