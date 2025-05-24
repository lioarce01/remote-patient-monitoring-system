from pydantic import BaseModel

class Observation(BaseModel):
    id: str
    patient_id: str
    heart_rate: float
    effective_date_time: str

class PredictionResponse(BaseModel):
    prediction: bool
    anomaly_score: float
