FROM golang:1.13.4-alpine3.10
LABEL maintainer="Delane Jackson <delane.jackson@gmail.com>"

# Create an app directory inside the container
RUN mkdir /app

# Copy all files from the application root on my system to the app directory in the container
ADD . /app

# From now on we'll work in the /app directory 
WORKDIR /app

# Build the go binary inside the app
RUN go build -o f5oauth2

# Now start the go based http server
CMD ["/app/f5oauth2"]

