FROM solr:9.2.1

COPY ./conf/ /opt/solr/server/solr/configsets/my_core_conf/conf/

ENV SOLR_OPTS="-Dsolr.disable.allowUrls=true"

CMD ["solr-precreate", "my_core", "/opt/solr/server/solr/configsets/my_core_conf"]