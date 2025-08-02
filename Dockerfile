FROM golang:1.24-bullseye AS builder

RUN apt-get update && apt-get install -y \
    g++ cmake wget unzip pkg-config git && \
    rm -rf /var/lib/apt/lists/*

# Install toml++ headers
RUN wget https://github.com/marzer/tomlplusplus/archive/refs/tags/v3.4.0.zip && \
    unzip v3.4.0.zip && \
    cp -r tomlplusplus-3.4.0/include/* /usr/local/include/ && \
    rm -rf v3.4.0.zip tomlplusplus-3.4.0

ENV CGO_ENABLED=1
ENV CPATH=/usr/local/include

WORKDIR /app
COPY . .

RUN go build -v -o /app/bin/sqlsec ./cmd/sqlsec


# Runtime stage
FROM debian:bullseye-slim

WORKDIR /app
COPY --from=builder /app/bin/sqlsec /usr/local/bin/sqlsec

EXPOSE 8888
ENTRYPOINT ["/usr/local/bin/sqlsec"]
