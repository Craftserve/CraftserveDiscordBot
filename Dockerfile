FROM golang:1.17-bullseye

ADD . /app
WORKDIR /app

RUN go build -o csrvbot csrvbot/cmd/bot
RUN chmod +x csrvbot

CMD ["./csrvbot"]