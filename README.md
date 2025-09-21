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

## [Endpoints](https://docs.realworld.show/specifications/backend/endpoints/)

```http
     # Auth
POST   /users
POST   /users/login
GET    /user
PUT    /user

       # Article, Favorite, Comments
POST   /articles
GET    /articles
                ?tag={tag1,tag2}
                &author={username}
                &favorited={username}
                &limit={limit}
                &offset={offset}
GET    /articles/feed
GET    /articles/{slug}
PUT    /articles/{slug}
DELETE /articles/{slug}
POST   /articles/{slug}/favorite
DELETE /articles/{slug}/favorite
POST   /articles/{slug}/comments
GET    /articles/{slug}/comments
DELETE /articles/{slug}/comments/{commentId}

       # Profiles
GET    /profiles/{username}
POST   /profiles/{username}/follow
DELETE /profiles/{username}/follow

       # Tags
GET    /tags
```

## [API Response Format](https://docs.realworld.show/specifications/backend/api-response-format/)

## Database (NoSQL features are not allowed, even though PostgreSQL has them)

```txt
- Users
       - id
idx    - email     - unique
idx    - username  - unique
       - password
       - image
       - bio
       - updated_at
       - created_at
- Articles
       - id
       - author_id
idx    - slug      - unique
       - body
       - title
       - description
       - updated_at
       - created_at
- Comments
       - id
       - articleid
       - authorid
       - body
       - created_at
- Tags
       - id
idx    - name      - unique
       - created_at
- Favorites
       - userid
       - articleid
       - created_at
- Follows
       - followerid
       - followingid
       - created_at
- ArticleTags
       - articleid
       - tagid
```

## Concepts

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

DB Integrations Test

```bash
make itest
```

Live reload the application

```bash
make watch
```

Run the test suite

```bash
make test
```

Clean up binary from the last build

```bash
make clean
```
