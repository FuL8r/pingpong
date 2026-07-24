FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /pingpong ./cmd/pingpong

FROM gcr.io/distroless/static:nonroot
COPY --from=build /pingpong /pingpong
ENV PINGPONG_ADDR=:8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/pingpong"]
