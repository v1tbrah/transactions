# transactions

## Installation

`git clone https://github.com/v1tbrah/transactions`

## Getting started

### Prepairing

* You need PostgreSQL installed. You can check it: `psql --version`. If it's not installed, visit https://www.postgresql.org/download/

### Options
The following options are set by default:
```
api server run address: `:5555`
gin mode: `release`
log level: `info`
```
* flag options:
```
   -a string
      api server run address
   -p string
      connection string to postgres db
   -g string
      gin mode
   -l string
      log level 
```
For example: `go run cmd/main.go -a=:5555 -d="host=localhost port=5432 user=postgres password=12345678 dbname=transactions sslmode=disable"`
* env options you can check in internal/config/parse

### Note!

* You definitely need to configure the db connection string

### Starting

* Open first terminal. Go to project working directory. Run "transactions" app. For example:
   ```
   cd ~/go/src/transactions
   go run cmd/main.go -a=:5555 -d="host=localhost port=5432 user=postgres password=12345678 dbname=transactions sslmode=disable"
   ```

* Because this is a training task:
  * You have just 5 users with id [1, 2, 3, 4, 5]
  * For receipt money you can do `POST RUN_API_ADDRESS/{user_id}/receipt/{sum}`
    * For example http://localhost:5555/1/receipt/1
    * You can find more examples in project working directory /http
  * For withdraw money you can do `POST RUN_API_ADDRESS/{user_id}/withdraw/{sum}`
      * For example http://localhost:5555/1/withdraw/1
      * You can find more examples in project working directory /http