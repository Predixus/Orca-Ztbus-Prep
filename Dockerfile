FROM postgres:17

RUN apt-get update && \
    apt-get install -y postgresql-17-partman && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
