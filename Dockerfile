# build stage
FROM golang:1.13-buster AS build

RUN mkdir /build
COPY ./ /build
WORKDIR /build
RUN go build -o skhus-backend .

# product stage
FROM chromedp/headless-shell:78.0.3902.4

RUN apt-get update && \
    apt-get install -y dumb-init && ca-certificates
RUN mkdir /app
WORKDIR /app
COPY --from=build /build/skhus-backend .

ENV PATH=$PATH:/headless-shell

EXPOSE 8080
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["./skhus-backend"]
