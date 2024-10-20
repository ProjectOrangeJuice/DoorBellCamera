FROM golang
WORKDIR ./

RUN go get "github.com/lib/pq"
RUN go get "github.com/dgrijalva/jwt-go"
RUN go get "github.com/gorilla/mux"
RUN go get "github.com/mattn/go-sqlite3"
RUN go get "github.com/streadway/amqp"
RUN go get "github.com/go-redis/redis"
RUN go get "github.com/gorilla/websocket"
RUN mkdir log
COPY *.go ./
RUN go build *.go


CMD ["./auth"]