FROM postgres:13

RUN localedef -i sv_SE -c -f UTF-8 -A /usr/share/locale/locale.alias sv_SE.UTF-8
ENV LANG sv_SE.utf8

ENV POSTGRES_USER faktura
ENV POSTGRES_PASSWORD faktura
ENV POSTGRES_DB faktura

# Set PGDATA outside persisted volume to wipe DB on container restart
ENV PGDATA /var/lib/postgresql-tmp/data

# Create database
COPY initdb.d /docker-entrypoint-initdb.d
CMD docker-entrypoint.sh \
      -c 'log_statement=all' \
      -c 'timezone=UTC' \
      -c 'shared_preload_libraries=pg_stat_statements' \
      -c 'pg_stat_statements.max=10000' \
      -c 'pg_stat_statements.track=all'
