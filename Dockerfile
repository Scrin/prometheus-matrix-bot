FROM golang:1.17

WORKDIR /go/src/github.com/Scrin/prometheus-matrix-bot/
COPY . ./
RUN go install .

CMD prometheus-matrix-bot
