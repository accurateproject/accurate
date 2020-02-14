
#
# Sample db and users creation. Replace here with your own details
#

sudo -u postgres dropdb -e accurate
sudo -u postgres dropuser -e accurate
sudo -u postgres psql  -c "CREATE USER accurate password 'accuRate';"
sudo -u postgres createdb -e -O accurate accurate
