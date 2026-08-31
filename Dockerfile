FROM golang:1.6

# Dependencies
RUN go get github.com/gorilla/securecookie && \
    go get github.com/gorilla/sessions && \
    go get github.com/hoisie/mustache && \
    go get github.com/unixpickle/f1le

# Application install
RUN go install github.com/unixpickle/f1le
# Setup app env
RUN mkdir /f1les

ENTRYPOINT /go/bin/f1le 8080 /f1les
EXPOSE 8080
