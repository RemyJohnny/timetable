# build stage 
FROM golang:1.24-alpine AS builder

WORKDIR /timetablebot
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main .

#final stage
FROM alpine:latest

RUN apk --no-cache add tzdata
COPY --from=builder /timetablebot/main .

ENV TZ=Europe/Minsk
ENV TIMETABLE_MONGODB_STRING="mongodb connection string"
ENV TIMETABLE_TG_BOT_TOKEN="telegram bot token"
ENV TIMETABLE_ADMINS_USERID="admin user id separated with |"
ENV SEMESTER_START_DATE="yyyy-mm-dd"

CMD [ "./main" ]