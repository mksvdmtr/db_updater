## DB Updater

#### Как иппользовать:
```
по примеру config.yml.example
прописать пути и доступы

make
./db_updater -c config.yml

примеры конфигурационных файлов которые читает db_updater:

mysql:

staging:
  adapter: mysql2
  encoding: utf8
  database: 'db'
  pool: 5
  username: 'user'
  password: 'pass'


postgresql:

staging:
  adapter: postgresql
  encoding: unicode
  database: 'db'
  pool: 5
  username: 'user'
  password: 'pass'
  host: "localhost"
  port: 5432
 
```