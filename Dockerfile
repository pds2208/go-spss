FROM golang:alpine3.10 as builder
LABEL stage=builder
WORKDIR /app

RUN mkdir /user && \
    echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
    echo 'nobody:x:65534:' > /user/group

RUN apk add --update alpine-sdk autoconf automake build-base clang cmake \
    libtool make m4 zlib-dev git gettext && \
    rm -rf /var/cache/apk/* && \
    git clone https://github.com/jmcnamara/libxlsxwriter.git && \
    cd libxlsxwriter && make && make install && \
    cd .. && rm -rf libxlsxwriter && \
    wget https://github.com/WizardMac/ReadStat/releases/download/v1.0.2/readstat-1.0.2.tar.gz && \
    zcat readstat-1.0.2.tar.gz | tar xvf - && \
    cd readstat-1.0.2 && ./configure && make && make install && mkdir -p /app/src

COPY . /app/src
WORKDIR /app/src
RUN go mod download && CGO_ENABLED=1 GOPATH=/app GOOS=linux GOARCH=amd64 go build -i -v -o libgo-spss.so -ldflags="-s -w -lreadstat"
WORKDIR /app/src/service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .

# using this multi-stage build reduces the image size from ~1GB to ~8MB

FROM scratch
WORKDIR /app

COPY --from=builder /user/group /user/passwd /etc/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/src/service/main /

WORKDIR /
USER nobody:nobody
ENTRYPOINT ["/main"]
EXPOSE 8080
