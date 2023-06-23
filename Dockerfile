
# Build stage for ANTLR JAR
FROM maven:3.9.2-amazoncorretto-17-debian-bullseye as antlr-jar

ENV DEBIAN_FRONTEND=noninteractive
ENV ANTLR_VERSION=4.13.0

RUN mount=type=cache,target=/var/cache/apt \
    apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        wget

WORKDIR /app
ADD https://github.com/antlr/antlr4/archive/refs/tags/$ANTLR_VERSION.tar.gz /app/
RUN tar -xzf $ANTLR_VERSION.tar.gz && \
    cd antlr4-$ANTLR_VERSION && \
    mount=type=cache,target=/root/.m2 \
    mvn clean install -DskipTests -Dmaven.repo.local=/root/.m2/repository && \
    cd ..

# Build stage for ANTLR generated Go code
FROM amazoncorretto:17-alpine-full as antlr-build

ENV ANTLR_VERSION=4.13.0

WORKDIR /app
COPY --from=antlr-jar \
    /app/antlr4-$ANTLR_VERSION/tool/target/antlr4-$ANTLR_VERSION-complete.jar \
    /app/antlr4-$ANTLR_VERSION-complete.jar
COPY --link antlr/grammar /app/grammar
RUN java -jar /app/antlr4-$ANTLR_VERSION-complete.jar -Dlanguage=Go -listener -no-visitor -o /app/generated/vcluster grammar/VCluster.g4 && \
    java -jar /app/antlr4-$ANTLR_VERSION-complete.jar -Dlanguage=Go -listener -no-visitor -o /app/generated/services grammar/Services.g4

# Build stage for Go
FROM golang:1.20.5-bullseye as go-build
WORKDIR /app
COPY --from=antlr-build /app/generated /app/antlr/generated
COPY go.mod go.sum ./
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
