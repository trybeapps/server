FROM elasticsearch:5

ENV ES_JAVA_OPTS="-Des.path.conf=/etc/elasticsearch"

ENV ES_JAVA_OPTS="-Xms1g -Xmx1g"

RUN /usr/share/elasticsearch/bin/elasticsearch-plugin install --batch ingest-attachment

EXPOSE 9200
EXPOSE 9300

CMD ["-E", "network.host=0.0.0.0", "-E", "discovery.zen.minimum_master_nodes=1"]
