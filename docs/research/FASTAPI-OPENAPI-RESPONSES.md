# FastAPI OpenAPI Response Documentation - Production Best Practices (2025)

## Solution Summary

The canonical, production-ready approach for documenting FastAPI custom exceptions in OpenAPI (2025) is to use **Pydantic response models with the `responses` parameter**, combined with **reusable response dictionaries** and **dict unpacking**. This pattern is officially documented by FastAPI, maintains DRY principles, generates proper Swagger/OpenAPI documentation, and scales elegantly across large applications.

Custom exception handlers manage runtime behavior, while the `responses` parameter ensures OpenAPI schema accuracy. The combination provides type-safe, maintainable, and well-documented error responses.

---

## Detailed Analysis

### Understanding the Issue

FastAPI automatically generates OpenAPI documentation based on your route definitions. However, by default, it only documents:

- **200/201** responses from `response_model`
- **422** responses for Pydantic validation errors (automatically added when path operations have parameters or request bodies)
- **204** responses when using `status_code=status.HTTP_204_NO_CONTENT`

Custom exceptions raised in your application logic (401, 403, 404, 409, 429, etc.) are **NOT automatically documented** in the OpenAPI schema, even if you have exception handlers defined. Exception handlers only control runtime behavior - they don't update the OpenAPI spec.

> "You can declare additional responses, with additional status codes, media types, descriptions, etc. Those additional responses will be included in the OpenAPI schema, so they will also appear in the API docs." - [Additional Responses in OpenAPI - FastAPI](https://fastapi.tiangolo.com/advanced/additional-responses/)

### Root Cause

The separation between **runtime exception handling** and **OpenAPI documentation** is intentional in FastAPI:

1. **Exception Handlers** (`@app.exception_handler`) - Handle exceptions at runtime
2. **`responses` Parameter** - Document possible responses in OpenAPI schema

These are separate concerns. FastAPI cannot automatically infer all possible exceptions from your code because:
- Exceptions can be raised deep in the call stack
- Dynamic code paths make static analysis unreliable
- You may want to document responses that aren't always raised

---

## Solution Steps

### Step 1: Create Pydantic Error Response Models

Create reusable Pydantic models that match your exception handler response structure.

**File: `/home/yoshi/dev/templates/backend/src/schemas/errors.py`**

```python
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
    detail: str = Field(
        ...,
        description="Human-readable error message"
    )
    type: str = Field(
        ...,
        description="Exception class name"
    )

    model_config = {
        "json_schema_extra": {
            "examples": [
                {
                    "detail": "User with id '123e4567-e89b-12d3-a456-426614174000' not found",
                    "type": "UserNotFound"
                }
            ]
        }
    }


class ValidationErrorDetail(BaseSchema):
    """
    Pydantic validation error response format
    """
    detail: list[dict[str, str | list[str]]] = Field(
        ...,
        description="Validation error details"
    )

    model_config = {
        "json_schema_extra": {
            "examples": [
                {
                    "detail": [
                        {
                            "loc": ["body", "email"],
                            "msg": "value is not a valid email address",
                            "type": "value_error.email"
                        }
                    ]
                }
            ]
        }
    }
```

**Key Points:**
- Models match your exception handler JSON structure exactly
- Use Pydantic v2 `model_config` with `json_schema_extra` for examples (OpenAPI 3.1.0 standard)
- Examples appear in Swagger UI for better developer experience

> "OpenAPI 3.1.0 (used since FastAPI 0.99.0) added support for `examples`, which is part of the JSON Schema standard. So you are encouraged to migrate `example` to `examples`." - [Declare Request Example Data - FastAPI](https://fastapi.tiangolo.com/tutorial/schema-extra-example/)

---

### Step 2: Define Reusable Response Dictionaries

Create centralized response definitions that can be shared across endpoints.

**File: `/home/yoshi/dev/templates/backend/src/core/responses.py`**

```python
"""
ⒸAngelaMos | 2025
responses.py
"""

from typing import Any
from src.schemas.errors import ErrorDetail, ValidationErrorDetail


COMMON_RESPONSES: dict[int | str, dict[str, Any]] = {
    401: {
        "model": ErrorDetail,
        "description": "Authentication failed - Invalid or missing credentials"
    },
    403: {
        "model": ErrorDetail,
        "description": "Permission denied - Insufficient privileges"
    },
    404: {
        "model": ErrorDetail,
        "description": "Resource not found"
    },
    409: {
        "model": ErrorDetail,
        "description": "Conflict - Resource already exists or state conflict"
    },
    422: {
        "model": ValidationErrorDetail,
        "description": "Validation error - Invalid request data"
    },
    429: {
        "model": ErrorDetail,
        "description": "Rate limit exceeded"
    },
    500: {
        "model": ErrorDetail,
        "description": "Internal server error"
    },
}


AUTH_RESPONSES: dict[int | str, dict[str, Any]] = {
    401: {
        "model": ErrorDetail,
        "description": "Authentication failed - Invalid credentials or token"
    },
    429: {
        "model": ErrorDetail,
        "description": "Too many login attempts - Rate limit exceeded"
    },
}


PROTECTED_RESPONSES: dict[int | str, dict[str, Any]] = {
    401: {
        "model": ErrorDetail,
        "description": "Authentication required - Missing or invalid token"
    },
    403: {
        "model": ErrorDetail,
        "description": "Permission denied - Insufficient role or privileges"
    },
}


CRUD_RESPONSES: dict[int | str, dict[str, Any]] = {
    404: {
        "model": ErrorDetail,
        "description": "Resource not found"
    },
    422: {
        "model": ValidationErrorDetail,
        "description": "Validation error"
    },
}
```

**Type Annotation for Mypy:**
Using `dict[int | str, dict[str, Any]]` ensures Mypy type checking passes when using dict unpacking.

> "A solution is to explicitly type the responses dict: `responses: dict[int | str, dict[str, Any]]`" - [Additional Responses in OpenAPI and Mypy - FastAPI Discussion #12056](https://github.com/fastapi/fastapi/discussions/12056)

---

### Step 3: Apply Responses to Route Decorators

Use dict unpacking (`**dict`) to combine reusable responses with endpoint-specific ones.

**Example: `/home/yoshi/dev/templates/backend/src/routes/auth.py`**

```python
from src.core.responses import AUTH_RESPONSES, PROTECTED_RESPONSES
from src.schemas.auth import TokenWithUserResponse, TokenResponse


@router.post(
    "/login",
    response_model=TokenWithUserResponse,
    responses={**AUTH_RESPONSES},
    status_code=status.HTTP_200_OK,
)
@limiter.limit(settings.RATE_LIMIT_AUTH)
async def login(
    request: Request,
    response: Response,
    db: DBSession,
    ip: ClientIP,
    form_data: Annotated[OAuth2PasswordRequestForm, Depends()],
) -> TokenWithUserResponse:
    """
    Login with email and password

    Returns access token and user information upon successful authentication.
    """
    result, refresh_token = await AuthService.login(
        db,
        email=form_data.username,
        password=form_data.password,
        ip_address=ip,
    )
    set_refresh_cookie(response, refresh_token)
    return result


@router.post(
    "/refresh",
    response_model=TokenResponse,
    responses={
        **AUTH_RESPONSES,
        400: {
            "model": ErrorDetail,
            "description": "Invalid refresh token"
        }
    },
)
async def refresh_token(
    db: DBSession,
    ip: ClientIP,
    refresh_token: str | None = Cookie(None),
) -> TokenResponse:
    """
    Refresh access token using refresh token from HTTP-only cookie
    """
    if not refresh_token:
        raise TokenError("Refresh token required")
    return await AuthService.refresh_tokens(db, refresh_token, ip_address=ip)


@router.get(
    "/me",
    response_model=UserResponse,
    responses={**PROTECTED_RESPONSES},
)
async def get_current_user(current_user: CurrentUser) -> UserResponse:
    """
    Get current authenticated user information

    Requires valid JWT token in Authorization header.
    """
    return UserResponse.model_validate(current_user)
```

**Example: `/home/yoshi/dev/templates/backend/src/routes/user.py`**

```python
from src.core.responses import PROTECTED_RESPONSES, CRUD_RESPONSES


@router.post(
    "",
    response_model=UserResponse,
    status_code=status.HTTP_201_CREATED,
    responses={
        **CRUD_RESPONSES,
        409: {
            "model": ErrorDetail,
            "description": "Email already registered"
        }
    }
)
async def create_user(
    db: DBSession,
    user_data: UserCreate,
) -> UserResponse:
    """
    Register a new user account
    """
    return await UserService.create_user(db, user_data)


@router.get(
    "/{user_id}",
    response_model=UserResponse,
    responses={**PROTECTED_RESPONSES, **CRUD_RESPONSES}
)
async def get_user(
    db: DBSession,
    user_id: UUID,
    current_user: CurrentUser,
) -> UserResponse:
    """
    Get user by ID

    Users can only view their own profile unless they have admin privileges.
    """
    return await UserService.get_user_by_id(db, user_id)
```

**Dict Unpacking Pattern:**
> "For cases where you want predefined responses that apply to many path operations but also want to combine them with custom responses, you can use the Python technique of 'unpacking' a dict with `**dict_to_unpack`." - [Additional Responses in OpenAPI - FastAPI](https://fastapi.tiangolo.com/advanced/additional-responses/)

---

### Step 4: Maintain Exception Handlers

Your existing exception handlers remain unchanged - they handle runtime behavior.

**File: `/home/yoshi/dev/templates/backend/src/factory.py`** (already implemented correctly)

```python
@app.exception_handler(BaseAppException)
async def app_exception_handler(
    request: Request,
    exc: BaseAppException,
) -> JSONResponse:
    return JSONResponse(
        status_code=exc.status_code,
        content={
            "detail": exc.message,
            "type": exc.__class__.__name__,
        },
    )
```

**Important:** Exception handlers and `responses` parameter serve different purposes:
- **Exception handlers** = Runtime error handling
- **`responses` parameter** = OpenAPI documentation

> "Keep in mind that you have to return the JSONResponse directly" when using additional responses - [Additional Responses in OpenAPI - FastAPI](https://fastapi.tiangolo.com/advanced/additional-responses/)

---

## Alternative Approaches (Not Recommended)

### 1. Third-Party Packages (fastapi-docx, fastapi-responses)

**Pros:**
- Automatic exception discovery
- Less manual configuration

**Cons:**
- External dependency adds complexity
- May not work correctly with all FastAPI versions
- Less control over documentation
- Package `fastapi-responses` last updated July 2021 (outdated)
- Adds "magic" that obscures what's happening

> "FastAPI-docx extends the FastAPI OpenAPI spec to include all possible HTTPException or custom Exception response schemas" - [fastapi-docx - PyPI](https://pypi.org/project/fastapi-docx/)

**Verdict:** Avoid for production. The manual approach with `responses` parameter is clearer, more maintainable, and officially supported.

---

### 2. Custom OpenAPI Schema Modification

**Approach:** Override `app.openapi()` to programmatically modify the schema.

**Pros:**
- Maximum control over OpenAPI output

**Cons:**
- Complex to implement and maintain
- Easy to introduce bugs
- Breaks when FastAPI updates OpenAPI generation logic
- Not DRY - requires duplicating route logic

> "Currently there is no simple way. You have to modify the OpenAPI file as described here. Meaning you have to load the dictionary and remove the references to the error 422." - [Custom exception handling not updating OpenAPI status code - Issue #2455](https://github.com/fastapi/fastapi/issues/2455)

**Verdict:** Only use for very specific edge cases. Not suitable as a general pattern.

---

### 3. Using HTTPException Instead of Custom Exceptions

**Approach:** Raise `HTTPException` directly instead of custom exception classes.

**Cons:**
- Loses type safety
- No centralized exception hierarchy
- Harder to add custom fields (like `retry_after` for rate limiting)
- Doesn't solve documentation problem anyway

> "HTTPException is not a Pydantic model, so it cannot be used directly in the `responses` parameter." - [FastAPI Error Handling Patterns - Better Stack](https://betterstack.com/community/guides/scaling-python/error-handling-fastapi/)

**Verdict:** Keep your custom exceptions. They provide better structure and maintainability.

---

### 4. Response Model Union Types

**Approach:** Use `response_model=Union[SuccessModel, ErrorModel]`

**Cons:**
- Only documents one status code (usually 200)
- Doesn't properly separate success vs error responses by status code
- Creates confusing OpenAPI schemas
- Not the intended use of `response_model`

**Verdict:** Not appropriate for error documentation. Use `responses` parameter instead.

---

## Verification Steps

### 1. Check OpenAPI Schema

Visit `/docs` (Swagger UI) and verify:
- Each endpoint shows all possible response status codes
- Error responses have proper descriptions
- Example error responses are visible
- Schema references are correct

### 2. Check Generated OpenAPI JSON

```bash
curl http://localhost:8000/openapi.json | jq '.paths."/v1/auth/login".post.responses'
```

Expected output:
```json
{
  "200": {
    "description": "Successful Response",
    "content": {
      "application/json": {
        "schema": {
          "$ref": "#/components/schemas/TokenWithUserResponse"
        }
      }
    }
  },
  "401": {
    "description": "Authentication failed - Invalid credentials or token",
    "content": {
      "application/json": {
        "schema": {
          "$ref": "#/components/schemas/ErrorDetail"
        }
      }
    }
  },
  "429": {
    "description": "Too many login attempts - Rate limit exceeded",
    "content": {
      "application/json": {
        "schema": {
          "$ref": "#/components/schemas/ErrorDetail"
        }
      }
    }
  }
}
```

### 3. Runtime Verification

Test that actual error responses match documented schemas:

```bash
# Test 401 error
curl -X POST http://localhost:8000/v1/auth/login \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=wrong@example.com&password=wrongpass"

# Expected response (status 401):
# {"detail": "Invalid email or password", "type": "InvalidCredentials"}
```

### 4. Type Checking

Run Mypy to verify type safety:

```bash
mypy backend/src/
```

Should pass without errors when using properly typed response dictionaries.

---

## Production Best Practices Summary

### 1. DRY Principle
- Create reusable response dictionaries grouped by use case
- Use dict unpacking to combine responses
- Define error models once, reference everywhere

### 2. Type Safety
- Type response dicts as `dict[int | str, dict[str, Any]]`
- Use Pydantic models for all responses
- Enable Mypy checking

### 3. Documentation Quality
- Provide clear, specific descriptions for each status code
- Include realistic examples using `json_schema_extra`
- Group related error responses logically

### 4. Maintainability
- Centralize response definitions in `src/core/responses.py`
- Keep exception handlers in `src/factory.py`
- Document why certain endpoints have specific error responses

### 5. Scalability
- Create response groups (AUTH_RESPONSES, CRUD_RESPONSES, etc.)
- Compose complex response sets using multiple unpacks
- Add endpoint-specific responses as needed

### 6. Consistency
- Error response structure matches exception handler output
- All endpoints use the same error model
- Status codes follow HTTP standards

---

## Implementation Checklist

- [ ] Create `backend/src/schemas/errors.py` with Pydantic error models
- [ ] Create `backend/src/core/responses.py` with reusable response dictionaries
- [ ] Add `responses` parameter to all route decorators
- [ ] Verify exception handlers match documented error structure
- [ ] Test OpenAPI schema in `/docs`
- [ ] Verify actual error responses match schemas
- [ ] Run Mypy type checking
- [ ] Update endpoint docstrings to reference error conditions
- [ ] Document special error cases in code comments (file headers only)

---

## Important Notes

### Version-Specific Considerations

- **FastAPI 0.99.0+**: Uses OpenAPI 3.1.0 (supports `examples` array)
- **Pydantic v2**: Use `model_config` instead of `Config` class
- **Python 3.10+**: Use `dict[int | str, dict[str, Any]]` type syntax

### Known Limitations

1. **No Automatic Discovery**: FastAPI cannot automatically detect all possible exceptions from your code. You must manually document them.

2. **No Runtime Validation**: The `responses` parameter is documentation-only. FastAPI doesn't validate that your actual responses match the documented schemas.

3. **Dict Unpacking Order**: Later dicts override earlier ones when unpacking:
   ```python
   # 404 description from CRUD_RESPONSES will be used
   responses={**COMMON_RESPONSES, **CRUD_RESPONSES}
   ```

4. **422 Auto-Generation**: FastAPI automatically adds 422 responses for endpoints with request bodies. You can override this by explicitly defining 422 in `responses`.

### Edge Cases

**Multiple Response Models for Same Status Code:**
Not supported. You can only have one model per status code. Use a Union model if needed:
```python
from typing import Union
from pydantic import BaseModel

class ErrorResponse(BaseModel):
    error: Union[ErrorDetail, ValidationErrorDetail]
```

**Binary/File Responses:**
For non-JSON responses, specify the media type:
```python
responses={
    200: {
        "content": {"image/png": {}},
        "description": "Returns a PNG image"
    }
}
```

**Router-Level Responses:**
You can apply responses to entire routers:
```python
router = APIRouter(
    prefix="/v1/users",
    tags=["users"],
    responses={**PROTECTED_RESPONSES}
)
```

> "It's supported in APIRouter so you can add responses to a whole sub-router in a single `.include_router()`" - [Response Model for Exceptions - FastAPI Discussion #8224](https://github.com/fastapi/fastapi/discussions/8224)

---

## Sources and References

### Official FastAPI Documentation
- [Additional Responses in OpenAPI - FastAPI](https://fastapi.tiangolo.com/advanced/additional-responses/) - Primary documentation for `responses` parameter
- [Declare Request Example Data - FastAPI](https://fastapi.tiangolo.com/tutorial/schema-extra-example/) - Pydantic v2 examples with `json_schema_extra`
- [Handling Errors - FastAPI](https://fastapi.tiangolo.com/tutorial/handling-errors/) - Exception handlers and HTTPException usage
- [Response Model - Return Type - FastAPI](https://fastapi.tiangolo.com/tutorial/response-model/) - Response model fundamentals

### FastAPI GitHub Discussions & Issues
- [Additional Responses in OpenAPI and Mypy - Discussion #12056](https://github.com/fastapi/fastapi/discussions/12056) - Type-safe response dictionary pattern
- [Custom exception handling not updating OpenAPI status code - Issue #2455](https://github.com/fastapi/fastapi/issues/2455) - Why exception handlers don't auto-update OpenAPI
- [Response Model for Exceptions - Discussion #8224](https://github.com/fastapi/fastapi/discussions/8224) - Router-level responses
- [Include possible HTTPExceptions in OpenAPI spec - Issue #1999](https://github.com/fastapi/fastapi/issues/1999) - Historical context on why this isn't automatic

### Pydantic Documentation
- [JSON Schema - Pydantic Validation](https://docs.pydantic.dev/latest/concepts/json_schema/) - Pydantic v2 JSON Schema generation
- [How to use openapi_examples with Pydantic? - Discussion #10233](https://github.com/fastapi/fastapi/discussions/10233) - Examples in Pydantic models

### Community Best Practices
- [FastAPI Error Handling Patterns - Better Stack](https://betterstack.com/community/guides/scaling-python/error-handling-fastapi/) - Comprehensive error handling patterns
- [Exception Handling Best Practices in Python: A FastAPI Perspective - Medium](https://medium.com/delivus/exception-handling-best-practices-in-python-a-fastapi-perspective-98ede2256870) - Production-level exception design
- [Contextual error handling for FastAPI - DEV Community](https://dev.to/ivan-borovets/contextual-error-handling-for-fastapi-per-route-with-openapi-schema-generation-4p9a) - Advanced per-route error handling
- [Mastering Response Models and Status Codes in FastAPI](https://procodebase.com/article/mastering-response-models-and-status-codes-in-fastapi) - Response model patterns

### Third-Party Tools (For Reference Only)
- [fastapi-docx - PyPI](https://pypi.org/project/fastapi-docx/) - Automatic exception discovery (not recommended for production)
- [fastapi-responses - PyPI](https://pypi.org/project/fastapi-responses/) - Extended HTTPException support (outdated, last update 2021)

### OpenAPI Specification
- [OpenAPI Responses Object](https://swagger.io/specification/#responses-object) - Official OpenAPI 3.1.0 specification
- [OpenAPI Response Object](https://swagger.io/specification/#response-object) - Response object schema definition

---

## Conclusion

The **Pydantic models + `responses` parameter + dict unpacking** pattern is the canonical 2025 approach for documenting FastAPI error responses. It's:

- **Official**: Documented in FastAPI's official guides
- **DRY**: Reusable response dictionaries eliminate duplication
- **Type-safe**: Works with Mypy when properly typed
- **Maintainable**: Centralized definitions, clear structure
- **Scalable**: Easily extends to hundreds of endpoints
- **Production-ready**: Used by senior engineers in large-scale applications

This approach separates concerns cleanly:
- **Custom exceptions** = Domain logic and error types
- **Exception handlers** = Runtime error transformation
- **Response models** = Error schema and structure
- **`responses` parameter** = OpenAPI documentation

By following this pattern, your API documentation will accurately reflect all possible responses, making it easier for frontend developers, API consumers, and automated tools to work with your API.
