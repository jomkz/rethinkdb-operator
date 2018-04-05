# RethinkDB Operator

A Kubernetes operator to manage RethinkDB instances.

## Usage

TODO

## Development

Build the operator, remember to update the version.

```
operator-sdk build jmckind/rethinkdb-operator:<VERSION>
```

Push the new image, remember to update the version.

```
docker push jmckind/rethinkdb-operator:<VERSION>
```

## Testing

Run the tests.

```
go test ./... -timeout 120s -v -short -cover -coverprofile=tmp/_output/coverage.out
```

Generate the code coverage report.

```
go tool cover -html=tmp/_output/coverage.out
```
