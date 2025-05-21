import os
from influxdb import InfluxDBClient

def get_influx_client():
    host = os.getenv("INFLUX_HOST", "influxdb")
    username = os.getenv("INFLUX_USER", "admin")
    password = os.getenv("INFLUX_PASS", "admin")
    database = os.getenv("INFLUX_DB", "telemetry")

    return InfluxDBClient(
        host=host,
        username=username,
        password=password,
        database=database,
        timeout=5,
        retries=3
    )
