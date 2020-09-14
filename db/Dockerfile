FROM postgres
ENV POSTGRES_USER avito
ENV POSTGRES_PASSWORD 'avito'
ENV POSTGRES_DB avito
COPY init.sql /docker-entrypoint-initdb.d/
