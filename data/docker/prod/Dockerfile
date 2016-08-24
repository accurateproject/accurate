FROM debian:latest
MAINTAINER Radu Fericean, radu@fericean.ro
RUN apt-get -y update

# set mysql password
RUN echo 'mysql-server mysql-server/root_password password accuRate' | debconf-set-selections && echo 'mysql-server mysql-server/root_password_again password accuRate' | debconf-set-selections

# add freeswitch gpg key
RUN gpg --keyserver pool.sks-keyservers.net --recv-key D76EDC7725E010CF && gpg -a --export D76EDC7725E010CF | apt-key add -

# add freeswitch apt repo
RUN echo 'deb http://files.freeswitch.org/repo/deb/debian/ jessie main' > /etc/apt/sources.list.d/freeswitch.list

# install dependencies
RUN apt-get update && apt-get -y install redis-server mysql-server python-pycurl python-mysqldb postgresql postgresql-client sudo wget git freeswitch-meta-vanilla

# add accurate apt-key
RUN wget -qO- http://apt.itsyscom.com/conf/accurate.gpg.key | apt-key add -

# add accurate repo
RUN cd /etc/apt/sources.list.d/; wget -q http://apt.itsyscom.com/conf/accurate.apt.list

# install accurate
RUN apt-get update && apt-get -y install accurate

# cleanup
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# set start command
CMD /root/code/data/docker/prod/start.sh
