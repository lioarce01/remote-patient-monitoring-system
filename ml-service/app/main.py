from fastapi import FastAPI
from fastapi import Request
from pydantic import BaseModel
from .model import predict as ml_predict
from .influx_client import fetch_observations
from sklearn.ensemble import IsolationForest
import joblib
import os
import logging

logging.basicConfig(level=logging.INFO)
app = FastAPI()

MODEL_DIR = os.path.join(os.path.dirname(__file__), "model")
MODEL_PATH = os.path.join(MODEL_DIR, "model.joblib")

class Observation(BaseModel):
    id: str
    patient_id: str
    heart_rate: float
    effective_date_time: str

class PredictionResponse(BaseModel):
    prediction: bool
    anomaly_score: float

@app.middleware("http")
async def log_requests(request: Request, call_next):
    logging.info(f"🔍 Incoming request: {request.method} {request.url}")
    return await call_next(request)

@app.on_event("startup")
def on_startup():
    try:
        success, msg = train_model()
        if not success:
            logging.warning(f"⚠️ Startup training skipped: {msg}")
    except Exception as e:
        logging.error(f"🚨 Failed to train model on startup: {e}")

@app.get("/health")
async def health():
    return {"status": "ok"}

@app.get("/v1/models")
def get_model_info():
    model_exists = os.path.exists(MODEL_PATH)
    return {
        "name": "isolation-forest",
        "version": "1.0",
        "status": "loaded" if model_exists else "not loaded"
    }

@app.post("/predict")
def predict(obs: Observation):
    vitals_dict = {
        "heart_rate": obs.heart_rate
    }
    prediction, score = ml_predict(vitals_dict)
    return PredictionResponse(prediction=prediction, anomaly_score=score)

@app.post("/train")
def retrain_model():
    try:
        success, msg = train_model()
        if success:
            return {"status": "success", "message": msg}
        else:
            return {"status": "skipped", "message": msg}
    except Exception as e:
        logging.error(f"🚨 Manual training failed: {e}")
        return {"status": "error", "message": str(e)}

def train_model():
    logging.info("⏳ Training model...")
    df = fetch_observations()
    if df.empty:
        logging.warning("⚠️ No data available to train the model.")
        return False, "No data available"

    df = df.dropna(subset=["heart_rate"])
    X = df[["heart_rate"]].to_numpy()

    model = IsolationForest(n_estimators=100, contamination=0.01, random_state=42)
    model.fit(X)

    os.makedirs(MODEL_DIR, exist_ok=True)
    joblib.dump({"model": model, "features": ["heart_rate"]}, MODEL_PATH)
    logging.info(f"✅ Model trained and saved at {MODEL_PATH}")
    return True, "Model trained successfully"
