FROM debian:trixie AS builder

RUN apt-get update && \
    apt-get install -y --no-install-recommends golang ca-certificates git python3 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Generate users.json
RUN cd example/idp2 && python3 create_users.py

# Build the binary
RUN cd example/idp2 && go build -o /idp2 .

FROM debian:trixie

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /idp2 /usr/bin/idp2
COPY --from=builder /build/example/idp2/users.json /etc/idp2/users.json

RUN mkdir -p /data
VOLUME /data
EXPOSE 8080
WORKDIR /data
ENTRYPOINT ["/usr/bin/idp2"]
