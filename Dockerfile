FROM golang

ADD . /go/src/facade
ENV GOBIN /go/bin
RUN go get github.com/Sirupsen/logrus
RUN go get stathat.com/c/jconfig
RUN go install /go/src/facade/facade.go
RUN cp /go/src/facade/config.json /go/bin/.
WORKDIR /go/bin
ENTRYPOINT /go/bin/facade

EXPOSE 8080

CMD [ "/go/bin/facade" ]
