import os
import joblib
import numpy as np
import logging

logging.basicConfig(level=logging.INFO)

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
model_path = os.path.join(BASE_DIR, "model", "model.joblib")

model = None
features = None

def load_model():
    global model, features
    if model is None or features is None:
        try:
            logging.info("Loading model...")
            data = joblib.load(model_path)
            model = data["model"]
            features = data["features"]
            logging.info("Model loaded.")
        except FileNotFoundError:
            logging.warning("⚠️ Model file not found. Prediction will be unavailable.")
            model, features = None, None
        except Exception as e:
            logging.error(f"🚨 Failed to load model: {e}")
            model, features = None, None

def predict(vitals: dict[str, float]) -> tuple[bool, float]:
    load_model()
    if model is None or features is None:
        raise RuntimeError("Model is not loaded. Train it before making predictions.")
    
    x = np.array([[vitals[feature] for feature in features]])
    result = model.predict(x)
    score = model.decision_function(x)
    return bool(result[0] == -1), float(score[0])

