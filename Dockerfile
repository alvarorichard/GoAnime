# Use the latest Ubuntu as a base image
FROM ubuntu:latest

# Install necessary packages for adding repositories and downloading software
RUN apt-get update && \
    apt-get install -y software-properties-common wget gnupg2 git sudo

# Add Golang's latest repository
RUN add-apt-repository ppa:longsleep/golang-backports

# Update and install Golang and other essentials
RUN apt-get update && apt-get install -y golang-go mpv yt-dlp

# Install fzf
RUN git clone --depth 1 https://github.com/junegunn/fzf.git ~/.fzf && \
    ~/.fzf/install

# Set the environment variable for Go
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

# Setup workspace
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Clone the specified Git repository
RUN git clone https://github.com/alvarorichard/GoAnime.git $GOPATH/src/GoAnime

# Set the working directory to the cloned repository inside the container
WORKDIR $GOPATH/src/GoAnime

# Run the install script
RUN sudo bash install.sh

# Set the entrypoint to the go-anime application
ENTRYPOINT ["go-anime"]
