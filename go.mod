module main

go 1.14

require (
	calendar v0.0.0
	docs v0.0.0
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/gin-gonic/gin v1.6.3
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14
	github.com/swaggo/gin-swagger v1.2.0
	heater v0.0.0
)

replace heater => ./src/heater

replace docs => ./src/docs

replace calendar => ./src/calendar
