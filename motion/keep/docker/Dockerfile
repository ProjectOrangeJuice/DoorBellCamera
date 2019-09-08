FROM denismakogon/gocv-alpine:4.0.1-buildstage as build-stage

RUN go get -u -d gocv.io/x/gocv
RUN cd $GOPATH/src/gocv.io/x/gocv && go build -o $GOPATH/bin/gocv-version ./cmd/version/main.go

WORKDIR ./
RUN go get "github.com/mattn/go-sqlite3"
RUN go get "github.com/streadway/amqp"
RUN go get "github.com/go-redis/redis"
RUN go get -u -d gocv.io/x/gocv
COPY *.go ./
RUN go build *.go

FROM denismakogon/gocv-alpine:4.0.1-runtime

COPY --from=build-stage /keep_motion /keep_motion


CMD ["./keep_motion"]
