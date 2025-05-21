import sys
import os
import logging

sys.path.append(os.path.join(os.path.dirname(__file__), ".."))

from app.services.training import train_model

logging.basicConfig(level=logging.INFO)

if len(sys.argv) < 2:
    logging.error("❌ Usage: python train.py <patient_id>")
    sys.exit(1)

patient_id = sys.argv[1]
success, msg = train_model(patient_id)

if success:
    print(f"✅ {msg}")
else:
    print(f"❌ {msg}")
