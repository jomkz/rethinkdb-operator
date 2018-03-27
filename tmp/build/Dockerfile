FROM alpine:3.6

ADD tmp/_output/bin/rethinkdb-operator /usr/local/bin/rethinkdb-operator

RUN adduser -D rethinkdb-operator
USER rethinkdb-operator
