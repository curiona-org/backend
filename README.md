# Roadmap Generator Backend

This repository contains the REST API backend for the Roadmap Generator thesis project. Built primarily with Golang and PostgreSQL. Adopts a monolithic and kind of clean architecture approach.

## Overview

The server is responsible for the following functions:

- User authentication with JWT and Google OAuth
- Roadmap CRUD operations
- Integration with OpenAI / DeepSeek API for generating roadmap content
- Crawling the internet to find relevant resources for a given roadmap (Youtube, Books, Articles, etc)

Technologies used:

- **Language:** Golang
- **Database:** PostgreSQL
- **Cache:** Redis / No-cache
- **Worker:** Asynq
- **Auth:** JWT / Google OAuth
- **LLM API:** OpenAI / DeepSeek / Groq
- **Tracing:** Jaeger & OpenTelemetry
- **Load Testing:** K6


## Running Locally

Copy the `.env.example` file to `.env` and fill in the required environment variables.

Use docker-compose to run the services locally. Make sure you have Docker and Docker Compose installed.

```bash
make docker-compose
```

Run the migrations to create the database schema.

```bash
make migrate
```

Start the server.

```bash
make run
```
