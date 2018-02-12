FROM mongo

RUN apt-get update
RUN apt-get -y install wget unzip
RUN wget https://cs.stanford.edu/people/alecmgo/trainingandtestdata.zip
RUN unzip trainingandtestdata.zip
RUN sed -i '1s;^;polarity,id,date,query,user,text\n;' training.1600000.processed.noemoticon.csv

RUN mkdir -p /data/db2 \
    && echo "dbpath = /data/db2" > /etc/mongodb.conf \
    && chown -R mongodb:mongodb /data/db2


RUN mongod --fork --logpath /var/log/mongod.log --dbpath /data/db2 --smallfiles \
    && mongoimport --host=localhost --drop --db social_net --collection tweets --type csv --headerline --file training.1600000.processed.noemoticon.csv \
    && mongod --dbpath /data/db2 --shutdown \
    && chown -R mongodb /data/db2

VOLUME /data/db2

CMD ["mongod", "--config", "/etc/mongodb.conf", "--smallfiles"]