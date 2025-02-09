# Start from the latest golang base image
FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Install dependencies
# This unfortunately does not install the latest yt-dlp
# RUN apt-get update && apt-get install -y \
#     yt-dlp \
#     ffmpeg

# Install dependencies directly from their sources
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install the latest yt-dlp
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp \
    && chmod +x /usr/local/bin/yt-dlp

# Install the latest ffmpeg (unfortunately this does not work)
# RUN curl -sSL https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz \
#     | tar -xJf - --strip-components=1 -C /usr/local/bin/ --wildcards '*/ffmpeg' '*/ffprobe'

# Install ffmpeg from the package manager (which does not have the latest version)
RUN apt-get update && apt-get install -y \
    ffmpeg

# Verify installation
RUN yt-dlp --version
RUN ffmpeg -version

# Build the Go app
RUN GOOS='linux' GOARCH='amd64' GO111MODULE='on'  go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]