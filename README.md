Nebulo
======
### What Is It?
This is a web app written in Go. It uses goquery to scrape data from http://aqicn.org/ and stores it in a PostgreSQL database which can be accessed by the API layer to return a simple JSON response.

Will love to gather some feedback on how to improve this, so feel free to submit an issue and I'll be happy to learn!

### Requirements
PostgreSQL database

#### Setting up PostgreSQL
(Instructions for Ubuntu)

1. Install postgresql

  ```bash
  sudo apt-get install postgresql postgresql-contrib
  ```
2. Log into the PostgreSQL administrative user

  ```shell
  sudo -i -u postgres
  ```
3. Create a new user called 'nebulo'

  ```shell
  createuser nebulo --superuser
  ```
4. Enter PostgreSQL and add a database for the new user

  ```shell
  psql
  createdb nebulo
  ```
  *(Ctrl+D to exit the PostgreSQL prompt)*

5. Log into the new user and launch PostgreSQL

  ```shell
  sudo -i -u nebulo
  psql
  ```

6. Run these SQL creation statements

  ```sql
  CREATE TABLE data (
      id integer NOT NULL,
      city_name character varying(100) NOT NULL,
      data integer,
      temp integer,
      advisory integer,
      "timestamp" bigint
  );

  CREATE TABLE locations (
      city_name character varying(100) NOT NULL,
      lat double precision,
      lng double precision
  );

  CREATE TABLE push_android (
      uuid character varying(1000) NOT NULL,
      id integer
  );

  CREATE TABLE push_ios (
      uuid character varying(100) NOT NULL,
      id integer
  );

  ```

7. Exit PostgreSQL and proceed to the Getting Started section

  *(Ctrl+D to exit the PostgreSQL prompt)*

### Getting Started
Clone the repository:

```bash
git clone git@github.com:jurvis/nebulo-backend.git
```

cd into your project directory and run the following to set up your GOPATH:

```bash
export GOPATH=$(pwd)
```

create and fill up these configuration files:
- `dbconfig.gcfg`
- `emailconfig.gcfg`
- `newrelic.gcfg`
- `pushconfig.gcfg`

To run:
` go run monitor.go `
