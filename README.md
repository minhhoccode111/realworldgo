# ![RealWorld Example App](logo.png)

> ### Golang codebase containing real world examples (CRUD, auth, advanced patterns, etc) that adheres to the [RealWorld](https://github.com/gothinkster/realworld) spec and API.

### [Demo](https://demo.realworld.build/)&nbsp;&nbsp;&nbsp;&nbsp;[RealWorld](https://github.com/gothinkster/realworld)

This codebase was created to demonstrate a fully fledged fullstack application built with **Golang** including CRUD operations, authentication, routing, pagination, and more.

We've gone to great lengths to adhere to the **Golang** community styleguides & best practices.

For more information on how to this works with other frontends/backends, head over to the [RealWorld](https://github.com/gothinkster/realworld) repo.

## How it works

> Describe the general architecture of your app here

## Getting started

> docker compose up

## Concepts:

## APIs:

```http
     # Auth
POST   /users
POST   /users/login
GET    /user
PUT    /user

       # Article, Favorite, Comments
POST   /articles
GET    /articles
GET    /articles/feed # prevent article with title 'feed'
GET    /articles?author={username}
GET    /articles?favorited={username}
GET    /articles?tag={tag1,tag2}

GET    /articles/{slug}
PUT    /articles/{slug}
DELETE /articles/{slug}
POST   /articles/{slug}/favorite
DELETE /articles/{slug}/favorite
POST   /articles/{slug}/comments
GET    /articles/{slug}/comments
DELETE /articles/{slug}/comments/{commentId}

       # Profiles
POST   /users/celeb
GET    /profiles/celeb_{username}
POST   /profiles/celeb_{username}/follow
DELETE /profiles/celeb_{username}/follow

       # Tags
GET    /tags
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

- [ ] Add tests

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
