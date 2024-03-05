FROM golang:1.22.0 AS build-stage

WORKDIR /app/project

COPY ./ ./

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/app/...

WORKDIR /app

RUN  rm -rf ./project

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /server /server

CMD ["/server"]