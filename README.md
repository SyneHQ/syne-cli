# Syne CLI

A command-line interface tool that provides PostgreSQL operations with authentication.

## Installation

### From GitHub Releases

1. Go to the [Releases](https://github.com/your-username/syne-cli/releases) page
2. Download the appropriate binary for your system:
   - For macOS: `syne-cli_Darwin_x86_64.tar.gz` (Intel) or `syne-cli_Darwin_arm64.tar.gz` (Apple Silicon)
   - For Linux: `syne-cli_Linux_x86_64.tar.gz` (64-bit) or `syne-cli_Linux_arm64.tar.gz` (ARM)
   - For Windows: `syne-cli_Windows_x86_64.zip`
3. Extract the binary to a location in your PATH

### From Source

```bash
go install github.com/your-username/syne-cli@latest
```

## Setup Authentication

1. Run the credentials setup script:
```bash
go run scripts/create_credentials.go <desired-username> <desired-password>
```
This will create a `.env` file with your hashed credentials.

## Commands

### Login
Test your authentication:
```bash
syne-cli login -u <username> -p <password>
```

### Copy Files to PostgreSQL
Copy files from a folder to a PostgreSQL table:
```bash
syne-cli copy <folder-path> <table-name> \
    -u <username> \
    -p <password> \
    --db-user <postgres-user> \
    --db-password <postgres-password> \
    --db-name <database-name> \
    [--host <host>] \
    [--port <port>] \
    [--ssl-mode <ssl-mode>]
```

#### Optional Flags
- `--host`: PostgreSQL host (default: "localhost")
- `--port`: PostgreSQL port (default: "5432")
- `--ssl-mode`: PostgreSQL SSL mode (default: "disable")

## Example Usage

1. First set up your CLI credentials:
```bash
go run scripts/create_credentials.go admin secretpassword
```

2. Test the login:
```bash
syne-cli login -u admin -p secretpassword
```

3. Copy CSV files from a folder to PostgreSQL (example using provided sample):
```bash
# The CLI will automatically create the table with appropriate column types:
syne-cli copy ./examples users \
    -u admin \
    -p secretpassword \
    --db-user postgres \
    --db-password dbpass \
    --db-name mydb
```

## Notes

- The copy command expects CSV files with headers in the specified folder
- The headers from the first row will be used as column names
- The data types will be automatically inferred from the second row
- Supported data types:
  - INTEGER for whole numbers
  - NUMERIC for decimal numbers
  - DATE for dates (YYYY-MM-DD format)
  - TIMESTAMP for date-times
  - BOOLEAN for true/false values
  - TEXT for everything else
- Authentication is required for all operations
- The CLI uses environment variables for storing authentication credentials
- PostgreSQL connection details are provided via command-line flags

## Limitations

1. CSV Format Requirements:
   - Files must have headers as the first row
   - All CSV files in a folder must have the same header structure
   - Headers must be valid PostgreSQL column names (no spaces, special characters)
   - Empty values are treated as NULL

2. Type Inference:
   - Types are inferred from the first data row only
   - Subsequent rows with different formats may cause import errors
   - All numeric values in a column must be consistent (all INTEGER or all NUMERIC)
   - Dates must be in YYYY-MM-DD format
   - Timestamps must be in YYYY-MM-DD[T ]HH:MM:SS format
   - Boolean values accept: true/false, t/f, yes/no (case insensitive)

3. Table Management:
   - Cannot modify existing tables (table must not exist)
   - No support for primary keys or indexes (must be added manually)
   - No support for custom column types or constraints

4. Performance:
   - All files are processed sequentially
   - Entire CSV file is loaded into memory during type inference
   - Large files may consume significant memory

5. Error Handling:
   - Entire transaction is rolled back if any row fails
   - No partial imports within a file
   - No resume capability for failed imports

## Release Process

To create a new release:

1. Tag the release:
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

2. GitHub Actions will automatically:
   - Build binaries for all supported platforms
   - Create a GitHub release
   - Upload the binaries as release assets
