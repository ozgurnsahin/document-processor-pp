import os
import logging
import signal
import sys
from fastapi import FastAPI
import uvicorn

from app.grpc_server import GRPCServer


logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

app = FastAPI(title="Document Processing Service")


@app.on_event("startup")
async def startup_event():
    global grpc_server

    grpc_port = os.getenv("GRPC_PORT")
    grpc_server = GRPCServer(port=grpc_port)
    grpc_server.start()

    logger.info("Application startup complete")


@app.on_event("shutdown")
async def shutdown_event():
    global grpc_server
    if grpc_server:
        grpc_server.stop(grace=5)
        logger.info("Application shutdown complete")


@app.get("/")
async def root():
    return {"message": "Document Processing Service API"}


@app.get("/health")
async def health_check():
    return {"status": "healthy"}


def handle_shutdown(signum, frame):
    logger.info(f"Received signal {signum}. Shutting down...")
    if grpc_server:
        grpc_server.stop(grace=5)
    sys.exit(0)


signal.signal(signal.SIGINT, handle_shutdown)
signal.signal(signal.SIGTERM, handle_shutdown)

if __name__ == "__main__":
    uvicorn.run("app:app", host="0.0.0.0", port=8081, reload=True)
