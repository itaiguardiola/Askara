# Askara Windows Deployment Guide

Welcome! This guide will help you set up and run Askara on Windows. Askara is an AI-powered knowledge base that lets you upload documents and ask questions about their contents using advanced language models.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Configuration](#configuration)
4. [Testing](#testing)
5. [Windows-Specific Notes](#windows-specific-notes)
6. [Troubleshooting](#troubleshooting)

---

## Prerequisites

Before you begin, install the following software on your Windows machine:

### 1. Docker Desktop for Windows

Docker is required to run Askara and its dependencies (vector database, etc.).

**Installation Steps:**

1. Download Docker Desktop from: https://www.docker.com/products/docker-desktop
2. Run the installer and follow the setup wizard
3. During installation, ensure these options are checked:
   - Use WSL 2 based engine (recommended)
   - Install required Windows components for WSL 2
4. Restart your computer when prompted
5. After restart, Docker Desktop will start automatically

**Verify Installation:**

Open PowerShell or Command Prompt and run:

```powershell
docker --version
docker run hello-world
```

You should see version information and a "Hello from Docker!" message.

### 2. Git for Windows

Git is required to clone the Askara repository.

**Installation Steps:**

1. Download Git from: https://git-scm.com/download/win
2. Run the installer with default settings
3. Choose "Use Git Bash from the Windows Command Prompt" when prompted
4. Complete the installation

**Verify Installation:**

Open PowerShell and run:

```powershell
git --version
```

### 3. Ollama (Optional - For Local Models)

If you want to use Ollama for local language models instead of OpenAI, install Ollama:

**Installation Steps:**

1. Download Ollama for Windows from: https://ollama.ai
2. Run the installer
3. Ollama will start automatically after installation

**Verify Installation:**

Open PowerShell and run:

```powershell
ollama --version
```

### 4. Node.js (Required for Development)

Node.js is needed to build the frontend and run development tools.

**Installation Steps:**

1. Download Node.js v19+ from: https://nodejs.org/ (choose LTS version)
2. Run the installer with default settings
3. When prompted, allow installation of necessary tools

**Verify Installation:**

Open PowerShell and run:

```powershell
node --version
npm --version
```

### 5. Go (Required for Building Backend)

Go is needed to compile the backend server.

**Installation Steps:**

1. Download Go from: https://golang.org/dl/
2. Download the Windows .msi file (amd64 for 64-bit systems)
3. Run the installer
4. Go will be added to your PATH automatically

**Verify Installation:**

Open PowerShell and run:

```powershell
go version
```

### Recommended Models to Pull (for Ollama)

Once Ollama is installed, you can pull language models. Popular choices:

```powershell
# Lightweight model (fastest)
ollama pull mistral

# Balanced model (good quality and speed)
ollama pull neural-chat

# High-quality model (slower but better answers)
ollama pull dolphin-mixtral

# Check available models
ollama list
```

---

## Quick Start

### Step 1: Clone the Repository

Open PowerShell in the directory where you want to store Askara:

```powershell
git clone https://github.com/itaiguardiola/askara.git
cd askara
```

### Step 2: Create Environment File

Copy the example environment file to create your configuration:

```powershell
Copy-Item .env.example .env
```

If `.env.example` doesn't exist, create a new `.env` file:

```powershell
New-Item -Path ".env" -ItemType File
```

Then open it with your text editor and add the appropriate configuration (see [Configuration Section](#configuration) below).

### Step 3: Install Dependencies

```powershell
npm install
```

This installs all Node.js dependencies needed for the frontend.

### Step 4: Pull Ollama Models (Optional)

If using Ollama for local language models:

```powershell
# Start Ollama service first (it should auto-start)
ollama serve

# In another PowerShell window, pull a model:
ollama pull mistral
```

Or if you prefer OpenAI models, set up your API key in the `.env` file instead (see Configuration section).

### Step 5: Start Docker Containers

Depending on your configuration, you need to run the appropriate vector database:

**Option A: Using Qdrant (Recommended for Windows)**

```powershell
# In PowerShell, from the project directory:
docker pull qdrant/qdrant
docker run -d -p 6333:6333 -v qdrant_storage:/qdrant/storage qdrant/qdrant
```

**Option B: Using Pinecone (Cloud-based)**

Pinecone is a cloud service - no local Docker needed. Skip this step and configure your `.env` file with Pinecone credentials.

### Step 6: Start the Application

In PowerShell, from the project directory:

```powershell
npm start
```

The server will start on port 8100.

### Step 7: Build Frontend (Development)

In another PowerShell window, from the project directory:

```powershell
npm run dev
```

This starts webpack in development mode for hot-reloading.

### Step 8: Access Askara

Open your web browser and navigate to:

```
http://localhost:8100
```

You should see the Askara interface. You're ready to use it!

---

## Configuration

### Environment Variables

The `.env` file controls how Askara behaves. Here's what each variable does:

#### 1. **OPENAI_API_KEY** (Optional, for OpenAI models)

If you want to use OpenAI's models for answers and embeddings:

```
OPENAI_API_KEY=sk-your-openai-api-key-here
```

Get your API key from: https://platform.openai.com/api-keys

#### 2. **OLLAMA_API_ENDPOINT** (Optional, for local Ollama models)

If you want to use local models via Ollama:

```
OLLAMA_API_ENDPOINT=http://localhost:11434
```

By default, Ollama runs on port 11434. Keep this as-is unless you changed Ollama's port.

#### 3. **Vector Database Configuration**

Choose ONE of the following:

**Option A: Qdrant (Recommended for Windows)**

```
QDRANT_API_ENDPOINT=http://localhost:6333
```

Start Qdrant with Docker (see Quick Start Step 5).

**Option B: Pinecone (Cloud-based)**

```
PINECONE_API_ENDPOINT=https://your-index-abc123.svc.region.pinecone.io
PINECONE_API_KEY=your-pinecone-api-key
```

Get credentials from: https://www.pinecone.io/

#### 4. **Server Port** (Optional)

```
PORT=8100
```

Default is 8100. Change this if port 8100 is already in use.

### Example Configuration Files

**Example 1: Using OpenAI + Qdrant**

```
OPENAI_API_KEY=sk-xxxxxxxxxx
QDRANT_API_ENDPOINT=http://localhost:6333
PORT=8100
```

**Example 2: Using Ollama + Qdrant (Local Only)**

```
OLLAMA_API_ENDPOINT=http://localhost:11434
QDRANT_API_ENDPOINT=http://localhost:6333
PORT=8100
```

**Example 3: Using OpenAI + Pinecone (Cloud)**

```
OPENAI_API_KEY=sk-xxxxxxxxxx
PINECONE_API_ENDPOINT=https://your-index.svc.region.pinecone.io
PINECONE_API_KEY=your-api-key
PORT=8100
```

### Switching Between Ollama and OpenAI

**To switch from OpenAI to Ollama:**

1. Edit `.env` file
2. Comment out or remove: `OPENAI_API_KEY=...`
3. Add: `OLLAMA_API_ENDPOINT=http://localhost:11434`
4. Pull a model: `ollama pull mistral` (in PowerShell)
5. Restart the application

**To switch from Ollama to OpenAI:**

1. Edit `.env` file
2. Comment out or remove: `OLLAMA_API_ENDPOINT=...`
3. Add: `OPENAI_API_KEY=sk-your-key-here`
4. Restart the application

### Model Selection

**OpenAI Models:**

The application automatically uses:
- `text-embedding-3-small` for document embeddings
- `gpt-4o` or `gpt-4-turbo` for question answering

**Ollama Models:**

Available models (pull with `ollama pull model-name`):

| Model | Speed | Quality | RAM |
|-------|-------|---------|-----|
| mistral | Very Fast | Good | 7GB |
| neural-chat | Fast | Very Good | 13GB |
| dolphin-mixtral | Moderate | Excellent | 26GB |
| llama2 | Moderate | Good | 7GB |
| dolphin-phi | Very Fast | Moderate | 3GB |

For a typical Windows machine with 16GB RAM, use:
- **Mistral** - Recommended, balanced
- **Neural-Chat** - Better quality if you have the RAM

### Common Configuration Issues

**Issue: "MISSING OPENAI API KEY"**

Solution: Make sure `OPENAI_API_KEY` is set in `.env`, or set the environment variable:

```powershell
$env:OPENAI_API_KEY="sk-your-key-here"
npm start
```

**Issue: "NO VECTOR DB CONFIGURED"**

Solution: Set either `QDRANT_API_ENDPOINT` or `PINECONE_API_ENDPOINT` in `.env`:

```
QDRANT_API_ENDPOINT=http://localhost:6333
```

**Issue: "Connection refused" for Qdrant**

Solution: Make sure Qdrant is running:

```powershell
docker ps | findstr qdrant
```

If not running, start it:

```powershell
docker run -d -p 6333:6333 -v qdrant_storage:/qdrant/storage qdrant/qdrant
```

---

## Testing

### Upload a Test Document

1. Open http://localhost:8100 in your browser
2. Look for the "Upload Documents" section
3. Click "Choose Files" or drag and drop a PDF or text file
4. Supported formats: PDF, .txt, .docx, .epub, .rtf
5. Click "Upload" and wait for processing (you'll see a progress indicator)

**Test Document Ideas:**

- Save a Wikipedia article as a text file
- Download a research paper as PDF
- Use a sample document from your computer

### Ask Sample Questions

Once documents are uploaded:

1. Go to the "Ask Questions" section
2. Type a question related to your document, like:
   - "What is the main topic of this document?"
   - "Summarize the key points"
   - "What does the document say about [specific topic]?"
3. Click "Ask" or press Enter
4. Wait for the response (usually 5-30 seconds depending on model)

**Expected Results:**

- You should see an answer generated from your document
- Below the answer, "Sources" section shows which document chunks were used
- Token count shows API usage (if using OpenAI)

### Verify Embeddings Are Working

Check that the vector database has processed your documents:

**For Qdrant:**

```powershell
# Open a browser and go to Qdrant admin panel:
http://localhost:6333/dashboard

# Or use curl to check collections:
curl http://localhost:6333/collections
```

**For Pinecone:**

1. Go to https://app.pinecone.io
2. Select your index
3. Check that statistics show documents have been uploaded
4. You should see vector count increasing as documents are processed

### View Upload Statistics

1. In Askara UI, look for "Document Management" or "Statistics"
2. You should see:
   - Total documents uploaded
   - Total text chunks created
   - Storage usage
3. You can delete documents from this section too

### Troubleshooting Tests

**Document upload fails:**
- Check file size (max 25MB per file)
- Try a simple text file first
- Check Docker/database is running

**Questions don't return answers:**
- Ensure documents are uploaded first
- Check that your API key is valid
- View browser console (F12) for error messages
- Check server logs in PowerShell

**Embeddings not working:**
- For Qdrant: Check `http://localhost:6333/dashboard`
- For Pinecone: Check the Pinecone dashboard
- Verify API endpoint is correct in `.env`

---

## Windows-Specific Notes

### 1. Firewall Settings

Windows Defender Firewall might block Askara. To allow it:

**For Docker:**

1. Open "Windows Defender Firewall" (search in Start menu)
2. Click "Allow an app through firewall"
3. Click "Change settings"
4. Look for "Docker Desktop" in the list
5. Check both "Private" and "Public" boxes
6. Click "OK"

**For Ollama:**

1. Follow the same steps
2. Find and enable "Ollama" in the firewall list

**To test if blocked:**

```powershell
# Test connection to Docker
docker ps

# Test connection to Qdrant
curl http://localhost:6333/collections

# Test connection to Ollama
curl http://localhost:11434/api/tags
```

### 2. Docker Networking on Windows

Docker on Windows uses Hyper-V or WSL 2. Keep these points in mind:

**WSL 2 (Recommended):**

- Better performance
- Access Docker containers via `localhost:port`
- Volume mounts work better
- Default setup with modern Docker Desktop

**Hyper-V:**

- Older approach
- May need to use `host.docker.internal` instead of `localhost`
- Less optimal performance

**To check your setup:**

```powershell
docker system info | findstr "Storage Driver"
```

Look for "WSL 2" in the output (you should see it if using WSL 2).

### 3. Volume Permissions

Docker on Windows handles permissions differently than Linux. If you get permission errors:

**Quick Fix:**

```powershell
# Reset Docker
docker system prune -a

# Restart Docker Desktop:
# Right-click Docker icon in taskbar > Restart
```

**For Qdrant persistence:**

The `-v qdrant_storage:/qdrant/storage` flag in the Docker command creates a Docker volume that persists data even if the container stops.

To see your volumes:

```powershell
docker volume ls
```

### 4. Performance Tips for RTX 3090

If you have an NVIDIA RTX 3090, you can speed up model inference:

**Enable GPU in Docker:**

1. Install NVIDIA Container Runtime: https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html
2. When starting Ollama in Docker, add:

```powershell
docker run --gpus all -d -p 11434:11434 ollama/ollama
```

**Performance Tuning:**

- Use smaller models (mistral) for faster responses
- Disable background applications while running inference
- Allocate enough RAM: Go to Docker Desktop Settings > Resources
  - CPU: 6+ cores recommended
  - Memory: 8GB+ (more for larger models)
  - Disk: 50GB+ free space

**Monitor GPU Usage:**

With RTX 3090, monitor GPU usage while running:

```powershell
# If NVIDIA drivers are installed:
nvidia-smi

# This shows GPU memory and utilization
```

### 5. Port Management

If port 8100 is already in use:

**Find what's using port 8100:**

```powershell
Get-Process -Id (Get-NetTCPConnection -LocalPort 8100).OwningProcess
```

**Use a different port:**

1. Edit `.env` file
2. Change: `PORT=8101` (or another available port)
3. Restart the application
4. Access at: `http://localhost:8101`

### 6. WSL 2 Path Considerations

If using WSL 2 for development, remember:

- Clone repository in WSL home directory, not Windows paths
- WSL paths: `/home/username/askara`
- NOT: `C:/Users/username/askara` (slower performance)

### 7. Database File Size

As you upload documents, the database grows:

**Qdrant storage location:**

```powershell
# See Docker volume info:
docker volume inspect qdrant_storage

# Location will be shown in "Mountpoint" field
```

**To manage storage:**

- Archive old documents periodically
- Delete unused documents in Askara UI
- Monitor disk space with: `Get-Volume`

### 8. Checking Memory Usage

Monitor system resources while running Askara:

```powershell
# Show real-time stats
Get-Process | Where-Object {$_.Name -like "*docker*" -or $_.Name -like "*ollama*" -or $_.Name -like "*node*"}  | Format-Table Name, WorkingSet
```

---

## Troubleshooting

### Common Issues and Solutions

#### Issue: Docker Desktop Won't Start

**Symptoms:** Docker Desktop icon doesn't appear, or crashes on startup

**Solutions:**

1. **Restart Docker Desktop:**
   - Right-click Docker icon in taskbar
   - Select "Restart"
   - Wait 2-3 minutes for startup

2. **Reset Docker (WARNING: deletes all containers):**
   ```powershell
   docker system prune -a --volumes
   docker system reset
   ```

3. **Update Docker:**
   - Open Docker Desktop
   - Check for updates in Settings > Update
   - Install latest version

4. **Enable Hyper-V:**
   - Search "Turn Windows features on or off"
   - Check "Hyper-V"
   - Restart computer

#### Issue: "Cannot connect to Docker daemon"

**Symptoms:** Error when running `docker ps` or starting containers

**Solutions:**

```powershell
# Start Docker Desktop service
Start-Service docker

# Or restart it completely
Restart-Service docker
```

#### Issue: Qdrant Container Won't Start

**Symptoms:** Port 6333 is accessible but no data

**Solutions:**

```powershell
# Check if container is running
docker ps

# If not running, check logs
docker logs $(docker ps -a | findstr "qdrant" | findstr "{cut -f1}")

# Restart container
docker restart <container_id>

# Or delete and recreate
docker rm -f qdrant
docker run -d -p 6333:6333 -v qdrant_storage:/qdrant/storage qdrant/qdrant
```

#### Issue: Ollama Models Won't Download

**Symptoms:** `ollama pull mistral` hangs or fails

**Solutions:**

1. **Check Ollama is running:**
   ```powershell
   Get-Process ollama
   ```

2. **Restart Ollama:**
   - Close Ollama system tray icon
   - Kill any remaining processes: `Stop-Process -Name ollama`
   - Start Ollama again from Start menu

3. **Check disk space:**
   - Models are large (5-50GB)
   - Run: `Get-Volume` to check free space
   - Delete or move files if needed

#### Issue: "Connection refused" errors

**Symptoms:** Can't connect to localhost services

**Solutions:**

1. **Check if services are running:**
   ```powershell
   # Check Qdrant
   curl http://localhost:6333

   # Check Ollama
   curl http://localhost:11434

   # Check Node app
   curl http://localhost:8100
   ```

2. **Restart services:**
   ```powershell
   # Restart Qdrant
   docker restart <container_id>

   # Restart Ollama
   Stop-Process -Name ollama
   # Then open Ollama again
   ```

#### Issue: High Memory or CPU Usage

**Symptoms:** Computer is very slow, freezes

**Solutions:**

1. **Check what's using resources:**
   ```powershell
   Get-Process | Sort-Object WorkingSet -Descending | Select-Object -First 10
   ```

2. **Reduce model size:**
   - Use smaller Ollama models (mistral instead of dolphin-mixtral)
   - Reduce Docker resource limits

3. **Close unnecessary applications:**
   - Close browsers, IDEs, other apps
   - Stop background services

4. **Increase system resources:**
   - Go to Docker Desktop Settings > Resources
   - Increase CPU and Memory limits
   - Restart Docker

#### Issue: "Port already in use"

**Symptoms:** Error when starting the app

**Solutions:**

```powershell
# Find what's using the port
Get-NetTCPConnection -LocalPort 8100

# Kill the process
Stop-Process -Id <PID> -Force

# Or use a different port in .env
PORT=8101
```

#### Issue: Uploads Fail or Documents Won't Process

**Symptoms:** Upload button doesn't work, no feedback

**Solutions:**

1. **Check server logs:**
   - Look at PowerShell window where `npm start` is running
   - Look for error messages

2. **Check file format:**
   - Supported: PDF, TXT, DOCX, EPUB, RTF
   - Try with a plain .txt file first

3. **Check file size:**
   - Individual file max: 25MB
   - Total upload max: 50MB
   - Compress or split large files

4. **Check disk space:**
   - Ensure you have 10GB+ free space for Docker volumes
   - Run: `Get-Volume`

5. **Restart the application:**
   ```powershell
   # Stop: Ctrl+C in PowerShell window
   # Wait 5 seconds
   npm start
   ```

#### Issue: Questions Don't Return Answers

**Symptoms:** Ask question button doesn't work, or returns error

**Solutions:**

1. **Ensure documents are uploaded:**
   - Check Askara UI for uploaded documents
   - If none, upload a test document first

2. **Check API keys:**
   - If using OpenAI: Verify `OPENAI_API_KEY` in `.env` is valid
   - Go to https://platform.openai.com/api-keys to check key status
   - Check account has credits available

3. **Check Ollama is running (if using local models):**
   ```powershell
   curl http://localhost:11434/api/tags
   ```

4. **Review browser console for errors:**
   - Press F12 to open Developer Tools
   - Go to Console tab
   - Look for red error messages
   - Screenshot and search for the error

### Getting Help

If you're stuck:

1. **Check the logs:**
   - Server logs in PowerShell
   - Browser console (F12)
   - Docker logs: `docker logs qdrant`

2. **Search GitHub issues:**
   - https://github.com/itaiguardiola/askara/issues

3. **Check documentation:**
   - Original README.md
   - TESTING.md for API endpoint examples

4. **Test with curl:**
   ```powershell
   # Test if server is running
   curl -v http://localhost:8100

   # Test upload endpoint
   curl -X POST http://localhost:8100/upload -F "files=@test.txt" -F "uuid=test-123"

   # Test question endpoint
   curl -X POST http://localhost:8100/api/questions `
     -F "question=What is AI?" `
     -F "model=gpt-4-turbo" `
     -F "uuid=test-123"
   ```

---

## Next Steps

Once Askara is running successfully:

1. **Upload real documents:**
   - Start with 2-3 documents to test
   - Gradually add more to build your knowledge base

2. **Explore models:**
   - Try different models to find what works best for you
   - OpenAI is better quality but costs money
   - Ollama is free but slower

3. **Optimize performance:**
   - Adjust Docker resource settings
   - Choose appropriate model sizes
   - Regular cleanup of unused documents

4. **Consider production setup:**
   - Once confident with local setup
   - Move to cloud deployment
   - Use managed vector database (Pinecone)

---

## Additional Resources

- **Askara GitHub:** https://github.com/itaiguardiola/askara
- **Docker Documentation:** https://docs.docker.com/
- **Ollama Models:** https://ollama.ai/library
- **Qdrant Documentation:** https://qdrant.tech/documentation/
- **Pinecone Documentation:** https://docs.pinecone.io/
- **OpenAI API:** https://platform.openai.com/docs/

---

## Quick Reference Commands

```powershell
# Clone and setup
git clone https://github.com/itaiguardiola/askara.git
cd askara
npm install

# Start Qdrant
docker run -d -p 6333:6333 -v qdrant_storage:/qdrant/storage qdrant/qdrant

# Pull Ollama model
ollama pull mistral

# Start server
npm start

# Start frontend build (in another PowerShell window)
npm run dev

# Access application
# Open browser to: http://localhost:8100

# Check running containers
docker ps

# View Qdrant admin panel
# Open browser to: http://localhost:6333/dashboard

# View Ollama running models
ollama list

# Stop containers
docker stop <container_id>

# Remove containers (careful!)
docker rm <container_id>

# Clean up Docker (WARNING: removes all unused containers)
docker system prune
```

---

**Last Updated:** 2025
**Askara Version:** Latest
**Status:** Windows-Compatible

Happy learning with Askara!
