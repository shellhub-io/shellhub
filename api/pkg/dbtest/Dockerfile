FROM mongo:4.4.8

COPY mongod.conf /etc/mongod.conf

RUN mkdir -p /var/log/mongodb

RUN chown -R mongodb:mongodb /var/log/mongodb

EXPOSE 27017

CMD ["mongod", "--config", "/etc/mongod.conf"]