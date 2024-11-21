# learning-url-shortener-with-go

This URL shortener is a simple application that allows you to shorten long URLs into shorter ones. It is entirely built with Go and uses PostgreSQL as the database. Any concurrency issues are handled and unique short URL generation is ensured with collision detection.

---

## Technologies Used:

- **Backend:** Go
- **Frontend:** HTML, Javascript, CSS
- **Database:** PostgreSQL

---

## Requirements:

- Go 1.23
- PostgreSQL 16.3

---

## Installation

```bash
git clone https://github.com/tberk-s/learning-url-shortener-with-go.git
```

## Setup Development:

| Variable Name | Description | Default Value |
|:---------------------|:------------|:--------------|
| `SERVER_ENV` | Server Environment | `` |
| `DB_HOST` | Database Host | `` |
| `DB_PORT` | Database Port | `` |
| `DB_NAME` | Database Name | `` |
| `DB_USER` | Database User Name | `` |
| `DB_PASSWORD` | Database Password | `` |

Create your own `.env` file and set the variables.

### Create Database:

```bash
cd /path/to/url-shortener/
source .env
# Create a PostgreSQL user (replace 'myuser' and 'mypassword' with your username and password)
psql -c "CREATE USER myuser WITH PASSWORD 'mypassword';"
dropdb --if-exists "${DATABASE_NAME}"
createdb "${DATABASE_NAME}" -O "myuser"
```

---

## Usage:

1. Run the server:
    - `go run cmd/web-app/main.go`
2. Open your browser and navigate to `http://localhost:8000`

3. Shorten a URL

4. Redirect to the original URL
