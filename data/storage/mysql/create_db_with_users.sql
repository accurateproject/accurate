
--
-- Sample db and users creation. Replace here with your own details
--

DROP DATABASE IF EXISTS accurate;
CREATE DATABASE accurate;

GRANT ALL on accurate.* TO 'accurate'@'localhost' IDENTIFIED BY 'accuRate';
