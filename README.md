# Infrar Plugins

**Transformation rules and cloud provider implementations for the Infrar Infrastructure Intelligence Platform**

## 📦 What are Infrar Plugins?

Plugins provide everything needed to deploy applications to cloud providers:

1. **Code Transformation Rules** - Convert Infrar SDK calls to provider-specific SDK code (boto3, google-cloud-storage, etc.)
2. **OpenTofu Modules** - Infrastructure-as-code templates for provisioning resources

Each plugin defines:
- **Transformation rules**: How to map Infrar API calls to provider-specific code
- **Code templates**: Provider-specific code patterns
- **Parameter mappings**: How to translate parameters between APIs
- **Dependencies**: Required packages for each provider
- **OpenTofu modules**: Infrastructure provisioning templates

## 🗂️ Plugin Structure

```
infrar-plugins/
└── packages/
    └── {capability}/              # e.g., storage, compute, secrets
        ├── aws/
        │   ├── rules.yaml         # Code transformation rules
        │   └── terraform/         # OpenTofu/Terraform module
        │       ├── main.tf
        │       ├── variables.tf
        │       ├── outputs.tf
        │       └── README.md
        ├── gcp/
        │   ├── rules.yaml
        │   └── terraform/
        └── azure/
            ├── rules.yaml
            └── terraform/         # (planned)
```

## 🚀 Available Plugins

### Storage Plugin ✅ AVAILABLE

**Capability**: Object storage operations (upload, download, delete, list)

**Providers**:
- ✅ **AWS S3** - Complete (4 operations)
- ✅ **GCP Cloud Storage** - Complete (4 operations)
- ⏳ **Azure Blob Storage** - Planned

**Operations**:
1. `upload(bucket, source, destination)` - Upload file to storage
2. `download(bucket, source, destination)` - Download file from storage
3. `delete(bucket, path)` - Delete object from storage
4. `list_objects(bucket, prefix)` - List objects in bucket

**Example Transformation**:

```python
# Input (Infrar SDK)
from infrar.storage import upload
upload(bucket='data', source='file.csv', destination='backup/file.csv')

# Output (AWS/boto3)
import boto3
s3 = boto3.client('s3')
s3.upload_file('file.csv', 'data', 'backup/file.csv')

# Output (GCP/Cloud Storage)
from google.cloud import storage
storage_client = storage.Client()
bucket = storage_client.bucket('data')
blob = bucket.blob('backup/file.csv')
blob.upload_from_filename('file.csv')
```

### Compute Plugin ✅ AVAILABLE

**Capability**: Deploy containerized web applications

**Providers**:
- ✅ **AWS ECS Fargate** - Complete
- ✅ **GCP Cloud Run** - Complete
- ⏳ **Azure Container Apps** - Planned

**Features**:
- Serverless container deployment
- Application Load Balancer (AWS) / HTTPS endpoints (GCP)
- Auto-scaling
- Health checks
- CloudWatch/Cloud Logging integration

**OpenTofu Modules**:
- `packages/compute/aws/terraform` - ECS Fargate deployment
- `packages/compute/gcp/terraform` - Cloud Run deployment

### Secrets Plugin ✅ AVAILABLE

**Capability**: Secure secrets management

**Providers**:
- ✅ **AWS Secrets Manager** - Complete
- ✅ **GCP Secret Manager** - Complete
- ⏳ **Azure Key Vault** - Planned

**Features**:
- Encrypted storage
- Version management
- IAM integration
- Automatic rotation support

**OpenTofu Modules**:
- `packages/secrets/aws/terraform` - AWS Secrets Manager
- `packages/secrets/gcp/terraform` - GCP Secret Manager

### Future Plugins 🔜

- **Database** - Relational database operations (planned Phase 2)
- **Messaging** - Queue and pub/sub operations (planned Phase 2)
- **Data Analytics** - Data warehousing and ETL (planned Phase 3)

## 📝 Transformation Rule Format

Plugins use YAML to define transformation rules:

```yaml
operations:
  - name: upload
    pattern: "infrar.storage.upload"
    target:
      provider: aws
      service: s3
      operation: upload_file

    transformation:
      imports:
        - "import boto3"

      setup_code: "s3 = boto3.client('s3')"

      code_template: "s3.upload_file({{ .source }}, {{ .bucket }}, {{ .destination }})"

      parameter_mapping:
        bucket: bucket
        source: source
        destination: destination

    requirements:
      - package: boto3
        version: ">=1.28.0"
```

## 🔌 Using Plugins

### Code Transformation

Plugins are loaded automatically by the Infrar Engine:

```go
// Load storage plugin for AWS
engine.LoadRules("./infrar-plugins/packages", types.ProviderAWS, "storage")

// Transform code
result := engine.Transform(sourceCode, types.ProviderAWS)
```

Or via CLI:

```bash
infrar transform --provider aws --plugins ./infrar-plugins/packages --input app.py
```

### Infrastructure Provisioning

Use OpenTofu modules to deploy infrastructure:

```hcl
# Deploy storage bucket on AWS
module "storage" {
  source = "./infrar-plugins/packages/storage/aws/terraform"

  bucket_name        = "my-app-data"
  versioning_enabled = true
}

# Deploy web application on AWS
module "web_app" {
  source = "./infrar-plugins/packages/compute/aws/terraform"

  app_name        = "my-web-app"
  container_image = "123456789.dkr.ecr.us-east-1.amazonaws.com/my-app:latest"
  container_port  = 8080
  cpu             = 512
  memory          = 1024
}

# Grant app access to bucket
resource "aws_iam_role_policy" "app_storage" {
  role = module.web_app.task_role_arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = ["s3:GetObject", "s3:PutObject"]
      Resource = "${module.storage.bucket_arn}/*"
    }]
  })
}
```

## 🛠️ Creating Custom Plugins

Want to add support for a new provider or capability? Follow these steps:

### 1. Create Plugin Structure

```bash
mkdir -p packages/{capability}/{provider}
cd packages/{capability}/{provider}
```

### 2. Define Transformation Rules

Create `rules.yaml` with your transformation rules (see format above)

### 3. Test Your Plugin

```bash
# Test transformation
echo "from infrar.capability import operation" | \
  infrar transform --provider your_provider --plugins ./packages
```

### 4. Submit Pull Request

We welcome community contributions!

## 📊 Plugin Status

| Capability | AWS | GCP | Azure | Code Transform | OpenTofu Modules | Status |
|------------|-----|-----|-------|----------------|------------------|--------|
| **Storage** | S3 | Cloud Storage | Blob (planned) | ✅ | ✅ | **MVP Ready** |
| **Compute** | ECS Fargate | Cloud Run | Container Apps (planned) | ⏳ | ✅ | **MVP Ready** |
| **Secrets** | Secrets Manager | Secret Manager | Key Vault (planned) | ⏳ | ✅ | **Phase 2** |
| **Database** | RDS | Cloud SQL | Azure SQL | ⏳ | ⏳ | Phase 2 |
| **Messaging** | SQS | Pub/Sub | Service Bus | ⏳ | ⏳ | Phase 2 |

## 📄 License

GNU General Public License v3.0 - see [LICENSE](LICENSE) file for details.

## 🔗 Related Repositories

- [infrar-engine](https://github.com/QodeSrl/infrar-engine) - Transformation engine
- [infrar-sdk-python](https://github.com/QodeSrl/infrar-sdk-python) - Python SDK
- [infrar-docs](https://github.com/QodeSrl/infrar-docs) - Documentation

---

**Part of the Infrar project** - Infrastructure Intelligence for the multi-cloud era
