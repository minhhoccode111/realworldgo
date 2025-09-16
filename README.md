# Go Web Authentication and Authorization

I built this solution from scratch
[here](https://github.com/minhhoccode111/go-practice/tree/main/web/auth) as a
foundation for my future projects.

**NOTE**: The code is rough, buggy, and not production-ready—use at your own
risk.

## Concepts:

- Router
- Filter
- Interface
- Middleware
- Pagination
- Soft-delete
- Authorization
- Authentication
- Use `PATCH` verb
- Database indexes
- Password hashing
- Database migration
- Input validation + sanitization
- Context for database query and middleware
- Concurrency for independent database query (get all users, count users)
- Use pointer for boolean to differentiate between "not provided" and
  "explicitly false" in Go

## Features:

- Basic authentication and authorization
- CRUD operations

## APIs:

```txt
       # public
POST   /auth/register
POST   /auth/login
GET    /users/all
       - filter email, status
       - pagination
       - TODO: only admin can get all users including inactive
GET    /users/{id}

       # auth
PATCH  /users/{id}
       - user update their profile, like email, bio, etc.
PATCH  /users/{id}/password
       - user update their password, must provide old password
PATCH  /users/{id}/status
       - user deactivate her account (soft-delete)
       - only admin can activate an account

       # authz
DELETE /users/{id}
       - admin hard-delete an account
       # TODO
PATCH  /users/{id}/role
       - admin change other user's role to admin
```

## Database:

```txt
- Users:
       - id
       - role
idx    - email     - unique
       - password
       - is_active
```

## Todo

- [ ] Add `PATCH  /users/{id}/role`: admin change other user's role to admin
- [ ] Add unit tests

## MakeFile

Run build make command with tests

```bash
make all
```

Build the application

```bash
make build
```

Run the application

```bash
make run
```

Create DB container

```bash
make docker-run
```

Shutdown DB Container

```bash
make docker-down
```

DB Integrations Test:

```bash
make itest
```

Live reload the application:

```bash
make watch
```

Run the test suite:

```bash
make test
```

Clean up binary from the last build:

```bash
make clean
```
