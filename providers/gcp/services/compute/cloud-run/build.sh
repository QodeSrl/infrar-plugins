#!/bin/bash
# Cloud Run Build Script
# Builds Python code into a container image and pushes to GCR

set -e

# Parse JSON input from stdin
INPUT=$(cat)

PROJECT_ID=$(echo "$INPUT" | jq -r '.project_id')
CODE=$(echo "$INPUT" | jq -r '.code')
IMAGE_NAME=$(echo "$INPUT" | jq -r '.image_name')
CREDENTIALS=$(echo "$INPUT" | jq -r '.credentials')

BUILD_DIR="/tmp/infrar-build-$$"
mkdir -p "$BUILD_DIR"

cleanup() {
    rm -rf "$BUILD_DIR"
}
trap cleanup EXIT

echo "[BUILD] Building Cloud Run container for $IMAGE_NAME" >&2

# Create infrar SDK
mkdir -p "$BUILD_DIR/infrar"
cat > "$BUILD_DIR/infrar/__init__.py" << 'EOF'
EOF

cat > "$BUILD_DIR/infrar/storage.py" << 'EOF'
"""Infrar Storage SDK - Simple wrapper around Google Cloud Storage"""
from google.cloud import storage
import os

def upload(bucket, source, destination=None):
    """Upload a file to Cloud Storage

    Args:
        bucket: Name of the GCS bucket
        source: Path to the local file to upload
        destination: Optional destination path in the bucket (defaults to source filename)

    Returns:
        Public URL of the uploaded file
    """
    if destination is None:
        destination = os.path.basename(source)

    client = storage.Client()
    bucket_obj = client.bucket(bucket)
    blob = bucket_obj.blob(destination)

    blob.upload_from_filename(source)
    print(f"✓ Uploaded {source} to gs://{bucket}/{destination}")

    return f"gs://{bucket}/{destination}"
EOF

# Write user code
echo "$CODE" > "$BUILD_DIR/user_code.py"

# Create Flask wrapper
cat > "$BUILD_DIR/main.py" << 'EOF'
import os
import sys
from flask import Flask, jsonify

app = Flask(__name__)

# Store execution result
execution_result = {"status": "not_executed", "output": "", "error": ""}

@app.route('/', methods=['GET', 'POST'])
def run():
    """Execute user code when endpoint is called"""
    global execution_result

    if execution_result["status"] == "not_executed":
        try:
            # Import and execute user code
            import user_code
            execution_result = {
                "status": "success",
                "output": "User code executed successfully",
                "error": ""
            }
        except Exception as e:
            execution_result = {
                "status": "error",
                "output": "",
                "error": str(e)
            }

    return jsonify(execution_result), 200 if execution_result["status"] == "success" else 500

@app.route('/health', methods=['GET'])
def health():
    return jsonify({"status": "healthy"}), 200

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8080))
    print(f"Starting server on port {port}...")
    sys.stdout.flush()
    app.run(host='0.0.0.0', port=port, debug=False)
EOF

# Create requirements.txt
cat > "$BUILD_DIR/requirements.txt" << 'EOF'
Flask==3.0.0
google-cloud-storage==2.10.0
requests==2.31.0
EOF

# Create Dockerfile
cat > "$BUILD_DIR/Dockerfile" << 'EOF'
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY infrar/ ./infrar/
COPY main.py .
COPY user_code.py .

ENV PORT=8080

CMD ["python", "main.py"]
EOF

# Build image
IMAGE_TAG="gcr.io/$PROJECT_ID/$IMAGE_NAME:latest"
echo "[BUILD] Building image: $IMAGE_TAG" >&2
docker build -t "$IMAGE_TAG" "$BUILD_DIR" >&2

# Authenticate Docker to GCR
echo "[BUILD] Authenticating to GCR" >&2
echo "$CREDENTIALS" | docker login -u _json_key --password-stdin gcr.io >&2

# Push image
echo "[BUILD] Pushing image to GCR" >&2
docker push "$IMAGE_TAG" >&2

# Output success
cat << EOF
{
  "success": true,
  "image": "$IMAGE_TAG",
  "message": "Container built and pushed successfully"
}
EOF
