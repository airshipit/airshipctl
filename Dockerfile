FROM golang:1.12

COPY . /go/airshipadm

WORKDIR /go/airshipadm

CMD go install && airshipadm version
