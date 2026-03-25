FROM golang:1.22-alpine AS builder
 WORKDIR /app

 COPY go.mod go.sum ./

 RUN go mod download


COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-w -s " -o server ./cmd/server


FROM alpine:3.19

WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/server .

COPY --from=builder /app/migrations ./migrations

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8084

CMD [ "./server" ]
