from fastapi import FastAPI
from .api.routes import router as api_router
from .api.middleware import log_requests

app = FastAPI()

app.middleware("http")(log_requests)
app.include_router(api_router)
