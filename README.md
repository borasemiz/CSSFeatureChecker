# CSS Feature Scraper

An AWS Lambda function that scrapes [caniuse.com](https://caniuse.com) for CSS feature browser coverage, stores results in S3, and sends an email notification whenever a feature newly crosses 90% global coverage.

## How it works

1. **Scrape** — fetches the latest caniuse dataset from the canonical JSON source
2. **Compare** — downloads the previous `output.csv` from S3 and finds features that newly crossed 90% coverage since the last run
3. **Store** — uploads a fresh `output.csv` (all CSS features with their coverage) back to S3
4. **Notify** — sends an SMTP email listing any newly-above-90% features

On the first run there is no previous CSV to compare against, so the email step is skipped and only the CSV is written.

## Packages

| Package | Responsibility |
|---------|----------------|
| `scraper` | Fetches and parses the caniuse JSON, filters CSS features |
| `output` | Writes results to CSV format |
| `compare` | Diffs old vs. new coverage to find newly-crossed features |
| `storage` | Reads and writes `output.csv` on S3 |
| `email` | Sends a summary email over SMTP with STARTTLS |

## Environment variables

| Variable | Description |
|----------|-------------|
| `BUCKET_NAME` | S3 bucket where `output.csv` is stored |
| `AWS_ACCESS_KEY_ID` | AWS credentials |
| `AWS_SECRET_ACCESS_KEY` | AWS credentials |
| `AWS_REGION` | AWS region (e.g. `eu-west-1`) |
| `SMTP_HOST` | SMTP server hostname (e.g. `smtp.gmail.com`) |
| `SMTP_PORT` | SMTP port — `587` for STARTTLS |
| `SMTP_USERNAME` | SMTP login username |
| `SMTP_PASSWORD` | SMTP login password or app password |
| `SMTP_EMAIL_FROM` | Sender address |
| `EMAIL_TO` | Recipient address |

For local development, create a `.env` file at the project root — it is loaded automatically via `godotenv`.

## IAM permissions

The IAM user or role running the Lambda needs the following S3 permissions on the target bucket:

```json
{
  "Effect": "Allow",
  "Action": ["s3:GetObject", "s3:PutObject"],
  "Resource": "arn:aws:s3:::YOUR_BUCKET_NAME/*"
}
```

## Running tests

```bash
go test ./...
```

The integration test in `main_test.go` calls the real Lambda handler and requires a populated `.env` file with valid credentials.

## Building for Lambda

```bash
GOOS=linux GOARCH=arm64 go build -o bootstrap main.go
zip function.zip bootstrap
```
