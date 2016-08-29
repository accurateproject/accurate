FROM debian:latest
MAINTAINER Radu Fericean, radu@fericean.ro

# add mongo repo keys
RUN apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv EA312927

# add mongo repo
RUN echo 'deb http://repo.mongodb.org/apt/debian wheezy/mongodb-org/3.2 main' | tee '/etc/apt/sources.list.d/mongodb-org-3.2.list'

# install dependencies
RUN apt-get -y update && apt-get -y install git redis-server python-pycurl sudo wget vim zsh mongodb-org tmux rsyslog ngrep curl

# add mongo conf
COPY mongod.conf /etc/mongod.conf

# add accurate user
RUN useradd -c accuRate -d /var/run/accurate -s /bin/false -r accurate

# install golang
RUN wget -qO- https://storage.googleapis.com/golang/go1.7.linux-amd64.tar.gz | tar xzf - -C /root/

#install glide
RUN GOROOT=/root/go GOPATH=/root/code /root/go/bin/go get github.com/Masterminds/glide

#install oh-my-zsh
RUN TERM=xterm sh -c "$(wget https://raw.github.com/robbyrussell/oh-my-zsh/master/tools/install.sh -O -)"; exit 0

# change shell for tmux
RUN chsh -s /usr/bin/zsh

# cleanup
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# set start command
CMD /root/code/src/github.com/accurateproject/accurate/data/docker/devel/start.sh
