# Build stage for ANTLR JAR
FROM maven:3.9.1-amazoncorretto-17-debian-bullseye as antlr-jar

ENV DEBIAN_FRONTEND=noninteractive

RUN mount=type=cache,target=/var/cache/apt \
    apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        wget

WORKDIR /app
COPY antlr/grammar /app/grammar
ADD https://github.com/antlr/antlr4/archive/refs/tags/4.12.0.tar.gz /app/
RUN tar -xzf 4.12.0.tar.gz && \
    cd antlr4-4.12.0 && \
    mount=type=cache,target=/root/.m2 \
    mvn clean install -DskipTests -Dmaven.repo.local=/root/.m2/repository && \
    cd ..

# Build stage for ANTLR generated Go code
FROM amazoncorretto:17-alpine-full as antlr-build

WORKDIR /app
COPY --from=antlr-jar \
    /app/antlr4-4.12.0/tool/target/antlr4-4.12.0-complete.jar \
    /app/antlr4-4.12.0-complete.jar
COPY --link antlr/grammar /app/grammar
RUN java -jar /app/antlr4-4.12.0-complete.jar -Dlanguage=Go -listener -no-visitor -o /app/generated/vcluster grammar/VCluster.g4 && \
    java -jar /app/antlr4-4.12.0-complete.jar -Dlanguage=Go -listener -no-visitor -o /app/generated/services grammar/Services.g4

# Build stage for Go
FROM golang:1.20.3-bullseye as go-build
WORKDIR /app
COPY --from=antlr-build /app/generated /app/antlr/generated
COPY go.mod go.sum .
RUN mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN go build -o virtual-cluster cmd/virtual-cluster/main.go

# Final stage
FROM ubuntu:22.04
RUN mount=type=cache,target=/var/cache/apt \
    apt-get update
WORKDIR /app
COPY --from=go-build /app/virtual-cluster /app/virtual-cluster
COPY --from=antlr-build /app/generated /app/antlr/generated
ENTRYPOINT ["/app/virtual-cluster"]
