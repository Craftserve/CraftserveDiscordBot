module csrvbot

go 1.13

require (
	github.com/bwmarrin/discordgo v0.28.1
	github.com/go-gorp/gorp v2.2.0+incompatible
	github.com/go-sql-driver/mysql v1.7.0
	github.com/lib/pq v1.7.1 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/poy/onpar v1.0.0 // indirect
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.9.0
	github.com/ziutek/mymysql v1.5.4 // indirect
)

replace github.com/go-gorp/gorp => github.com/Rekseto/gorp v2.2.1-0.20221012142044-f062c65fa536+incompatible
