# HackArena 2024 backend
---
This repo contains code for backend app of HackArena 2024. You can checkout frontend [here](https://github.com/INIT-SGGW/HackArena2024-frontend). The backend seerver run as a single build binary on backend server.

## Project set up

Open command line and go to directory of your choosing.

### `git clone git@github.com:INIT-SGGW/HackArena2024-backend.git`

Clones respository to local machine.

Install the go from the [official source](https://go.dev/doc/install)

### `go tidy` or `go download`

Restore all packages in project

### `go run hackarena-backend.go`

Runs the app in the development mode.

### `go build hackarena-backend.go`

Builds the app for production, it builds as single binary.


## Database

The project use ORM libary [GORM](https://gorm.io/index.html) as a result the db migration is performd using `repository.SyncDB()` (After connecting to DB), which was declared in `model\model.go` file.

The project use following environmental variables:
```env
HACKDB_USER="DBUserName"
HACKDB_PWD="PassForDBUser"
HA_API_KEY="ApiKeyForStandardRequest"
HA_ADMIN_API_KEY="ApiKeyForAdminRequests"
SECRET_JWT="JWT-secret"
HA_EMAIL_USER="EmailUserForEmailSending"
HA_EMAIL_PWD="PasswordForEmailUser"
HA_EMAIL_HOST="AddresToSendEmails"
HA_EMAIL_PORT="PortToEmailServer"
HA_WEB_URL="AdressToFrontendPage"
HA_EMAIL_TEMP_PATH="PathToEmailTemplatesFolder"
HA_ADMIN_FILE_STORAGE="PathToStoreFilesUploadedByAdmins"
HA_ALL_FILE_STORAGE="PathForTemporalStorageWithAllSolutions"
```

Also server use folowing env file for db detailes:
```env
DB_DRIVER=postgresql
DB_HOST=localhost
DB_PORT=5432
DB_NAME=hackarena
DB_FILE_STORAGE="/absolute/path/to/send/solutions"
```