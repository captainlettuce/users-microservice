# Users microservice

`users` microservice is responsible for handling CRUD actions on users

### Quickstart

```shell
# Run the app, the grpc-service listens on port 8000 by default
docker compose up -d

# Generate files
docker compose -f compose.dev.yaml run users go generate

# Run tests (needs generated files)
docker compose -f compose.dev.yaml run users go test ./...

# Run linter (needs generated files)
docker compose -f compose.dev.yaml run users golangci-lint run -v ./...

# Run dev environment (with debugger and hot-reload)
docker compose -f compose.dev.yaml up -d
```

### Api

The api exposes the following functions:

- **users.v1** (see `proto/api.proto` for spec)
    - **add** - Add a new user, returns error if the userId should exist
    - **update** - Update an existing user, returns error if the user does not exist
    - **delete** - Remove a user based on user Id
    - **list** - List filtered, paginated, users
    - **subscribe** - Subscribe to user change events. Optionally specify (non-nil) userId or change type to listen
      for `create, update, delete`
- **grpc.health.v1** according to [spec](https://github.com/grpc/grpc/blob/master/doc/health-checking.md)

### Settings

All app settings are set through environment variables

| Env              | Type                     | Default                   | Description                                           |
|------------------|--------------------------|---------------------------|-------------------------------------------------------|
| DEBUG            | boolean                  | false                     | Toggle debug output                                   |
| LOG_FORMAT       | text \| json             | json                      | Format for log output                                 |
| MONGO_URI        | string                   | mongodb://localhost:27017 | mongodb connection uri to use                         |
| MONGO_DB         | string                   | users                     | mongodb database to use                               |
| MONGO_COLLECTION | string                   | users                     | mongo collection to use                               |
| SHUTDOWN_GRACE   | positive integer         | 5                         | Seconds to wait before forcefully terminating on exit |
| NATS_URI         | string                   | nats://nats:4222          | connection uri for nats                               |
| GRPC_PORT        | positive integer 1-65535 | 8000                      | port to bind grpc server to                           |

### Project structure

- **cmd** - application setup
    - **users** - server application
- **dev** - config files for local development
- **docker**
- **generated** - generated go-code from protobuf-files
- **internal** - internal application logic and types
    - **domain** - domain-layer, business rules and flows
    - **mocks** - generated mocks
    - **pubsub** - service for communicating over pubsub (nats in this case)
    - **repository** - database repository
    - **server** - transport-layer, in this case grpc
        - **service** - the service implementing the grpc server
    - **types** - types
    - **internal.go** - application context and layer-interfaces
- **pkg/utils** - helper utils not part of core domain-logic in any way
- **proto** - protobuf files
- **scripts** - helper scripts for building and running the application

### Stuff intentionally skipped as out of scope

- **tracing/metrics** - Setup is usually very dependent on the infrastructure setup ime, also it takes a lot of
  boilerplate
  coding that should (imo) be abstracted away to some internal library to ensure consistency in the telemetry produced
  across the application (in the "cluster-of-microservices"-sense)
- **Optimization** - There has been minimal optimization (only index in db is on `user.email` for example), this is
  because I didn't want to practice premature optimization and instead focused on clean readable code. Data-driven
  optimizations can be done with more profiling of live, real-world, usage of the system
- **Input validation** - The app does not perform any input validation except validating that the data will not break
  the app. For example a userId should always be checked for nil, but there is no minimum-length-requirement on a users
  nickname or any kind of validation of email addresses, this is considered business-rules and are deemed out-of-scope
  as they need a lot of thought and clarification on how the responsibilities are divided amongst the microservices
- **Passwords** - It is assumed that the passwords are hashed upstream (along with being validated) and that
  authentication/authorization is handled elsewhere for all callers of the api, i.e. anyone requesting the password-hash
  can do so
- **Errors** - It is assumed that all callers are good actors, so all errors are very honest. This might be a problem in
  a prod-environment where you may not want to leak internal errors.
