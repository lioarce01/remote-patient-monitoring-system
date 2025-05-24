import os
from app.services.training import train_model
from app.db.influx_client import fetch_all_patient_ids

def batch_train_all():
    patient_ids = fetch_all_patient_ids()

    if not patient_ids:
        print("No patient IDs found.")
        return

    for patient_id in patient_ids:
        print(f"Training model for patient {patient_id}...")
        success, msg = train_model(patient_id)
        status = "✅" if success else "❌"
        print(f"{status} {msg}")

if __name__ == "__main__":
    batch_train_all()
