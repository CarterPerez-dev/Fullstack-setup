"""
ⒸAngelaMos | 2025
user.py
"""

from uuid import UUID
from sqlalchemy.ext.asyncio import (
    AsyncSession,
)

from src.core.exceptions import (
    EmailAlreadyExists,
    InvalidCredentials,
    UserNotFound,
)
from src.core.security import (
    hash_password,
    verify_password,
)
from src.schemas.user import (
    AdminUserCreate,
    UserCreate,
    UserListResponse,
    UserResponse,
    UserUpdate,
    UserUpdateAdmin,
)
from src.models.User import User
from src.repositories.user import UserRepository


class UserService:
    """
    Business logic for user operations
    """
    @staticmethod
    async def create_user(
        session: AsyncSession,
        user_data: UserCreate,
    ) -> UserResponse:
        """
        Register a new user
        """
        if await UserRepository.email_exists(session, user_data.email):
            raise EmailAlreadyExists(user_data.email)

        hashed = await hash_password(user_data.password)
        user = await UserRepository.create_user(
            session,
            email = user_data.email,
            hashed_password = hashed,
            full_name = user_data.full_name,
        )
        return UserResponse.model_validate(user)

    @staticmethod
    async def get_user_by_id(
        session: AsyncSession,
        user_id: UUID,
    ) -> UserResponse:
        """
        Get user by ID
        """
        user = await UserRepository.get_by_id(session, user_id)
        if not user:
            raise UserNotFound(str(user_id))
        return UserResponse.model_validate(user)

    @staticmethod
    async def get_user_model_by_id(
        session: AsyncSession,
        user_id: UUID,
    ) -> User:
        """
        Get user model by ID (for internal use)
        """
        user = await UserRepository.get_by_id(session, user_id)
        if not user:
            raise UserNotFound(str(user_id))
        return user

    @staticmethod
    async def update_user(
        session: AsyncSession,
        user: User,
        user_data: UserUpdate,
    ) -> UserResponse:
        """
        Update user profile
        """
        update_dict = user_data.model_dump(exclude_unset = True)
        updated_user = await UserRepository.update(
            session,
            user,
            **update_dict
        )
        return UserResponse.model_validate(updated_user)

    @staticmethod
    async def change_password(
        session: AsyncSession,
        user: User,
        current_password: str,
        new_password: str,
    ) -> None:
        """
        Change user password
        """
        is_valid, _ = await verify_password(current_password, user.hashed_password)
        if not is_valid:
            raise InvalidCredentials()

        hashed = await hash_password(new_password)
        await UserRepository.update_password(session, user, hashed)

    @staticmethod
    async def deactivate_user(
        session: AsyncSession,
        user: User,
    ) -> UserResponse:
        """
        Deactivate user account
        """
        updated = await UserRepository.update(
            session,
            user,
            is_active = False
        )
        return UserResponse.model_validate(updated)

    @staticmethod
    async def list_users(
        session: AsyncSession,
        page: int,
        size: int,
    ) -> UserListResponse:
        """
        List users with pagination
        """
        skip = (page - 1) * size
        users = await UserRepository.get_multi(
            session,
            skip = skip,
            limit = size
        )
        total = await UserRepository.count(session)
        return UserListResponse(
            items = [UserResponse.model_validate(u) for u in users],
            total = total,
            page = page,
            size = size,
        )

    @staticmethod
    async def admin_create_user(
        session: AsyncSession,
        user_data: AdminUserCreate,
    ) -> UserResponse:
        """
        Admin creates a new user
        """
        if await UserRepository.email_exists(session, user_data.email):
            raise EmailAlreadyExists(user_data.email)

        hashed = await hash_password(user_data.password)
        user = await UserRepository.create(
            session,
            email = user_data.email,
            hashed_password = hashed,
            full_name = user_data.full_name,
            role = user_data.role,
            is_active = user_data.is_active,
            is_verified = user_data.is_verified,
        )
        return UserResponse.model_validate(user)

    @staticmethod
    async def admin_update_user(
        session: AsyncSession,
        user_id: UUID,
        user_data: UserUpdateAdmin,
    ) -> UserResponse:
        """
        Admin updates a user
        """
        user = await UserRepository.get_by_id(session, user_id)
        if not user:
            raise UserNotFound(str(user_id))

        update_dict = user_data.model_dump(exclude_unset = True)

        if "email" in update_dict:
            existing = await UserRepository.get_by_email(
                session,
                update_dict["email"]
            )
            if existing and existing.id != user_id:
                raise EmailAlreadyExists(update_dict["email"])

        updated_user = await UserRepository.update(
            session,
            user,
            **update_dict
        )
        return UserResponse.model_validate(updated_user)

    @staticmethod
    async def admin_delete_user(
        session: AsyncSession,
        user_id: UUID,
    ) -> None:
        """
        Admin deletes a user (hard delete)
        """
        user = await UserRepository.get_by_id(session, user_id)
        if not user:
            raise UserNotFound(str(user_id))

        await UserRepository.delete(session, user)
