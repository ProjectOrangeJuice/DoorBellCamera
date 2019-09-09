FROM golang
WORKDIR ./

RUN go get "github.com/lib/pq"
RUN go get "github.com/gorilla/mux"
RUN go get "github.com/mattn/go-sqlite3"
RUN go get "github.com/streadway/amqp"
COPY *.go ./
COPY web/ ./web/
RUN go build *.go


CMD ["./service"]