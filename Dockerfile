# build stage
FROM golang:1.12-stretch AS build

RUN mkdir /build
COPY ./ /build
WORKDIR /build
RUN go build -o skhus-backend .

# product stage
FROM chromedp/headless-shell:latest

RUN mkdir /app
WORKDIR /app
COPY --from=build /build/skhus-backend .

ENV PATH=$PATH:/headless-shell

EXPOSE 8080
ENTRYPOINT ["./skhus-backend"]
