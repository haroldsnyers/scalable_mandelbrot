# Two-stage build:
#    first  FROM prepares a binary file in full environment ~780MB
#    second FROM takes only binary file ~10MB

FROM golang:1.14-alpine AS build

ENV GO111MODULE=on

## We create an /app directory within our
## image that will hold our application source files
RUN mkdir /app
## We copy everything in the root directory
## into our /app directory
ADD . /app
## We specify that we now wish to execute
## any further commands inside our /app directory
WORKDIR /app
RUN go mod init go/app
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app .

# second stage to obtain a very small image
FROM scratch
COPY --from=build /app ./
# RUN chmod +x ./app
CMD ["./app"]