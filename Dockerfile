FROM golang
RUN go get github.com/dim13/gone
EXPOSE 8001
ENTRYPOINT /go/bin/gone
