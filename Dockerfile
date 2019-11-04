FROM golang:alpine

# Copy configuration
WORKDIR /etc/venera
COPY configs/venera.conf venera.conf

# Bind port
EXPOSE 80/tcp

# Storage
WORKDIR "/var/lib/venera"
COPY content/ .

# Build venera
WORKDIR /go/src/racoondev.tk/gitea/racoon/venera
COPY . .
ENV GOBIN=/go/bin
RUN go install -v ./venera.go && rm -rf *

CMD ["venera", "-verbose"]