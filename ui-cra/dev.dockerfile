FROM alpine

RUN mkdir /html
COPY ./build /html

CMD ["sh", "-c", "mv -v /html/* /target/"]
