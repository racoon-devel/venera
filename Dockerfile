FROM golang:1.17.1

RUN wget "http://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-1.14.0.tar.gz" && \
   tar -zxv -C /usr/local/ -f libtensorflow-cpu-linux-x86_64-1.14.0.tar.gz && \
   rm -f libtensorflow-cpu-linux-x86_64-1.14.0.tar.gz

ENV LD_LIBRARY_PATH /usr/local/lib/
ENV LIBRARY_PATH /usr/local/lib/

# Install chrome
RUN apt-get update && apt-get install -y \
	apt-transport-https \
	ca-certificates \
	curl \
	gnupg \
	--no-install-recommends \
	&& curl -sSL https://dl.google.com/linux/linux_signing_key.pub | apt-key add - \
	&& echo "deb https://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google-chrome.list \
	&& apt-get update && apt-get install -y \
	google-chrome-stable \
	fontconfig \
	fonts-ipafont-gothic \
	fonts-wqy-zenhei \
	fonts-thai-tlwg \
	fonts-kacst \
	fonts-symbola \
	fonts-noto \
    libx11-xcb1 \
    libxt6 \
    libdbus-glib-1-2 \
	--no-install-recommends \
	&& apt-get purge --auto-remove -y curl gnupg \
	&& rm -rf /var/lib/apt/lists/*

# Copy configuration
WORKDIR /etc/venera
COPY configs/venera.conf venera.conf

# Bind port
EXPOSE 80/tcp

# Storage
WORKDIR "/var/lib/venera"
COPY content/ .

# Build venera
WORKDIR /go/src/github.com/racoon-devel/venera
COPY . .
ENV GOBIN=/go/bin
RUN go install -v ./venera.go && rm -rf *

CMD ["venera", "-verbose"]