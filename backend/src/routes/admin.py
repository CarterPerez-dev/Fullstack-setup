"""
ⒸAngelaMos | 2025
admin.py
"""

from uuid import UUID
from typing import Annotated

from fastapi import (
    APIRouter,
    Depends,
    Query,
    status,
)

from src.config import (
    settings,
    UserRole,
)
from src.core.dependencies import (
    DBSession,
    RequireRole,
)
from src.core.responses import (
    AUTH_401,
    CONFLICT_409,
    FORBIDDEN_403,
    NOT_FOUND_404,
)
from src.schemas.user import (
    AdminUserCreate,
    UserListResponse,
    UserResponse,
    UserUpdateAdmin,
)
from src.models.User import User
from src.services.user import UserService


router = APIRouter(prefix = "/admin", tags = ["admin"])

AdminOnly = Annotated[User, Depends(RequireRole(UserRole.ADMIN))]


@router.get(
    "/users",
    response_model = UserListResponse,
    responses = {
        **AUTH_401,
        **FORBIDDEN_403
    },
)
async def list_users(
    db: DBSession,
    _: AdminOnly,
    page: int = Query(default = 1,
                      ge = 1),
    size: int = Query(
        default = settings.PAGINATION_DEFAULT_SIZE,
        ge = 1,
        le = settings.PAGINATION_MAX_SIZE
    ),
) -> UserListResponse:
    """
    List all users (admin only)
    """
    return await UserService.list_users(db, page, size)


@router.post(
    "/users",
    response_model = UserResponse,
    status_code = status.HTTP_201_CREATED,
    responses = {
        **AUTH_401,
        **FORBIDDEN_403,
        **CONFLICT_409
    },
)
async def create_user(
    db: DBSession,
    _: AdminOnly,
    user_data: AdminUserCreate,
) -> UserResponse:
    """
    Create a new user (admin only, bypasses registration)
    """
    return await UserService.admin_create_user(db, user_data)


@router.get(
    "/users/{user_id}",
    response_model = UserResponse,
    responses = {
        **AUTH_401,
        **FORBIDDEN_403,
        **NOT_FOUND_404
    },
)
async def get_user(
    db: DBSession,
    _: AdminOnly,
    user_id: UUID,
) -> UserResponse:
    """
    Get user by ID (admin only)
    """
    return await UserService.get_user_by_id(db, user_id)


@router.patch(
    "/users/{user_id}",
    response_model = UserResponse,
    responses = {
        **AUTH_401,
        **FORBIDDEN_403,
        **NOT_FOUND_404,
        **CONFLICT_409
    },
)
async def update_user(
    db: DBSession,
    _: AdminOnly,
    user_id: UUID,
    user_data: UserUpdateAdmin,
) -> UserResponse:
    """
    Update user (admin only)
    """
    return await UserService.admin_update_user(db, user_id, user_data)


@router.delete(
    "/users/{user_id}",
    status_code = status.HTTP_204_NO_CONTENT,
    responses = {
        **AUTH_401,
        **FORBIDDEN_403,
        **NOT_FOUND_404
    },
)
async def delete_user(
    db: DBSession,
    _: AdminOnly,
    user_id: UUID,
) -> None:
    """
    Delete user (admin only, hard delete)
    """
    await UserService.admin_delete_user(db, user_id)
