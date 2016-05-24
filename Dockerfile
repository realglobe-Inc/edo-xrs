FROM golang:1.6-wheezy

MAINTAINER Junichi Miyazaki

RUN cd /tmp && \
    wget 'ftp://ftp.csx.cam.ac.uk/pub/software/programming/pcre/pcre-8.37.tar.gz' && \
    tar xvzf pcre-8.37.tar.gz && \
    cd pcre-8.37/ && \
    ./configure --enable-utf8 --enable-unicode-properties --prefix=/usr && \
    make install

RUN mkdir -p /go/src/github.com/realglobe-Inc/edo-xrs
WORKDIR /go/src/github.com/realglobe-Inc/edo-xrs

COPY . /go/src/github.com/realglobe-Inc/edo-xrs
RUN go-wrapper download
RUN go-wrapper install

RUN sed -i -e"s/url=mongodb:\/\/localhost/url=mongodb:\/\/mongo:27017/" conf/app.conf

CMD ["go-wrapper", "run"]
