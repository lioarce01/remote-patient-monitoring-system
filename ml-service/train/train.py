import pandas as pd
from sklearn.ensemble import IsolationForest
import joblib
from influx_client import fetch_observations

df = fetch_observations()
df = df.dropna(subset=["heart_rate"])

model = IsolationForest(contamination=0.01, random_state=42)
model.fit(df[["heart_rate"]])

FEATURE_NAMES = ["heart_rate"]

joblib.dump({"model": model, "features": FEATURE_NAMES}, "model/model.joblib")

print("✅ Modelo entrenado y guardado.")
