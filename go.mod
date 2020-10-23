module main

go 1.14

require (
	calendar v0.0.0
	docs v0.0.0
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/gin-gonic/gin v1.6.3 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14 // indirect
	github.com/swaggo/gin-swagger v1.2.0 // indirect
	gobot.io/x/gobot v1.14.0 // indirect
	heater v0.0.0
)

replace heater => ./src/heater

replace docs => ./src/docs

replace calendar => ./src/calendar
