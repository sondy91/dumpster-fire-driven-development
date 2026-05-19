module github.com/themetrix/daily-metrics

go 1.26.2

replace github.com/themetrix/metrics-api => ../metrics-api

require (
	github.com/mattn/go-sqlite3 v1.14.42
	github.com/themetrix/metrics-api v0.0.0-00010101000000-000000000000
)
