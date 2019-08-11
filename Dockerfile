FROM golang:1.12 AS build
RUN go install
RUn go build -o backend .

FROM chromedp/headless-shell:78.0.3876.0 AS runtime
COPY --from=build backend /bin/backend

CMD ["backend"]
