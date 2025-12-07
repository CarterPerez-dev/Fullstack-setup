"""
ⒸAngelaMos | 2025
main.py
"""
import uvicorn

from src.config import settings
from src.factory import create_app


app = create_app()

if __name__ == "__main__":
    uvicorn.run(
        "src.main:app",
        host = settings.HOST,
        port = settings.PORT,
        reload = settings.RELOAD,
    )
