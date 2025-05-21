import os
import joblib
import logging
from sklearn.ensemble import IsolationForest
from ..db.influx_client import fetch_observations

MODEL_DIR = os.path.join(os.path.dirname(__file__), "..", "model")

def get_model_path(patient_id: str) -> str:
    return os.path.join(MODEL_DIR, f"{patient_id}_model.joblib")

def train_model(patient_id: str) -> tuple[bool, str]:
    logging.info(f"⏳ Training model for patient {patient_id}...")
    
    df = fetch_observations(patient_id)
    if df.empty:
        logging.warning(f"⚠️ No data for patient {patient_id}.")
        return False, "No data available"

    df = df.dropna(subset=["heart_rate"])
    if len(df) < 20:
        logging.warning(f"⚠️ Not enough data to train model for patient {patient_id}. Only {len(df)} records.")
        return False, f"Not enough data to train model (need at least 30, got {len(df)})"

    X = df[["heart_rate"]].to_numpy()
    model = IsolationForest(n_estimators=100, contamination=0.05, random_state=42)
    model.fit(X)

    os.makedirs(MODEL_DIR, exist_ok=True)
    patient_model_path = get_model_path(patient_id)
    joblib.dump({"model": model, "features": ["heart_rate"]}, patient_model_path)

    logging.info(f"✅ Model for patient {patient_id} trained and saved at {patient_model_path}")
    logging.info(f"📦 Model file timestamp: {os.path.getmtime(patient_model_path)}")
    
    return True, "Model trained successfully"
