FROM golang:alpine as builder
LABEL maintainer="Isaac Asensio <isaac.asensio@gmail.com>"

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	ca-certificates

COPY . /go/src/github.com/isaacasensio/byebye-slack-channel

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		git \
		gcc \
		libc-dev \
		libgcc \
		make \
	&& cd /go/src/github.com/isaacasensio/byebye-slack-channel \
	&& make static \
	&& mv byebye-slack-channel /usr/bin/byebye-slack-channel \
	&& apk del .build-deps \
	&& rm -rf /go \
	&& echo "Build complete."

FROM scratch

COPY --from=builder /usr/bin/byebye-slack-channel /usr/bin/byebye-slack-channel
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

ENTRYPOINT [ "byebye-slack-channel" ]
CMD [ "--help" ]