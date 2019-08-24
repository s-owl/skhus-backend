# build stage
FROM golang:1.13rc1-buster AS build

RUN mkdir /build
COPY ./ /build
WORKDIR /build
RUN go build -o skhus-backend .

# product stage
FROM chromedp/headless-shell:77.0.3865.42

RUN mkdir /app
WORKDIR /app
COPY --from=build /build/skhus-backend .

ENV PATH=$PATH:/headless-shell

EXPOSE 8080
ENTRYPOINT ["./skhus-backend"]
