FROM golang AS build
RUN mkdir /build
WORKDIR /build
COPY . .
RUN go mod tidy
RUN go env
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot

FROM alpine
ENV GOOGLE_APPLICATION_CREDENTIALS=key.json
COPY --from=build /build/bot /
CMD ["/bot"]