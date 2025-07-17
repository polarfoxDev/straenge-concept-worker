# Concept worker for stränge.de riddle generation

This repository contains the **concept worker** microservice for generating riddles for [strangui](https://github.com/polarfoxDev/strangui).

## Overview

The concept worker is a Go-based service that automates the creation of riddle concepts. It interacts with Redis for job queue management and uses an AI backend (via OpenAI API) to generate creative riddle ideas, themes, and word pools. These concepts are then pushed to a Redis queue for further processing by other services.

## Features

- **Automated riddle concept generation** using AI
- **Queue management** via Redis
- **Configurable super solutions** for custom riddle themes
- **Logging** with logrus for easy debugging and monitoring

## How It Works

1. The worker monitors the Redis queue `generate-riddle`.
2. If the queue has fewer than 15 jobs, it generates new riddle concepts using AI.
3. Each concept includes a super solution, theme description, and word pool.
4. Concepts are packaged as jobs and pushed to the Redis queue.
5. The worker runs continuously, refilling the queue as needed.

## Project Structure

- [`main.go`](main.go): Main worker loop, Redis integration, configuration, and logging.
- [`concepts.go`](concepts.go): Logic for generating riddle concepts using the AI backend.
- `m/ai`: AI integration for generating super solutions, themes, and word pools.
- `m/models`: Data models for riddle concepts and jobs.

## Setup

### Prerequisites

- Go 1.20+
- Redis server
- OpenAI API key

### Installation

Clone the repository and install dependencies:

```bash
git clone https://github.com/polarfoxDev/straenge-concept-worker.git
cd straenge-concept-worker
go mod tidy
```

### Configuration

Create a `.env` file in the root directory with the following variables:

```env
REDIS_URL=localhost:6379
OPENAI_API_KEY=your-openai-api-key
LOG_LEVEL=info
PREDEFINED_SUPER_SOLUTIONS=solution1,solution2,solution3 # optional
```

### Running the Worker

Start the worker:

```bash
go run main.go
```

## Contributing

Any contributions you make are greatly appreciated.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue.
Don't forget to give the project a star! Thanks!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

Distributed under the **MIT** License. See [`LICENSE`](./LICENSE) for more information.