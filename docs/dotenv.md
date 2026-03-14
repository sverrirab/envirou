# Loading .env files

Many projects use `.env` files to store configuration. Envirou can load these directly into your shell.

## Basic usage

Load the default `.env` file from the current directory:

```bash
ev dotenv
```

Load a specific file:

```bash
ev dotenv .env.local
```

## Layering multiple files

You can load multiple files in one command. Files are processed in order — later values override earlier ones:

```bash
ev dotenv .env .env.local
```

This is the same pattern used by Docker Compose and most dotenv libraries: base configuration in `.env`, local overrides in `.env.local`.

## Supported syntax

```bash
# Comments are ignored
KEY=value
QUOTED="hello world"
SINGLE_QUOTED='hello world'
export EXPORTED=value
EMPTY_VALUE=

# Lines without = are ignored
```

## Example: Project-specific environments

A typical project might have:

**.env** (committed to git):
```bash
DATABASE_HOST=localhost
DATABASE_PORT=5432
API_URL=https://api.dev.example.com
```

**.env.local** (in .gitignore):
```bash
DATABASE_PASSWORD=mysecret
API_URL=https://api.staging.example.com
```

Load both:
```bash
cd myproject
ev dotenv .env .env.local
```

The `API_URL` from `.env.local` overrides the one in `.env`. All variables are exported into your current shell.

## Combining with profiles

You can use `ev dotenv` alongside profiles. For example, load project-specific variables from `.env` and then activate an AWS profile:

```bash
ev dotenv .env
ev set awsprod
```
