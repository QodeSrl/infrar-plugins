# Infrar Plugins

**Transformation rules and cloud provider implementations for the Infrar Infrastructure Intelligence Platform**

## 📦 What are Infrar Plugins?

Plugins contain the transformation rules that tell the Infrar Engine how to convert provider-agnostic code (using Infrar SDK) into native cloud provider SDK code (boto3, google-cloud-storage, etc.).

Each plugin defines:
- **Transformation rules**: How to map Infrar API calls to provider-specific code
- **Code templates**: Provider-specific code patterns
- **Parameter mappings**: How to translate parameters between APIs
- **Dependencies**: Required packages for each provider

## 🗂️ Plugin Structure

```
infrar-plugins/
└── packages/
    └── {capability}/              # e.g., storage, database, messaging
        ├── capability.yaml         # Capability definition (future)
        ├── aws/
        │   └── rules.yaml         # AWS transformation rules
        ├── gcp/
        │   └── rules.yaml         # GCP transformation rules
        └── azure/
            └── rules.yaml         # Azure transformation rules (future)
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

### Future Plugins 🔜

- **Database** - Relational database operations (planned Phase 2)
- **Messaging** - Queue and pub/sub operations (planned Phase 2)
- **Compute** - Container deployment (planned Phase 2)
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

| Capability | AWS | GCP | Azure | Status |
|------------|-----|-----|-------|--------|
| **Storage** | ✅ S3 | ✅ Cloud Storage | ⏳ Planned | MVP Ready |
| **Database** | ⏳ RDS | ⏳ Cloud SQL | ⏳ Planned | Phase 2 |
| **Messaging** | ⏳ SQS | ⏳ Pub/Sub | ⏳ Planned | Phase 2 |
| **Compute** | ⏳ ECS | ⏳ Cloud Run | ⏳ Planned | Phase 2 |

## 📄 License

Apache License 2.0 - see [LICENSE](LICENSE) file for details.

## 🔗 Related Repositories

- [infrar-engine](https://github.com/QodeSrl/infrar-engine) - Transformation engine
- [infrar-sdk-python](https://github.com/QodeSrl/infrar-sdk-python) - Python SDK
- [infrar-docs](https://github.com/QodeSrl/infrar-docs) - Documentation

---

**Part of the Infrar project** - Infrastructure Intelligence for the multi-cloud era
