FROM golang:latest

RUN wget "http://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-1.14.0.tar.gz" && \
   tar -zxv -C /usr/local/ -f libtensorflow-cpu-linux-x86_64-1.14.0.tar.gz && \
   rm -f libtensorflow-cpu-linux-x86_64-1.14.0.tar.gz

ENV LD_LIBRARY_PATH /usr/local/lib/
ENV LIBRARY_PATH /usr/local/lib/

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