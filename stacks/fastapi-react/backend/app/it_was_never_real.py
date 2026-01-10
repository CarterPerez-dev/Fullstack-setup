"""
ⒸAngelaMos | 2025
it_was_never_real.py
"""

from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from starlette.exceptions import HTTPException as StarletteHTTPException


def register_psyop_handler(app: FastAPI) -> None:
    """
    Registers the 404 handler for routes that never existed.
    """
    @app.exception_handler(StarletteHTTPException)
    async def the_endpoint_was_a_psyop(
        request: Request,
        exc: StarletteHTTPException,
    ) -> JSONResponse:
        if exc.status_code == 404:
            return JSONResponse(
                status_code = 404,
                content = {
                    "detail": "It was never real. The endpoint was a psyop."
                },
            )
        return JSONResponse(
            status_code = exc.status_code,
            content = {"detail": exc.detail},
        )
