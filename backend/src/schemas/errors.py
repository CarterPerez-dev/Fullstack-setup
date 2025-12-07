"""
ⒸAngelaMos | 2025
errors.py
"""

from pydantic import Field
from src.schemas.base import BaseSchema


class ErrorDetail(BaseSchema):
    """
    Standard error response format
    """
    detail: str = Field(..., description = "Human readable error message")
    type: str = Field(..., description = "Exception class name")

    model_config = {
        "json_schema_extra": {
            "examples": [
                {
                    "detail": "User with id '123' not found",
                    "type": "UserNotFound"
                }
            ]
        }
    }
