FROM alpine

COPY /bin/airshipctl /bin/airshipctl

CMD airshipctl help
