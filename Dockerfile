FROM golang:1.12 as builder

WORKDIR /go/src/app

ENV GO111MODULE=on \
    CGO_ENABLED=0  \
    GOOS=linux

COPY go.mod go.sum ./
RUN go mod download

COPY . /go/src/app

RUN go build -a -installsuffix cgo -o main .


FROM gcr.io/distroless/base
COPY --from=builder /go/src/app/main /
CMD ["./main"]