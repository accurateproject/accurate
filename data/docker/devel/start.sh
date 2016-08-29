/etc/init.d/rsyslog start
/etc/init.d/redis-server start
/etc/init.d/mongod start

# create a link to data dir
ln -s /root/code/src/github.com/accurateproject/accurate/data /usr/share/accurate
# create link to accurate dir
ln -s /root/code/src/github.com/accurateproject/accurate /root/cc

# create accurate user for mongo
mongo --eval 'db.createUser({"user":"accurate", "pwd":"accuRate", "roles":[{role: "userAdminAnyDatabase", db: "admin"}]})' admin

#env vars
export GOROOT=/root/go; export GOPATH=/root/code; export PATH=$GOROOT/bin:$GOPATH/bin:$PATH

# build and install accurate
cd /root/cc
#glide -y devel.yaml install
./build.sh

# create cgr-engine and cgr-loader link
ln -s /root/code/bin/cgr-engine /usr/bin/
ln -s /root/code/bin/cgr-loader /usr/bin/

# expand freeswitch conf
#cd /usr/share/accurate/tutorials/fs_evsock/freeswitch/etc/ && tar xzf freeswitch_conf.tar.gz

#cd /root/.oh-my-zsh; git pull

cd /root/cc
echo "for cgradmin run: cgr-engine -config_dir data/conf/samples/cgradmin"
echo 'export GOROOT=/root/go; export GOPATH=/root/code; export PATH=$GOROOT/bin:$GOPATH/bin:$PATH'>>/root/.zshrc

DISABLE_AUTO_UPDATE="true" zsh
