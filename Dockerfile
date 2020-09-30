# stage 1: build the binary
# we are using alpine Linux with the latest version of golang
FROM golang:1.15-alpine as golang

# first install some dependencies
# (we are using the static versions for each for them as we want these libraries included statically, not dynamically!)
# czmq requires libzmq which in turn requires libsodium
# on alpine Linux we also need to install some specific tools to build C and C++ programs
# libsodium also requires libuuid, which is included in util-linux-dev
RUN apk add --no-cache libzmq-static czmq-static czmq-dev libsodium-static build-base util-linux-dev git

# now we do the magic command mentioned here
# https://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker?noredirect=1&lq=1
# this fools the C compiler into thinking we have glibc installed while we are actually using musl
# since we are compiling statically, it makes sense to use musl as it is smaller
# (and it uses the more permissive MIT license if you want to distribute your binary in some form, but check your other libraries before!)
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

# create your project directory for the Go project
WORKDIR /go/src/mqtt-to-zmq

# copy in all your Go files, assuming these are in the same directory as your Dockerfile
COPY . .

# here is the first hack: we need to tell CGO to use g++ instead of gcc or else it will struggle with libzmq, which is written in C++
# creating and empty C++ file actually works for this
RUN touch ./dummy.cc

# now run go install (go build could also work here but your binary would end up in a different directory)
# the -ldflags will be passed along the CGO toolchain
# -extldflags is especially important here, it has two important flags that we need:
# -static tells the compiler to use static linking, which does the actual magic of putting everything into one binary
# -luuid is needed to correctly find the uuid library that czmq uses
RUN go get
RUN go install -a -ldflags '-linkmode external -w -s -extldflags "-static -luuid" ' .

# stage 2: here is your actual image that will later run on your Docker host
# you can also use alpine here if you so choose
FROM scratch

# now we just copy over the completed binary from the old builder image
COPY --from=golang /go/bin/mqtt-to-zmq bin

# and we start our program
ENTRYPOINT ["./bin"]
