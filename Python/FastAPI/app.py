from fastapi import FastAPI

app = FastAPI()

@app.get("/healthz")
async def health():
    return {"status": "ok"}

@app.get("/version")
async def health():
    return {"version": "1.0"}