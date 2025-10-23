# Complete Web Application Deployment - AWS

This example shows a complete end-to-end deployment using Infrar plugins on AWS.

## Architecture

```
┌─────────────────────────────────────────────┐
│            Internet                          │
└──────────────────┬──────────────────────────┘
                   │
                   ↓
         ┌─────────────────────┐
         │ Application Load    │
         │ Balancer (ALB)      │
         └─────────┬───────────┘
                   │
                   ↓
         ┌─────────────────────┐
         │ ECS Fargate         │
         │ (2 instances)       │
         │                     │
         │ Uses:               │
         │ - infrar-sdk code   │
         │   (transformed)     │
         │ - IAM role          │
         └──┬────────────────┬─┘
            │                │
            ↓                ↓
   ┌────────────────┐  ┌───────────────────┐
   │ S3 Bucket      │  │ Secrets Manager   │
   │ (app data)     │  │ (db password)     │
   └────────────────┘  └───────────────────┘
```

## What Gets Deployed

1. **S3 Bucket** - For application data storage
   - Versioning (optional)
   - Lifecycle rules (archive to Glacier)
   - Encryption enabled

2. **Secrets Manager** - For sensitive data
   - Database password
   - Encrypted at rest

3. **ECS Fargate** - For containerized application
   - Application Load Balancer
   - Auto-scaling (2 instances)
   - CloudWatch Logs
   - IAM role with S3 and Secrets Manager permissions

## Prerequisites

1. **AWS Account** with credentials configured
2. **OpenTofu/Terraform** installed (`>= 1.0`)
3. **Docker image** pushed to ECR
4. **AWS CLI** configured

## Quick Start

### 1. Build and Push Docker Image

```bash
# Create ECR repository
aws ecr create-repository --repository-name my-web-app

# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin \
  123456789.dkr.ecr.us-east-1.amazonaws.com

# Build and push
docker build -t my-web-app .
docker tag my-web-app:latest \
  123456789.dkr.ecr.us-east-1.amazonaws.com/my-web-app:latest
docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/my-web-app:latest
```

### 2. Configure Variables

Create `terraform.tfvars`:

```hcl
app_name        = "my-web-app"
environment     = "dev"
region          = "us-east-1"
container_image = "123456789.dkr.ecr.us-east-1.amazonaws.com/my-web-app:latest"
container_port  = 8080
cpu             = 512
memory          = 1024
desired_count   = 2
db_password     = "your-secure-password"  # Use environment variable in production
```

### 3. Deploy

```bash
tofu init
tofu plan
tofu apply
```

### 4. Get Application URL

```bash
tofu output application_url
# Output: http://my-web-app-alb-123456789.us-east-1.elb.amazonaws.com
```

### 5. Test the Application

```bash
curl http://$(tofu output -raw load_balancer_dns)
```

## Application Code

Your Python application should use the transformed Infrar SDK code:

```python
# Original code (before transformation)
from infrar.storage import upload, download

# After transformation (deployed on AWS)
import boto3
import os

# Environment variables provided by OpenTofu
BUCKET_NAME = os.environ['BUCKET_NAME']

# S3 client (uses IAM role automatically)
s3 = boto3.client('s3')

def upload_file(local_path, remote_path):
    s3.upload_file(local_path, BUCKET_NAME, remote_path)

def download_file(remote_path, local_path):
    s3.download_file(BUCKET_NAME, remote_path, local_path)

# Access database password from Secrets Manager
def get_db_password():
    import json
    secrets = boto3.client('secretsmanager')
    secret_arn = os.environ['DATABASE_SECRET']
    response = secrets.get_secret_value(SecretId=secret_arn)
    return response['SecretString']
```

## Monitoring

### View Application Logs

```bash
# Get log group name
LOG_GROUP=$(tofu output -raw log_group_name)

# View logs
aws logs tail $LOG_GROUP --follow
```

### Check ECS Service

```bash
aws ecs describe-services \
  --cluster my-web-app-cluster \
  --services my-web-app
```

### Monitor Costs

```bash
aws ce get-cost-and-usage \
  --time-period Start=2025-01-01,End=2025-01-31 \
  --granularity MONTHLY \
  --metrics BlendedCost \
  --filter file://cost-filter.json
```

## Cleanup

```bash
tofu destroy
```

## Cost Estimate

Approximate monthly costs for this configuration:

- **ECS Fargate** (2 × 0.5 vCPU × 1GB): ~$30/month
- **Application Load Balancer**: ~$16/month
- **S3 Storage** (100GB): ~$2/month
- **Secrets Manager** (1 secret): ~$0.40/month
- **CloudWatch Logs** (5GB): ~$1/month

**Total**: ~$50/month

## Next Steps

- Add database (RDS) using Phase 2 plugins
- Configure custom domain with Route 53
- Set up CI/CD pipeline
- Add auto-scaling policies
- Configure HTTPS with ACM certificate
