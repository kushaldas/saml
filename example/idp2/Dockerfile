FROM fedora:41
MAINTAINER Kushal Das <kushal@sunet.se>
COPY ./idp2 /usr/bin/
COPY ./users.json /etc/idp2/
RUN mkdir /data
VOLUME /data
EXPOSE 8080
WORKDIR /data
ENTRYPOINT ["/usr/bin/idp2"]
