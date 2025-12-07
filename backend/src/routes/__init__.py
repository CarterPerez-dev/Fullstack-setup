"""
AngelaMos | 2025
__init__.py
"""

from src.routes.admin import router as admin_router
from src.routes.auth import router as auth_router
from src.routes.health import router as health_router
from src.routes.user import router as user_router


__all__ = [
    "admin_router",
    "auth_router",
    "health_router",
    "user_router",
]
