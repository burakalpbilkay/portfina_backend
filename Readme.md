# Finance Platform Backend (Go Service)

This service provides API endpoints to:
- Trigger Airflow DAGs
- Check DAG status
- Retrieve bond results & cashflows
- Upload CSV files for ingestion DAGs

## 🛠 Usage

### Build & Run (Docker Compose)
```bash
docker-compose up --build
```

Backend runs on: http://localhost:8081

### API Endpoints

#### ✅ DAG Status
- **GET** `/status/bond_enrichment_dag`
- **GET** `/status/{dag_id}`

#### ✅ Trigger DAG
- **POST** `/trigger/bond_enrichment_dag`

#### ✅ Bond Data
- **GET** `/bond_results`
- **GET** `/bond_cashflows`

#### ✅ CSV Uploads
- **POST** `/upload/{key}`
  - key: bond, interest_rate, inflation_index, inflation_expectation, exchange_rate, foreign_exchange, forward_curve

Example Upload:
```bash
curl -X POST http://localhost:8081/upload/bond \
    -F "file=@/path/to/Bond.csv"
```

## 📂 Folder Structure
- `main.go` → Starts HTTP server & routes
- `handlers/` → API endpoint handlers
- `utils/` → DB connection helpers
- `Dockerfile` → Build Go service
- `docker-compose.yml` → Compose services

## 🔗 Prerequisites
- Airflow accessible at `http://host.docker.internal:8080`
- PostgreSQL running with `finance_db`

