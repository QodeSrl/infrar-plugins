# Infrar Plugins

**Modular cloud provider and service implementations for the Infrar Infrastructure Intelligence Platform**

## Overview

Plugin system providing three-level orchestration architecture for deploying applications to cloud providers. Each provider, category, and service has its own orchestrator that generates Terraform configurations dynamically.

## Architecture

### Three-Level Orchestration

1. **Service Level** - Individual cloud services (e.g., cloud-run, cloud-storage)
   - Generates service-specific Terraform resources
   - Handles service-level configuration and builds

2. **Category Level** - Service categories (e.g., compute, storage)
   - Combines services within a category
   - Coordinates service orchestrators

3. **Provider Level** - Cloud providers (e.g., GCP)
   - Assembles complete provider configuration
   - Coordinates category orchestrators

## Current Implementation

### Supported Cloud Providers

#### Google Cloud Platform (GCP) ✅

**Compute Services:**
- **Cloud Run** - Deploy containerized Python applications with automatic scaling
  - Automatic container building from source code
  - Integrated Infrar SDK for storage operations
  - HTTP endpoint with health checks
  - Auto-scaling 0-10 instances

**Storage Services:**
- **Cloud Storage** - Object storage with global availability
  - Bucket creation and management
  - Versioning and lifecycle rules
  - CORS configuration support
  - IAM integration

## Plugin Structure

```
providers/
└── gcp/
    ├── orchestrator/                  # Provider-level orchestration
    │   ├── main.go
    │   └── orchestrate                # Binary
    ├── terraform-config/              # Provider Terraform templates
    │   ├── provider-block.tf.tmpl
    │   ├── variables.tf.tmpl
    │   └── tfvars.tmpl
    └── services/
        ├── compute/
        │   ├── orchestrator/          # Category orchestrator
        │   │   ├── main.go
        │   │   └── orchestrate
        │   └── cloud-run/
        │       ├── orchestrator/      # Service orchestrator
        │       │   ├── main.go
        │       │   └── orchestrate
        │       ├── terraform/         # Service Terraform templates
        │       │   ├── main.tf
        │       │   ├── variables.tf
        │       │   ├── outputs.tf
        │       │   └── tfvars.tmpl
        │       ├── service.yaml       # Service metadata
        │       └── build.sh           # Container build script
        └── storage/
            ├── orchestrator/          # Category orchestrator
            │   ├── main.go
            │   └── orchestrate
            └── cloud-storage/
                ├── orchestrator/      # Service orchestrator
                │   ├── main.go
                │   └── orchestrate
                ├── terraform/         # Service Terraform templates
                │   ├── main.tf
                │   ├── variables.tf
                │   ├── outputs.tf
                │   └── tfvars.tmpl
                └── service.yaml       # Service metadata
```

## How It Works

### 1. Code Analysis
The platform analyzes user code to detect required capabilities:
```python
import infrar.storage
infrar.storage.upload(bucket='my-bucket', source='file.txt')
```
Detected capabilities: `storage`, `compute` (for execution)

### 2. Service Recommendation
Based on capabilities, recommend cloud services:
- Storage capability → Cloud Storage
- Compute capability → Cloud Run

### 3. Orchestrated Terraform Generation

**Provider Orchestrator** (gcp/orchestrator)
- Receives: capabilities, context, credentials, custom variables
- Calls: category orchestrators for compute and storage
- Generates: provider.tf, combines all service resources

**Category Orchestrator** (compute/orchestrator)
- Receives: compute capabilities
- Calls: cloud-run service orchestrator
- Generates: combined category resources

**Service Orchestrator** (cloud-run/orchestrator)
- Receives: service parameters
- Generates: main.tf, variables.tf, terraform.tfvars for Cloud Run

### 4. Container Build (Cloud Run only)

**build.sh Script:**
1. Receives user's Python code via stdin
2. Creates Infrar SDK (`infrar/storage.py`)
3. Wraps code in Flask application
4. Builds Docker container
5. Authenticates to Google Container Registry
6. Pushes image to `gcr.io/project-id/app-name:latest`
7. Returns image URL

### 5. Infrastructure Deployment
Platform runs terraform:
```bash
terraform init
terraform plan
terraform apply
```

Resources created:
- Cloud Storage bucket (custom named)
- Cloud Run service running user's container
- IAM bindings for access

## Service Interface

### service.yaml
```yaml
name: cloud-run
display_name: Cloud Run
category: compute
provider: gcp
capabilities:
  - compute
description: Deploy containerized applications with automatic scaling
```

### Orchestrator Binary (Go)

**Input (JSON via stdin):**
```json
{
  "command": "generate",
  "capabilities": ["compute"],
  "context": {
    "project_name": "my-app",
    "environment": "production",
    "region": "us-central1"
  },
  "credentials": {
    "gcp_service_account_json": "..."
  },
  "parameters": {
    "service_name": "my-service",
    "container_image": "gcr.io/project/image:latest"
  }
}
```

**Output (JSON via stdout):**
```json
{
  "success": true,
  "files": {
    "main.tf": "...",
    "variables.tf": "...",
    "terraform.tfvars": "..."
  },
  "metadata": {
    "services_included": ["cloud-run"],
    "warnings": [],
    "required_apis": ["run.googleapis.com"]
  }
}
```

### build.sh (Optional, for containerization)

**Input (JSON via stdin):**
```json
{
  "project_id": "my-gcp-project",
  "code": "import infrar.storage\n...",
  "image_name": "my-app",
  "credentials": "{\"project_id\": \"...\"}"
}
```

**Output (JSON via stdout):**
```json
{
  "success": true,
  "image": "gcr.io/my-gcp-project/my-app:latest",
  "message": "Container built and pushed successfully"
}
```

## Template Functions

Available in all `.tmpl` files:

- **`tfstring`** - Quote value for Terraform: `{{ .name | tfstring }}` → `"value"`
- **`default`** - Provide default: `{{ .var | default "fallback" }}`
- **`sanitize`** - Lowercase and clean: `{{ .ProjectName | sanitize }}` → `"my-project"`

## Development

### Building Orchestrators

```bash
# Provider orchestrator
cd providers/gcp/orchestrator
go build -o orchestrate main.go

# Category orchestrators
cd providers/gcp/services/compute/orchestrator
go build -o orchestrate main.go

cd providers/gcp/services/storage/orchestrator
go build -o orchestrate main.go

# Service orchestrators
cd providers/gcp/services/compute/cloud-run/orchestrator
go build -o orchestrate main.go

cd providers/gcp/services/storage/cloud-storage/orchestrator
go build -o orchestrate main.go
```

### Testing Orchestrator

```bash
echo '{
  "command": "generate",
  "capabilities": ["compute"],
  "context": {
    "project_name": "test",
    "environment": "dev",
    "region": "us-central1"
  },
  "credentials": {},
  "parameters": {}
}' | ./providers/gcp/services/compute/cloud-run/orchestrator/orchestrate | jq .
```

### Testing Build Script

```bash
echo '{
  "project_id": "test-project",
  "code": "import infrar.storage\nprint(\"hello\")",
  "image_name": "test-app",
  "credentials": "{}"
}' | ./providers/gcp/services/compute/cloud-run/build.sh
```

## Infrar SDK (Included in Containers)

The build script automatically includes the Infrar Python SDK:

**infrar/storage.py:**
```python
from google.cloud import storage

def upload(bucket, source, destination=None):
    """Upload file to Cloud Storage"""
    if destination is None:
        destination = os.path.basename(source)

    client = storage.Client()
    bucket_obj = client.bucket(bucket)
    blob = bucket_obj.blob(destination)
    blob.upload_from_filename(source)

    return f"gs://{bucket}/{destination}"
```

## Roadmap

### Phase 1 (Current - MVP)
- ✅ GCP provider
- ✅ Cloud Storage
- ✅ Cloud Run
- ✅ Three-level orchestration
- ✅ Automatic containerization
- ✅ Custom variable support

### Phase 2 (Planned)
- AWS provider (Lambda, S3)
- Azure provider (Container Apps, Blob Storage)
- Additional GCP services (Cloud Functions, Cloud SQL)
- Secret management integration

### Phase 3 (Future)
- Database plugins
- Messaging plugins
- Multi-cloud deployments
- Cost optimization

## License

GNU General Public License v3.0

## Related Repositories

- [infrar-platform](https://github.com/QodeSrl/infrar-platform) - Platform backend
- [infrar-engine](https://github.com/QodeSrl/infrar-engine) - Transformation engine
- [infrar-sdk-python](https://github.com/QodeSrl/infrar-sdk-python) - Python SDK
- [infrar-docs](https://github.com/QodeSrl/infrar-docs) - Documentation

---

**Part of the Infrar project** - Infrastructure Intelligence for the multi-cloud era
