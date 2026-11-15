FROM golang:1.25-alpine AS build

#Stage 1: build

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -o /out/node ./cmd/node

# STAGE 2:RUN

FROM alpine:3
COPY --from=build /out/node /node
EXPOSE 8080
ENTRYPOINT ["/node"]




