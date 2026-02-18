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