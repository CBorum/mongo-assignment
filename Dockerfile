FROM golang:jessie

RUN go get github.com/gin-gonic/gin
RUN go get gopkg.in/mgo.v2

WORKDIR /go/src/app

ADD src src

ENV DB_URL=mongodb://172.17.0.2/16:27017/social_net

CMD [ "go", "run", "src/main.go" ]