# Docker Hub Deployment Guide

This guide covers building, testing, and deploying Askara to Docker Hub for easy distribution.

## For the Developer (Building & Publishing)

### Prerequisites

1. **Docker Desktop** installed and running
2. **Docker Hub account** - Create one at https://hub.docker.com
3. **Git** installed
4. **Access to this repository**

### Step 1: Clone the Repository

```bash
# Create a project folder
mkdir askara-docker
cd askara-docker

# Clone the repository
git clone https://github.com/itaiguardiola/Askara.git
cd Askara

# Switch to the Ollama integration branch
git checkout claude/askara-next-steps-011CUtErzJPJHAFP2xpkh2Le
```

### Step 2: Build the Docker Image

```bash
# Build the image (this will take 5-10 minutes)
docker build -t askara:latest .

# Verify the build succeeded
docker images | grep askara
```

**Expected output:**
```
askara       latest    abc123def456   2 minutes ago   XXX MB
```

### Step 3: Test Locally (IMPORTANT!)

Before pushing to Docker Hub, test the image works correctly.

#### Test with Ollama (Recommended for Windows RTX 3090)

1. **Install and Start Ollama** on your Windows machine:
   ```powershell
   # Download from https://ollama.ai/download
   # After installation:
   ollama serve
   ```

2. **Pull required models**:
   ```powershell
   ollama pull llama2
   ollama pull nomic-embed-text
   ```

3. **Create .env file**:
   ```bash
   cp .env.example .env
   ```

   Edit `.env` and set:
   ```bash
   LLM_PROVIDER=ollama
   OLLAMA_HOST=http://host.docker.internal:11434
   OLLAMA_MODEL=llama2
   OLLAMA_EMBEDDING_MODEL=nomic-embed-text
   QDRANT_API_ENDPOINT=http://qdrant:6333
   ```

4. **Start the application**:
   ```bash
   # Windows users:
   docker-compose -f docker-compose.yml -f docker-compose.windows.yml up -d

   # Linux/Mac users:
   docker-compose up -d
   ```

5. **Test the application**:
   - Open http://localhost:8100
   - Upload a PDF or text file
   - Ask a question about the uploaded file
   - Verify you get a response

6. **Check logs** to confirm Ollama connection:
   ```bash
   docker-compose logs askara-web
   ```

   Look for:
   ```
   LLM provider initialized successfully
   ```

#### Test with OpenAI (Alternative)

If you want to test with OpenAI instead:

1. **Create .env file**:
   ```bash
   LLM_PROVIDER=openai
   OPENAI_API_KEY=sk-your-actual-key-here
   QDRANT_API_ENDPOINT=http://qdrant:6333
   ```

2. **Start and test** as above

### Step 4: Tag the Image for Docker Hub

```bash
# Login to Docker Hub
docker login

# Tag the image with your Docker Hub username
# Replace 'yourusername' with your actual Docker Hub username
docker tag askara:latest yourusername/askara:latest
docker tag askara:latest yourusername/askara:v1.0.0

# If you want to create a specific tag for Ollama support:
docker tag askara:latest yourusername/askara:ollama
```

### Step 5: Push to Docker Hub

```bash
# Push all tags
docker push yourusername/askara:latest
docker push yourusername/askara:v1.0.0
docker push yourusername/askara:ollama
```

**This will take 5-15 minutes** depending on your internet speed.

### Step 6: Verify on Docker Hub

1. Go to https://hub.docker.com
2. Navigate to your repository: `yourusername/askara`
3. Verify the tags are present: `latest`, `v1.0.0`, `ollama`

### Step 7: Create Repository Description

On Docker Hub, add this description to your repository:

```markdown
# Askara - AI-Powered Document Q&A System

Upload documents (PDF, TXT, EPUB) and ask questions using AI. Get answers based on your document content.

## Features
- **Local AI** with Ollama (FREE) or cloud with OpenAI (paid)
- **GPU Accelerated** - Optimized for NVIDIA GPUs (RTX 3090 tested)
- **Document Management** - Upload, view, delete documents
- **Vector Search** - Uses Qdrant or Pinecone
- **Streaming Responses** - Real-time AI answers

## Quick Start

### Option 1: Ollama (FREE - Local AI)

1. Install Ollama: https://ollama.ai/download
2. Start Ollama and pull models:
   ```bash
   ollama serve
   ollama pull llama2
   ollama pull nomic-embed-text
   ```
3. Create `.env` file:
   ```bash
   LLM_PROVIDER=ollama
   OLLAMA_HOST=http://host.docker.internal:11434
   QDRANT_API_ENDPOINT=http://qdrant:6333
   ```
4. Run:
   ```bash
   docker run -d \
     --name askara \
     -p 8100:8100 \
     --env-file .env \
     --add-host=host.docker.internal:host-gateway \
     yourusername/askara:latest
   ```
5. Open http://localhost:8100

### Option 2: OpenAI (Paid)

1. Get API key from https://platform.openai.com
2. Create `.env` file:
   ```bash
   LLM_PROVIDER=openai
   OPENAI_API_KEY=sk-your-key-here
   QDRANT_API_ENDPOINT=http://qdrant:6333
   ```
3. Run with docker-compose (see GitHub repo)

## Documentation

Full documentation: https://github.com/itaiguardiola/Askara

## Tags

- `latest` - Latest stable build
- `v1.0.0` - Specific version
- `ollama` - Optimized for Ollama

## System Requirements

- Docker Desktop (Windows/Mac) or Docker Engine (Linux)
- 4GB RAM minimum, 8GB+ recommended
- For Ollama: 8GB+ RAM, GPU optional but recommended
```

---

## For End Users (Downloading & Running)

### Windows Users with RTX 3090 (Recommended Setup)

#### Prerequisites

1. **Docker Desktop for Windows**
   - Download: https://www.docker.com/products/docker-desktop
   - Enable WSL 2 backend in settings
   - Allocate at least 8GB RAM in settings

2. **Ollama for Windows**
   - Download: https://ollama.ai/download
   - Install and start Ollama
   - Ollama will automatically use your RTX 3090 GPU

3. **Git for Windows** (to clone config files)
   - Download: https://git-scm.com/download/win

#### Quick Start (5 minutes)

**Step 1: Install Ollama and Pull Models**

Open PowerShell and run:
```powershell
# Start Ollama (if not already running)
ollama serve

# In a new PowerShell window, pull models
ollama pull llama2
ollama pull nomic-embed-text

# Verify models are downloaded
ollama list
```

**Step 2: Create Project Folder**

```powershell
# Create folder
mkdir C:\Askara
cd C:\Askara

# Clone just the config files
git clone --depth 1 --branch claude/askara-next-steps-011CUtErzJPJHAFP2xpkh2Le https://github.com/itaiguardiola/Askara.git config
cd config
```

**Step 3: Configure Environment**

```powershell
# Copy example config
copy .env.example .env

# Edit .env with Notepad
notepad .env
```

Make sure these values are set:
```bash
LLM_PROVIDER=ollama
OLLAMA_HOST=http://host.docker.internal:11434
OLLAMA_MODEL=llama2
OLLAMA_EMBEDDING_MODEL=nomic-embed-text
QDRANT_API_ENDPOINT=http://qdrant:6333
```

Save and close Notepad.

**Step 4: Start Askara**

```powershell
# Pull the image from Docker Hub (replace 'yourusername' with actual username)
docker pull yourusername/askara:latest

# Start with docker-compose
docker-compose -f docker-compose.yml -f docker-compose.windows.yml up -d
```

**Step 5: Open Askara**

Open your browser and go to: http://localhost:8100

You should see the Askara interface!

**Step 6: Test It**

1. Upload a PDF or text file
2. Wait for "All files uploaded successfully"
3. Ask a question about your document
4. Get AI-powered answers using your RTX 3090!

#### Verify GPU Usage

To confirm your RTX 3090 is being used:

```powershell
# In a new PowerShell window while asking questions
nvidia-smi
```

You should see `ollama` using GPU memory.

#### Stopping Askara

```powershell
cd C:\Askara\config
docker-compose down
```

#### Updating to Latest Version

```powershell
docker-compose down
docker pull yourusername/askara:latest
docker-compose up -d
```

---

### Alternative: Running Without Docker Compose

If you prefer a single `docker run` command:

```bash
# Start Qdrant first
docker run -d \
  --name askara-qdrant \
  -p 6333:6333 \
  -v qdrant-data:/qdrant/storage \
  qdrant/qdrant:latest

# Start Askara
docker run -d \
  --name askara-web \
  -p 8100:8100 \
  -e LLM_PROVIDER=ollama \
  -e OLLAMA_HOST=http://host.docker.internal:11434 \
  -e OLLAMA_MODEL=llama2 \
  -e OLLAMA_EMBEDDING_MODEL=nomic-embed-text \
  -e QDRANT_API_ENDPOINT=http://askara-qdrant:6333 \
  --add-host=host.docker.internal:host-gateway \
  --link askara-qdrant \
  -v askara-data:/app/data \
  yourusername/askara:latest

# Open http://localhost:8100
```

---

## Troubleshooting

### "Connection refused" to Ollama

**Problem:** Askara can't connect to Ollama

**Solutions:**
1. Make sure Ollama is running: `ollama serve`
2. Check Ollama is accessible: `curl http://localhost:11434/api/tags`
3. Verify `OLLAMA_HOST=http://host.docker.internal:11434` in `.env`
4. Ensure `extra_hosts` is in docker-compose (should be by default)

### "Model not found"

**Problem:** Ollama says model doesn't exist

**Solutions:**
```bash
ollama pull llama2
ollama pull nomic-embed-text
ollama list  # Verify they're there
```

### Docker build fails

**Problem:** `docker build` command fails

**Solutions:**
1. Make sure you're in the Askara directory
2. Check Docker Desktop is running
3. Try: `docker system prune` then rebuild
4. Check internet connection (build downloads dependencies)

### Slow performance

**Problem:** Responses are very slow

**Solutions:**
1. Use smaller model: `OLLAMA_MODEL=mistral`
2. Check GPU usage: `nvidia-smi`
3. Increase Docker Desktop RAM allocation (Settings → Resources)
4. Close other GPU-intensive applications

### Port already in use

**Problem:** "port 8100 is already allocated"

**Solutions:**
```bash
# Find what's using port 8100
netstat -ano | findstr :8100

# Stop it or change Askara port in .env:
PORT=8200
```

---

## Docker Hub Repository Management

### Updating Your Image

When you make changes:

```bash
# Pull latest code
git pull origin claude/askara-next-steps-011CUtErzJPJHAFP2xpkh2Le

# Rebuild
docker build -t askara:latest .

# Test locally
docker-compose up -d
# Test at http://localhost:8100

# Tag with new version
docker tag askara:latest yourusername/askara:v1.1.0
docker tag askara:latest yourusername/askara:latest

# Push
docker push yourusername/askara:v1.1.0
docker push yourusername/askara:latest
```

### Creating Release Notes

On Docker Hub, document what changed:

```
## v1.1.0 - 2025-01-15

### Added
- Feature X
- Feature Y

### Fixed
- Bug Z

### Changed
- Improvement W
```

---

## Performance Tips

### For RTX 3090 Users

1. **Use quantized models** for faster inference:
   ```bash
   ollama pull llama2:13b-q4_0  # 13B model, 4-bit quantization
   ```

   Update `.env`:
   ```bash
   OLLAMA_MODEL=llama2:13b-q4_0
   ```

2. **Monitor GPU usage**:
   ```bash
   # Watch GPU in real-time
   nvidia-smi -l 1
   ```

3. **Adjust context window** (if responses are slow):
   - Edit `llm/ollama.go`
   - Change `NumCtx` parameter
   - Rebuild image

### For CPU-Only Users

Use smaller models:
```bash
ollama pull llama2:7b
# or
ollama pull mistral:7b
```

---

## Support

- **GitHub Issues**: https://github.com/itaiguardiola/Askara/issues
- **Documentation**: See `README.md`, `OLLAMA_INTEGRATION.md`, `WINDOWS_SETUP.md`
- **Ollama Help**: https://github.com/ollama/ollama

---

## License

[Include your license information here]

---

## Credits

- Built with OpenAI SDK and Ollama
- Uses Qdrant vector database
- React frontend, Go backend
