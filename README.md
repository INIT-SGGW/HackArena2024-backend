# HackArena 2024 backend
---
This repo contains code for frontend app of HackArena 2024. You can checkout frontend [here](https://github.com/INIT-SGGW/HackArena2024-frontend).

## Project set up

Open command line and go to directory of your choosing.

### `git clone git@github.com:INIT-SGGW/HackArena2024-backend.git`

Clones respository to local machine.

Install the go from the [official source](https://go.dev/doc/install)

### `go tidy`

Restore all packages in project

### `go run hackarena-backend.go`

Runs the app in the development mode.

### `go build hackarena-backend.go`

Builds the app for production, it builds as single binary.


## Database

The project use ORM libary [GORM](https://gorm.io/index.html) as a result the db migration is performd using `repository.SyncDB()` (After connecting to DB)

Also project use the following environmental variable for connection and secret handling:

- HACKDB_USER - user in database
- HACKDB_PWD - password for database user
- HA_API_KEY - the key using in authentication header (`AuthMiddleweare()` method)
- SECRET_JWT - secret for generating JWT token 

Other connection details like: db driver, port, database name and host are define in constant in repository.