DELETE FROM mysql.user ;
CREATE USER 'ironic'@'localhost' identified by '$(MARIADB_PASSWORD)' ;
GRANT ALL on *.* TO 'ironic'@'localhost' WITH GRANT OPTION ;
DROP DATABASE IF EXISTS test ;
CREATE DATABASE IF NOT EXISTS  ironic ;
FLUSH PRIVILEGES ;

