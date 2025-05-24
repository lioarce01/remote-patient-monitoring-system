module github.com/lioarce01/remote_patient_monitoring_system/pkg/common

go 1.23.0

toolchain go1.23.6

require (
	github.com/influxdata/influxdb1-client v0.0.0-20220302092344-a9ab5670611c
	github.com/segmentio/kafka-go v0.4.48
	gorm.io/gorm v1.26.1
)

replace github.com/lioarce01/remote_patient_monitoring_system/pkg/common => ./

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sync v0.9.0 // indirect
	golang.org/x/text v0.20.0 // indirect
)

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/lib/pq v1.10.9
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	gorm.io/driver/postgres v1.5.11
)
