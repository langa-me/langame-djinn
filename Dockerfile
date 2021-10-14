FROM golang as build

WORKDIR /all

COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo \
    -o /go/bin/server \
    cmd/djinn/main.go

FROM scratch

COPY --from=build /go/bin/server /server
COPY --from=build /all/config.yaml /config.yaml
COPY --from=build /all/svc.dev.json /svc.dev.json

ENV GOOGLE_APPLICATION_CREDENTIALS /svc.dev.json

ENTRYPOINT ["/server", "./config.yaml", "no-auth"]