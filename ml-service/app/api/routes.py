from fastapi import APIRouter
from fastapi import Query, HTTPException
from apscheduler.schedulers.background import BackgroundScheduler
from ..entities.schema import Observation, PredictionResponse
from ..models.model import predict as ml_predict
from ..services.training import train_model
from app.scripts.batch_training import batch_train_all
from app.db.influx_client import fetch_all_patient_ids
from typing import List
import os
import logging

router = APIRouter()
scheduler = BackgroundScheduler()

scheduler_started = False

MODEL_PATH = os.path.join(os.path.dirname(__file__), "..", "model", "model.joblib")

@router.on_event("startup")
async def startup_event():
    global scheduler_started
    logging.info("Startup: Running batch training for all patients...")
    batch_train_all()

    if not scheduler_started:
        scheduler.add_job(batch_train_all, "interval", hours=24, id="batch_train_all")
        scheduler.start()
        scheduler_started = True
        logging.info("Scheduler started for batch training every 24 hours.")

@router.get("/v1/models")
def get_model_info():
    model_exists = os.path.exists(MODEL_PATH)
    return {
        "name": "isolation-forest",
        "version": "1.0",
        "status": "loaded" if model_exists else "not loaded"
    }

@router.get("/health")
def healthcheck():
    return {
        "status": "Service healthy"
    }

@router.post("/predict")
def predict(obs: Observation):
    vitals = {"heart_rate": obs.heart_rate}
    try:
        prediction, score = ml_predict(obs.patient_id, vitals)
        return PredictionResponse(prediction=prediction, anomaly_score=score)
    
    except FileNotFoundError:
        raise HTTPException(
            status_code=404,
            detail=f"Model for patient {obs.patient_id} not found. Please trigger training first."
        )
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Failed to predict: {str(e)}"
        )


@router.post("/train/all")
def train_all_patients():
    try:
        patient_ids: List[str] = fetch_all_patient_ids()
        if not patient_ids:
            raise HTTPException(status_code=404, detail="No patient IDs found")

        results = []
        for pid in patient_ids:
            success, msg = train_model(pid)
            results.append({
                "patient_id": pid,
                "status": "success" if success else "skipped",
                "message": msg
            })

        return {"results": results}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))