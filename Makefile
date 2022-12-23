BINARY_NAME=build/youtube_thumbnail_api

build:
	go build -o ${BINARY_NAME} src/main/main.go

run:
	go build -o ${BINARY_NAME} src/main/main.go
	./${BINARY_NAME}