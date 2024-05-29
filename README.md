# Pengam

A simple monitoring library solution. The concept is the following:

- an application monitors itself: cpu / memory / ksoftirq
- the application sends data to the data store only if there is a change in the metrics. otherwise no data point is sent.
- the application regularly updates its own hearbeat.

# Advantages:

The standard monitoring solutions send data metrics to the central entity (database -> Grafana, Datadog etc.) on a regular base. That is an overhead because under normal circumstances most applications I met send very similar metrics (CPU, memory) resulting in a lot of data but not much information. This simple approach sends ONLY information - changes -, no data (in terms of information content).

Numbers: application running for 2 days. The internal check loop was set to 15 seconds. Generated data points: 68 rows in the PostgreSQL database.

# Usage

## Create DB schema

We use a database schema `metrics`, but can be any database schema or any table name. These will be passed as parameter anyway. However, the table structure must be as below. Any database can be used that handles time series data on an efficient way _with index_. PostgreSQL does.


```sql
CREATE TABLE metrics.monitor (
  ts timestamp(1) default now(),
  identifier TEXT, -- whatever metadata that helps to identify the application
  ip inet,
  cpu_percent  FLOAT,
  memory_percent FLOAT,
  ksoftirqd       FLOAT,
  process_name TEXT
);

CREATE INDEX idx_timestamp_brin_monitor ON metrics.monitor USING brin(ts);

CREATE TABLE metrics.alive (
  active boolean,
  lastPing timestamp,
  identifier TEXT UNIQUE, -- whatever metadata that helps to identify the application
  ip inet
);
```

It uses a BRIN indexed PostgreSQL table. If BRIN is new to you, see [https://www.postgresql.org/docs/current/brin.html](here).

## Application

- need a database handler
- need to start the watching hook
- need to shut down gracefully. The latter is important: the application must unregister itself from the hearbeat table.


## Monitoring

Visualize the data from the data table as you will. Check the heartbeat table for alerting. Eg:

```SQL
SELECT * FROM metrics.alive WHERE active = true AND lastPing > NOW()-INTERVAL('30s');

```

This query returns those entries that _should_ be up but sent no heartbeat in 30s. Alarm can be set for that.

Note: the application has an inner loop of 15 second check.
