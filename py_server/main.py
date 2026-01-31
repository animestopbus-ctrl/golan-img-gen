import logging
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from py_server.inference import generate_image  # Note: py_server. for import in container

app = FastAPI()

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class PromptRequest(BaseModel):
    prompt: str

@app.post("/generate")
async def generate(request: PromptRequest):
    try:
        image_bytes = generate_image(request.prompt)
        return image_bytes  # FastAPI will handle as Response(content=image_bytes, media_type="image/png")
    except Exception as e:
        logger.error(f"Generation failed: {e}")
        raise HTTPException(status_code=500, detail="Image generation failed")