====================================
Contributing to XYZ
====================================

Thank you for your interest in contributing to **XYZ**!

----

Development Setup
=================

1. Clone the repository::

    git clone https://github.com/CarterPerez-dev/XYZ.git
    cd stripe-referral

2. Create a virtual environment::

    python -m venv venv
    source venv/bin/activate  # On Windows: venv\Scripts\activate

3. Install in development mode::

    pip install -e ".[dev]"

4. Set up your environment::

    cp .env.example .env
    # Edit .env with your database URL and API keys

5. Run database migrations::

    alembic upgrade head

----

Code Style
==========

This project follows strict coding standards:

Formatting
----------

- **Formatter**: Use ``yapf`` for code formatting
- **Linter**: Use ``ruff`` for linting only (not formatting)
- **Type hints**: Full type hints everywhere using modern syntax (``str | None``, ``list[str]``)
- **Docstrings**: Vertical multi-line format only
- **Imports**: Vertical multi-line for 2+ imports with trailing commas
- **Constants**: All magic numbers and strings must be constants or enums

Run formatters and linters
---------------------------

::

    # Format code
    yapf -ir src/

    # Lint
    ruff check src/

    # Type check
    mypy src/

----

Testing
=======

Run tests before submitting::

    pytest -v

----

Pull Request Process
====================

1. Create a feature branch from ``main``
2. Make your changes following code style guidelines
3. Add tests for new functionality
4. Ensure all tests pass
5. Update documentation if needed
6. Submit a pull request

----

Commit Messages
===============

Use conventional commit format:

- ``feat:`` New features
- ``fix:`` Bug fixes
- ``docs:`` Documentation changes
- ``refactor:`` Code refactoring
- ``test:`` Test additions/changes
- ``chore:`` Maintenance tasks

Example::

    feat: XYZ

    - XYZ
    - XYZ
    - XYZ

----

Questions?
==========

Open an issue or reach out to the maintainers!
