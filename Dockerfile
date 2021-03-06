FROM alpine:3.4
EXPOSE 5000

RUN apk update && \
	apk add curl

RUN curl -L https://github.com/erasche/chado-jbrowse-connector/releases/download/v0.9.3/chado-jbrowse-connector_linux_amd64 > /usr/bin/chado-jbrowse-connector && \
	chmod +x /usr/bin/chado-jbrowse-connector

ENTRYPOINT ["/usr/bin/chado-jbrowse-connector"]
