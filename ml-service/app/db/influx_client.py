import pandas as pd
import logging
from .credentials import get_influx_client

logger = logging.getLogger(__name__)

def fetch_observations(patient_id: str) -> pd.DataFrame:
    try:
        client = get_influx_client()
        query = f"""
            SELECT heart_rate FROM vitals
            WHERE heart_rate > 0 AND "patient_id" = '{patient_id}'
        """
        logger.info(f"Query to InfluxDB: {query}")

        result = client.query(query)
        points = list(result.get_points())

        if not points:
            logger.warning(f"No observations found for patient_id {patient_id}")
            return pd.DataFrame()

        df = pd.DataFrame(points)
        logger.info(f"Retrieved {len(df)} observations for patient_id {patient_id}")
        return df

    except Exception as e:
        logger.error(f"Error fetching observations for patient {patient_id}: {e}")
        return pd.DataFrame()

def fetch_all_patient_ids() -> list[str]:
    try:
        client = get_influx_client()
        query = 'SHOW TAG VALUES FROM "vitals" WITH KEY = "patient_id"'
        result = client.query(query)
        values = list(result.get_points())
        return [v["value"] for v in values]
    except Exception as e:
        logger.error(f"Failed to fetch patient IDs: {e}")
        return []