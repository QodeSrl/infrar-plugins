# Complete Web Application Deployment - GCP

This example shows a complete end-to-end deployment using Infrar plugins on GCP.

## Architecture

```
┌─────────────────────────────────────────────┐
│            Internet                          │
└──────────────────┬──────────────────────────┘
                   │
                   ↓
         ┌─────────────────────┐
         │ Cloud Run Service   │
         │ (HTTPS endpoint)    │
         │ (0-10 instances)    │
         │                     │
         │ Uses:               │
         │ - infrar-sdk code   │
         │   (transformed)     │
         │ - Service account   │
         └──┬────────────────┬─┘
            │                │
            ↓                ↓
   ┌────────────────┐  ┌───────────────────┐
   │ Cloud Storage  │  │ Secret Manager    │
   │ (app data)     │  │ (db password)     │
   └────────────────┘  └───────────────────┘
```

## What Gets Deployed

1. **Cloud Storage Bucket** - For application data
   - Versioning (optional)
   - Lifecycle rules (move to Nearline)
   - Uniform bucket-level access

2. **Secret Manager** - For sensitive data
   - Database password
   - Automatic encryption

3. **Cloud Run** - For containerized application
   - HTTPS endpoint (automatic)
   - Auto-scaling (0-10 instances)
   - Cloud Logging
   - Service account with Storage and Secret Manager permissions

## Prerequisites

1. **GCP Project** with billing enabled
2. **OpenTofu/Terraform** installed (`>= 1.0`)
3. **Docker image** pushed to GCR or Artifact Registry
4. **gcloud CLI** configured

## Quick Start

### 1. Enable Required APIs

```bash
gcloud services enable \
  run.googleapis.com \
  storage.googleapis.com \
  secretmanager.googleapis.com \
  --project=my-gcp-project
```

### 2. Build and Push Docker Image

```bash
# Configure Docker for GCR
gcloud auth configure-docker

# Build and push
docker build -t my-web-app .
docker tag my-web-app:latest gcr.io/my-gcp-project/my-web-app:latest
docker push gcr.io/my-gcp-project/my-web-app:latest
```

### 3. Configure Variables

Create `terraform.tfvars`:

```hcl
app_name        = "my-web-app"
project_id      = "my-gcp-project"
environment     = "dev"
region          = "us-central1"
location        = "US"
container_image = "gcr.io/my-gcp-project/my-web-app:latest"
container_port  = 8080
cpu             = 1000  # 1 vCPU
memory          = 512   # 512 MB
min_instances   = 0     # Scale to zero
max_instances   = 10
db_password     = "your-secure-password"  # Use environment variable in production
```

### 4. Deploy

```bash
tofu init
tofu plan
tofu apply
```

### 5. Get Application URL

```bash
tofu output application_url
# Output: https://my-web-app-abc123-uc.a.run.app
```

### 6. Test the Application

```bash
curl $(tofu output -raw application_url)
```

## Application Code

Your Python application should use the transformed Infrar SDK code:

```python
# Original code (before transformation)
from infrar.storage import upload, download

# After transformation (deployed on GCP)
from google.cloud import storage
import os

# Environment variables provided by OpenTofu
BUCKET_NAME = os.environ['BUCKET_NAME']

# Cloud Storage client (uses service account automatically)
storage_client = storage.Client()

def upload_file(local_path, remote_path):
    bucket = storage_client.bucket(BUCKET_NAME)
    blob = bucket.blob(remote_path)
    blob.upload_from_filename(local_path)

def download_file(remote_path, local_path):
    bucket = storage_client.bucket(BUCKET_NAME)
    blob = bucket.blob(remote_path)
    blob.download_to_filename(local_path)

# Access database password from Secret Manager
def get_db_password():
    from google.cloud import secretmanager
    client = secretmanager.SecretManagerServiceClient()
    project_id = os.environ['GCP_PROJECT']
    secret_name = os.environ['DATABASE_SECRET']
    name = f"projects/{project_id}/secrets/{secret_name}/versions/latest"
    response = client.access_secret_version(request={"name": name})
    return response.payload.data.decode('UTF-8')
```

## Monitoring

### View Application Logs

```bash
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=my-web-app" \
  --project=my-gcp-project \
  --limit=50 \
  --format=json
```

### Check Cloud Run Service

```bash
gcloud run services describe my-web-app \
  --region=us-central1 \
  --project=my-gcp-project
```

### Monitor Costs

```bash
gcloud billing accounts list
gcloud billing projects describe my-gcp-project
```

## Cleanup

```bash
tofu destroy
```

## Cost Estimate

Approximate monthly costs for this configuration (with scale-to-zero):

- **Cloud Run** (avg 2 instances, 1 vCPU, 512MB): ~$25/month
- **Cloud Storage** (100GB): ~$2/month
- **Secret Manager** (1 secret): ~$0.06/month
- **Cloud Logging** (5GB): ~$0.50/month
- **Networking** (10GB egress): ~$1/month

**Total**: ~$30/month (can be lower with scale-to-zero)

## Advantages of GCP

- **Scale to Zero**: No cost when idle
- **HTTPS Automatic**: No certificate management needed
- **Simpler Architecture**: No load balancer configuration required
- **Global Deployment**: Easy to deploy to multiple regions

## Next Steps

- Add database (Cloud SQL) using Phase 2 plugins
- Configure custom domain
- Set up CI/CD with Cloud Build
- Enable Cloud CDN for static assets
- Configure Cloud Armor for DDoS protection
