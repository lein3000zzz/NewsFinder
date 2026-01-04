# [NewsFinder](https://t.me/crypto_NewsFinderBot)
___

NewsFinder is a Go-based application designed to process and analyze news events, (particularly those related to cryptocurrency markets in case of this exact repo.
It consumes raw news data from Kafka, performs hard and soft deduplication, natural language processing (NLP) using a CryptoBERT model, and tag detection for relevant entities. 
Analyzed data is stored in a PostgreSQL database with vector embeddings for similarity searches and produced back to Kafka for downstream consumption. 

- Stars are very much appreciated!
- Contributions are welcome! Feel free to open issues or submit pull requests.
- Forks are even more welcome as the projects can be used in various ways and adapted to different needs, even
  by simply replacing the BERT model with a different one and adjusting the analysis result structure.
    - That's basically why [LimitFinder](https://github.com/lein3000zzz/LimitFinder-whitepaper) is not open-sourced - it's a more specific use case.

#### Related topics
1. [News Producer Bun template](https://github.com/lein3000zzz/NFProducer-template-bun)
    - [Direct Bitget example based on the template](https://github.com/lein3000zzz/NFProducer-bitget)
2. [NewsAnalyzed Consumer Bun template](https://github.com/lein3000zzz/NFConsumer-template-bun)
3. [The telegram Bot](https://t.me/crypto_NewsFinderBot)
   - Available commands as of now: `/switch`, `/add`, `/latest` and `/start` 

Already implemented features:

- Event Consumption and Production: Integrates with Apache Kafka for ingesting news events and outputting analyzed results.
    - Protobuf is used for message serialization to ensure efficient data exchange.
  
- Content Normalization: Cleans and normalizes news content for consistent processing.

- Parallel Processing: Utilizes Go's concurrency features to handle multiple news events simultaneously for improved throughput.
  - Processing is still sequential within each datasource to maintain order and prevent race conditions.

- Deduplication: Implements hard (exact hash match) and soft (semantic similarity via vector embeddings) deduplication to avoid processing duplicates.
  - Uses pgvector extension in PostgreSQL for efficient vector similarity searches.
  - Uses all-minilm-L6-v2 model from SentenceTransformers (in particular, [my fork](https://github.com/lein3000zzz/all-minilm-l6-v2-go) of its [go implementation by clems4ever](https://github.com/clems4ever/all-minilm-l6-v2-go))) for generating embeddings.

- NLP Analysis: Utilizes a quantized [CryptoBERT model](https://huggingface.co/kk08/CryptoBERT) via ONNX Runtime to classify sentiment (bearish/bullish) on news content.

- Tag Detection: Identifies cryptocurrency-related tags (e.g., symbols, assets) by fetching and caching data from exchange APIs (Currently - Binance, Bitget).

- Data Storage: Leverages PostgreSQL with pgvector extension for efficient storage and querying of news with embeddings.

- Secret Management: Uses HashiCorp Vault for secure storage of some configuration params.

Upcoming features:

- User-friendly interface for adding sources as currently they are manually added via DB inserts.

- Enhanced Tag Detection: Expand tag detection capabilities to include more exchanges and asset types.
  - Probably make some changes to the detector logic overall.

- Proper metrics collection and monitoring.

- Modular architecture improvements for easier extensibility and automatic data updates.

- Finish configuration via Vault for all sensitive parameters.

- Comprehensive testing suite for all components.

- Other improvements that I'm not yet sure about :)

### Architecture

- Communicator: Handles Kafka interactions for consuming NewsEvent messages and producing NewsAnalyzed messages.
  - 2 kafka clients are used - one for consuming and one for producing.

- Analyzer: Orchestrates NLP and tag detection on normalized content.
  - NLP is done via ONNX Runtime with a quantized CryptoBERT model.
  - Tag detection fetches data from external exchange APIs and caches it.

- DataManager: Manages database operations, including lookups and insertions.

- Dedup: Performs hard (by hash) and soft (by embeddings) deduplication checks.

- App: Coordinates workers for concurrent message processing.

Data flow:

- Raw news events are consumed from Kafka and given the ingested_at timestamp.

- Deduplication checks are performed.
  - Currently, if a hard dedup is found, the message is skipped.
  - If a soft dedup is found, the record is still stored.
    - This is configurable.
    - This is opinionated.
    - May be changed in the future.

- Content is analyzed for sentiment and tags.

- Results are stored in the database and produced back to Kafka.

### Prerequisites

- Docker and Docker Compose

- Go 1.25.1 or later for development

- Access to external APIs (Binance, Bitget) for tag data

### Installation

- Clone the repository:

```Bash
git clone https://github.com/lein3000zzz/NewsFinder.git
cd NewsFinder
```

- Ensure Docker and Docker Compose are installed.

- Ensure you have Vault initialized.

Start the services:
```Bash
docker-compose up -d

This will launch the application, PostgreSQL database, Kafka broker, Kafka UI, and Vault.
```

### Usage

- Once deployed, the application automatically starts consuming messages from the newsevents topic and producing to newsanalyzed.
  - Configurable.

- Monitor logs via Docker Compose: docker-compose logs nfapp.
  - Probably gonna add loki support later.

- View Kafka topics via Kafka UI at http://localhost:8080.

- Query the database directly for stored news.

### Configuration
```
POSTGRES_USER: Username for PostgreSQL database connection.

POSTGRES_PASSWORD: Password for PostgreSQL database connection.

POSTGRES_DB: Name of the PostgreSQL database.

PG_DSN: Full Data Source Name (DSN) string for PostgreSQL connection, including host, user, password, database, port, and SSL mode.

ONNX_PATH: Path to the ONNX Runtime library file (platform-specific, e.g., DLL for Windows or SO for Linux).

MODEL_TOKENIZER_PATH: Path to the tokenizer JSON file for the CryptoBERT model.

MODEL_ONNX_PATH: Path to the quantized ONNX model file for CryptoBERT.

KAFKA_ADDR: Address of the Kafka broker.

KAFKA_CONSUMER_GROUP: Consumer group ID for Kafka message consumption.

KAFKA_CONSUMER_TOPIC: Kafka topic from which raw news events are consumed.

KAFKA_PRODUCER_TOPIC: Kafka topic to which analyzed news results are produced.

KAFKA_USERNAME: Username for Kafka authentication.

KAFKA_PASSWORD: Password for Kafka authentication.

VAULT_ADDR: Address of the HashiCorp Vault server.

VAULT_TOKEN: Authentication token for Vault access.

VAULT_KEYS: Comma-separated list of Vault unseal keys.
```

- In the vault, there are currently only 2 (envs are to be moved) urls for tag detectors stored:
  - `BINANCE_URL` - for Binance API base URL
  - `BITGET_URL` - for Bitget API base URL

- Adjust these as needed for your environment. 

### License

This project is licensed under the terms specified in LICENSE.