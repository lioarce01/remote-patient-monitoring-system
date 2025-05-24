from fastapi import Request

async def log_requests(request: Request, call_next):
    print(f"🔍 Incoming request: {request.method} {request.url}")
    return await call_next(request)
