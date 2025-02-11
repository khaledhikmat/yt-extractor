BUILD_DIR="./build"
DIST_DIR="./dist"

clean_build:
	if [ -d "${BUILD_DIR}" ]; then rm -r ${BUILD_DIR}; fi

clean_dist:
	if [ -d "${DIST_DIR}" ]; then rm -r ${DIST_DIR}; fi; mkdir ${DIST_DIR}

test:
	echo "Invoking test cases..."
	#go test ./service/cloudconvert -run TestConversion -v -count=1

build: clean_dist clean_build test
	GOOS='linux' GOARCH='amd64' GO111MODULE='on' go build -o "${BUILD_DIR}/yt-extractor-app" .

# docker images | grep 'yt-extractor' | awk '{print $3}' | xargs docker rmi -f
# docker ps -a | grep 'yt-extractor' | awk '{print $1}' | xargs docker rm -f
# docker system prune -a
dockerize: clean_dist clean_build test build
	docker buildx build --platform linux/amd64 -t khaledhikmat/yt-extractor-app:latest .

push-2-hub: clean_dist clean_build test build dockerize
	docker login
	docker push khaledhikmat/yt-extractor-app:latest

start: clean_dist clean_build test
	go run main.go

compose-up:
	# force a pull and run detached
	# for some reason, running from here does not run due to amd64 platform
	docker compose up -d --pull always

compose-down:
	docker compose down
