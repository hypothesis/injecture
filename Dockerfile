FROM gliderlabs/alpine:3.6
MAINTAINER Hypothes.is Project and contributors

ENV GOPATH="/go" PATH="/go/bin:${PATH}" PORT=8080

# Install system build and runtime dependencies.
RUN apk-install ca-certificates collectd collectd-disk supervisor

# Create the injecture user, group, and home directory.
RUN addgroup -S injecture \
  && adduser -S -G injecture -h /go/src/github.com/hypothesis/injecture injecture
WORKDIR /go/src/github.com/hypothesis/injecture

# Copy packaging.
COPY bin ./bin
COPY conf/supervisord.conf .

# Copy collectd config
COPY conf/collectd.conf /etc/collectd/collectd.conf
RUN mkdir /etc/collectd/collectd.conf.d \
  && chown injecture:injecture /etc/collectd/collectd.conf.d

# collectd always tries to write to this immediately after enabling the logfile plugin.
# Even though we later configure it to write to stdout. So we do have to make sure it's
# writeable.
RUN touch /var/log/collectd.log && chown injecture:injecture /var/log/collectd.log

# Copy application.
COPY static ./static
COPY injecture ./injecture
COPY serve.go Makefile Gopkg.lock Gopkg.toml ./

# Install build deps, build, and then clean up.
RUN apk-install --virtual build-deps git go make musl-dev \
  && go get -u github.com/golang/dep/... \
  && dep ensure \
  && make build \
  && apk del build-deps

EXPOSE "${PORT}"
USER injecture
CMD ["supervisord", "-c", "supervisord.conf", "-j", "/tmp/supervisord.pid"]
