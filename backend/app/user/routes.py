"""
ⒸAngelaMos | 2025
routes.py
"""

from uuid import UUID

from fastapi import (
    APIRouter,
    status,
)

from core.dependencies import (
    CurrentUser,
    DBSession,
)
from core.responses import (
    AUTH_401,
    CONFLICT_409,
    NOT_FOUND_404,
)
from .schemas import (
    UserCreate,
    UserResponse,
    UserUpdate,
)
from .service import UserService


router = APIRouter(prefix = "/users", tags = ["users"])


@router.post(
    "",
    response_model = UserResponse,
    status_code = status.HTTP_201_CREATED,
    responses = {**CONFLICT_409},
)
async def create_user(
    db: DBSession,
    user_data: UserCreate,
) -> UserResponse:
    """
    Register a new user
    """
    return await UserService.create_user(db, user_data)


@router.get(
    "/{user_id}",
    response_model = UserResponse,
    responses = {
        **AUTH_401,
        **NOT_FOUND_404
    },
)
async def get_user(
    db: DBSession,
    user_id: UUID,
    _: CurrentUser,
) -> UserResponse:
    """
    Get user by ID
    """
    return await UserService.get_user_by_id(db, user_id)


@router.patch(
    "/me",
    response_model = UserResponse,
    responses = {**AUTH_401},
)
async def update_current_user(
    db: DBSession,
    current_user: CurrentUser,
    user_data: UserUpdate,
) -> UserResponse:
    """
    Update current user profile
    """
    return await UserService.update_user(db, current_user, user_data)
