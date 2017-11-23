FROM golang:1.9 as builder
ENV REPO=/go/src/github.com/yuuki0xff/temvote/
# build a executable file
COPY *.go $REPO/
RUN cd $REPO && go get && go build
RUN mv $REPO/temvote /srv/
# create database initialization sql
COPY db.sqlite3.sql tables.sql /srv/
RUN cat /srv/*.sql >/srv/init.sql

FROM debian:stable-slim
VOLUME /srv/data/
RUN apt-get update && apt-get install -y ca-certificates && apt-get clean
COPY static/ /srv/static/
COPY template/ /srv/template/
COPY --from=builder /srv/temvote /srv/init.sql /srv/
CMD ["/srv/temvote"]
