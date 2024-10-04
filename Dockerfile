FROM golang:1.22 AS build
WORKDIR /server/
COPY . .
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
RUN GOOS=linux go build -installsuffix cgo -o hw-cloud cmd/main/main.go

FROM busybox
COPY --from=build /server/hw-cloud /server/hw-cloud
EXPOSE 8080
ENV ENV local
WORKDIR /server/
ENTRYPOINT ["./hw-cloud"]