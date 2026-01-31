import logging
import torch
from diffusers import AutoPipelineForText2Image
from io import BytesIO
from threading import Thread

logger = logging.getLogger(__name__)

# Load model at startup
logger.info("Loading model...")
pipe = AutoPipelineForText2Image.from_pretrained(
    "stabilityai/sdxl-turbo",
    torch_dtype=torch.float32,
    variant="fp32"
)
pipe.to("cpu")  # CPU as specified
# Optimizations
torch.set_num_threads(torch.get_num_threads())  # Default

def generate_image(prompt: str) -> bytes:
    # Add system prompt
    full_prompt = f"{prompt}, high quality, detailed, realistic"
    
    image = pipe(
        prompt=full_prompt,
        num_inference_steps=4,
        guidance_scale=0.0,
        width=1024,
        height=1024
    ).images[0]
    
    buffered = BytesIO()
    image.save(buffered, format="PNG")
    logger.info("Image generated")
    return buffered.getvalue()