FROM golang:1.15

COPY . /go/src/prometheus-matrix-bot/
RUN go get prometheus-matrix-bot/...
RUN go install prometheus-matrix-bot

CMD prometheus-matrix-bot
