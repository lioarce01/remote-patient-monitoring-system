from influxdb import InfluxDBClient
import pandas as pd
import os
import logging

logger = logging.getLogger(__name__)

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

def fetch_observations() -> pd.DataFrame:
    try:
        client = get_influx_client()
        query = """
            SELECT heart_rate FROM vitals
            WHERE heart_rate > 0
        """
        result = client.query(query)
        points = list(result.get_points())

        if not points:
            logger.warning("⚠️ No observations retrieved from InfluxDB.")
            # Retorna un DataFrame vacío con las columnas necesarias
            return pd.DataFrame()

        df = pd.DataFrame(points)

        logger.info(f"✅ Retrieved {len(df)} observations from InfluxDB.")
        return df

    except Exception as e:
        logger.error(f"🚨 Failed to fetch observations from InfluxDB: {e}")
        # Retorna un DataFrame vacío con columnas
        return pd.DataFrame()