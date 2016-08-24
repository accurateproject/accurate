/etc/init.d/mysql start
/etc/init.d/postgresql start
/etc/init.d/redis-server start

cd /usr/share/accurate/storage/mysql && ./setup_cgr_db.sh root accuRate
cd /usr/share/accurate/storage/postgres && ./setup_cgr_db.sh

/usr/share/accurate/tutorials/fs_evsock/freeswitch/etc/init.d/freeswitch start
/usr/share/accurate/tutorials/fs_evsock/cgrates/etc/init.d/cgrates start

bash --rcfile /root/.bashrc
