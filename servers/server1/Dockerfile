FROM golang:1.14-alpine AS build

WORKDIR /src/
COPY server1.go go.* /src/
RUN CGO_ENABLED=0 go build -o app .

FROM scratch
COPY --from=build /src/ .
CMD ["./app"]