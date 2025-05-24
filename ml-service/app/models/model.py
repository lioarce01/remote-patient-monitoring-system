import os
import joblib
import numpy as np
import logging

logging.basicConfig(level=logging.INFO)

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__))) 
MODEL_DIR = os.path.join(BASE_DIR, "model")

def predict(patient_id: str, vitals: dict[str, float]) -> tuple[bool, float]:
    model_path = os.path.join(MODEL_DIR, f"{patient_id}_model.joblib")

    try:
        logging.info(f"🔍 Loading model for patient: {patient_id} at {model_path}")
        data = joblib.load(model_path)
        model = data["model"]
        features = data["features"]
    except FileNotFoundError:
        logging.error(f"❌ Model not found for patient {patient_id}")
        raise FileNotFoundError(f"Model not found for patient {patient_id}")
    except Exception as e:
        logging.error(f"🚨 Error loading model for patient {patient_id}: {e}")
        raise RuntimeError(f"Failed to load model for patient {patient_id}")

    x = np.array([[vitals[feature] for feature in features]])
    result = model.predict(x)
    score = model.decision_function(x)

    return bool(result[0] == -1), float(score[0])
