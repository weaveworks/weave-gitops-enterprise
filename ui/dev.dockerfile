FROM alpine:3.18

RUN mkdir /html
COPY ./build /html

CMD ["sh", "-c", "mv -v /html/* /target/"]
