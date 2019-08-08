FROM golang:alpine

# Copy configuration
WORKDIR /etc/venera
COPY configs/venera.conf venera.conf

# Bind port
EXPOSE 80/tcp

# Storage
VOLUME ["/var/lib/venera"]

# Build venera
WORKDIR /go/src/github.com/nightwizard0/venera
COPY . .
ENV GOBIN=/go/bin
RUN go install -v ./venera.go && rm -rf *

CMD ["venera"]