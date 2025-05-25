module github.com/maklybae/plagiarism-checker/storage

go 1.24.3

require (
	github.com/Masterminds/squirrel v1.5.4
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx v3.6.2+incompatible
	github.com/jackc/pgx/v5 v5.7.5
	github.com/maklybae/plagiarism-checker/genproto/go v0.0.0-00010101000000-000000000000
	github.com/pressly/goose/v3 v3.24.3
	google.golang.org/grpc v1.72.1
)

replace github.com/maklybae/plagiarism-checker/genproto/go => ../genproto/go

require (
	github.com/cockroachdb/apd v1.1.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
