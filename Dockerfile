FROM golang:1.12

COPY . /go/airshipctl

WORKDIR /go/airshipctl

CMD go install && airshipctl version
