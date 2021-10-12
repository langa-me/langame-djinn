FROM golang as build

WORKDIR /all

COPY . .

# # Installs protoc and protoc-gen-go plugin
# ARG VERS="3.15.8"
# ARG ARCH="linux-x86_64"
# RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v${VERS}/protoc-${VERS}-${ARCH}.zip \
#     --output-document=./protoc-${VERS}-${ARCH}.zip && \
#     apt update && apt install -y unzip && \
#     unzip -o protoc-${VERS}-${ARCH}.zip -d protoc-${VERS}-${ARCH} && \
#     mv protoc-${VERS}-${ARCH}/bin/* /usr/local/bin && \
#     mv protoc-${VERS}-${ARCH}/include/* /usr/local/include && \
#     go get -u github.com/golang/protobuf/protoc-gen-go

# # Generate Golang protobuf files
# RUN protoc --go_out=. --go-grpc_out=. djinn.proto

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

ENTRYPOINT ["/server"]