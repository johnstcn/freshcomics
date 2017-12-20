FROM golang:1.9.2-alpine
RUN mkdir -p /go/bin \
  && mkdir -p /go/pkg \
  && mkdir -p /go/src

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

RUN mkdir -p /go/src/github.com/johnstcn/freshcomics
ADD . $GOPATH/src/github.com/johnstcn/freshcomics

WORKDIR /go/src/github.com/johnstcn/freshcomics
RUN go build frontend/freshcomics-frontend.go
#RUN go build crawler/freshcomics-crawler.go

FROM alpine:latest
MAINTAINER Cian Johnston <public@cianjohnston.ie>
COPY --from=0 /go/src/github.com/johnstcn/freshcomics/freshcomics-frontend .
#COPY --from=0 /go/src/github.com/johnstcn/freshcomics/freshcomics-crawler .

ENTRYPOINT ["/freshcomics-frontend"]