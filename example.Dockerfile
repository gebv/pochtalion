FROM golang:1.7

COPY ./ /usr/local/src/pochtalion

RUN go get -v github.com/labstack/echo/... \
    gopkg.in/asaskevich/govalidator.v4 \
    gopkg.in/mailgun/mailgun-go.v1

RUN cd /usr/local/src/pochtalion/ && \
    GOPATH=${PWD}:${GOPATH} \
    CGO_ENABLED=0 \
    go build -o /usr/local/bin/pochtalion -v \
    -a --installsuffix cgo -ldflags \
    "-s" \
    cmd/web/main.go 

WORKDIR /usr/local/src/pochtalion

ENTRYPOINT ["pochtalion"]