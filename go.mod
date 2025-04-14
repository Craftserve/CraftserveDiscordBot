module csrvbot

go 1.21
toolchain go1.24.1

require (
	github.com/Craftserve/monies v0.0.0-20230628121509-708cba760847
	github.com/bwmarrin/discordgo v0.28.1
	github.com/getsentry/sentry-go v0.27.0
	github.com/go-gorp/gorp v2.2.0+incompatible
	github.com/go-sql-driver/mysql v1.7.0
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.9.0
)

require (
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/lib/pq v1.7.1 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/poy/onpar v1.0.0 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/ziutek/mymysql v1.5.4 // indirect
	golang.org/x/crypto v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)

replace github.com/go-gorp/gorp => github.com/Rekseto/gorp v2.2.1-0.20221012142044-f062c65fa536+incompatible
