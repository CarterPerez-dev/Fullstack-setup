==================
My fullstack project set up for:
Fastapi - Nginx - Docker - React TS - SCSS
==================
----

Setup
=====

Option 1: GitHub Template (Recommended)
---------------------------------------

Click the **"Use this template"** button above, or:

.. code-block:: bash

   gh repo create my-project --template CarterPerez-dev/fullstack-template
   cd my-project

Option 2: Clone
---------------

.. code-block:: bash

   git clone https://github.com/CarterPerez-dev/fullstack-template.git my-project
   cd my-project
   rm -rf .git && git init

Prerequisites (Optional but Recommended)
----------------------------------------

Install `just <https://github.com/casey/just>`_ command runner:

.. code-block:: bash

   curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/bin

Then add ``~/bin`` to your PATH if not already.

Installation
------------

.. code-block:: bash

   chmod +x setup.sh
   ./setup.sh

Or if you have just:

.. code-block:: bash

   just setup

This will:

- Copy ``.env.example`` → ``.env`` with generated ``SECRET_KEY``
- Move template files (LICENSE, CONTRIBUTING, etc.) to root
- Install backend dependencies (uv sync)
- Install frontend dependencies (pnpm install)

Next Steps
----------

1. Edit ``.env`` with your configuration
2. Start development: ``just dev-up``
3. After creating models: ``just migration-local "initial"`` then ``just migrate-local head``

Run ``just`` to see all available commands.

----

Documentation
=============

Coming soon...
----

Contributing
============

See `CONTRIBUTING.rst <CONTRIBUTING.rst>`_ for development setup and guidelines.

----

License
=======

MIT License - see `LICENSE <LICENSE>`_ for details.

----

Security
========

See `SECURITY.rst <SECURITY.rst>`_ for our security policy and how to report vulnerabilities.
