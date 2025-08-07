FROM golang:1.24-alpine as builder

WORKDIR /

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /shorter

FROM scratch

COPY --from=builder /shorter /shorter

ENTRYPOINT ["/shorter"]
