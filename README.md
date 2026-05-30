# homeserver

## Compile Go Binary
`GOOS=linux GOARCH=amd64 go build -o cmd/executable/storage-agent.exe ./cmd/storage-agent`

## Make script executable (install.sh script inside cmd/installer/)
`chmod +x install.sh`

## Run as root
`sudo ./install.sh`

## Storage-agent 
- ./storage/ contains files for the agent and mounting 
- ./cmd/storage-agent/ contains the main.go file as the entry point
- uses the same database.go as the server 
- executable created at ./cmd/executable/ 

## Go-server
- ./internals/ contains files for the server
- ./configs/ contains the database.go
- ./cmd/server/ contains the main.go for the server

## run onlyoffice docker container
`docker run -d \
  --name onlyoffice \
  -p 8082:80 \
  -e JWT_ENABLED=true \
  -e JWT_SECRET=supersecret123 \
  onlyoffice/documentserver`

## run go server
`go run -tags "fts5" cmd/server/main.go`

## compile homeserver binary
`go build -tags "fts5" -o cmd/executable/homeserver.exe cmd/server/main.go`

## production directory structure
/etc/homeserver/
└── homeserver.env

/var/lib/homeserver/
└── homeserver.db

/usr/local/bin/
├── homeserver
└── storage-agent