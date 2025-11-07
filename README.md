# Askara

**Askara** is an AI-powered knowledge base that enables users to upload custom documents and ask questions about their contents using advanced language models and vector databases.

Built on a robust Golang backend with a React frontend, Askara leverages OpenAI embeddings and vector databases (Pinecone or Qdrant) to provide accurate, context-aware answers from your uploaded documents.

## ✨ Features

- **Document Upload**: Support for PDF, .txt, .rtf, .docx, and .epub files
- **AI-Powered Q&A**: Ask natural language questions and get accurate answers based on your documents
- **Source Attribution**: See the specific file names and text snippets that inform each answer
- **Vector Database Options**: Choose between Pinecone or Qdrant for vector storage
- **Token Tracking**: Monitor OpenAI API usage with built-in token counting
- **User-Friendly Interface**: Drag-and-drop file uploads with a clean React UI

## 🚀 Quick Start

### Prerequisites

- **Node.js**: v19 or higher
- **Go**: v1.17 or higher
- **Poppler**: Required for PDF processing
  - Ubuntu: `sudo apt-get install -y poppler-utils`
  - macOS: `brew install poppler`

### API Keys Setup

1. **OpenAI API Key**: Create `secret/openai_api_key`
   ```bash
   echo "your_openai_api_key_here" > secret/openai_api_key
   ```

2. **Pinecone API Key**: Create `secret/pinecone_api_key`
   ```bash
   echo "your_pinecone_api_key_here" > secret/pinecone_api_key
   ```

   Note: When setting up your Pinecone index, use a vector size of `1536` (OpenAI ada-002 embedding dimension).

3. **Pinecone Endpoint**: Create `secret/pinecone_api_endpoint`
   ```bash
   echo "https://your-index.svc.region.pinecone.io" > secret/pinecone_api_endpoint
   ```

### Running the Development Environment

1. **Install dependencies**:
   ```bash
   npm install
   ```

2. **Start the Golang server** (runs on port `:8100`):
   ```bash
   npm start
   ```

3. **In another terminal, run webpack** to compile the frontend:
   ```bash
   npm run dev
   ```

4. **Access the application** at http://localhost:8100

## 📚 How It Works

### Upload Flow
1. Users upload documents via drag-and-drop or file selection
2. Text is extracted from documents (PDF, epub, docx, txt)
3. Content is chunked into manageable segments
4. Each chunk is converted to embeddings using OpenAI's API
5. Embeddings are stored in your chosen vector database with metadata

### Question Answering Flow
1. User submits a question
2. Question is converted to an embedding vector
3. Vector database retrieves the most relevant document chunks
4. OpenAI generates an answer using the retrieved context
5. Answer is displayed with source attribution and token usage

## 🏗️ Architecture

- **Backend**: Golang web server (`vault-web-server/`)
- **Frontend**: React with webpack bundling
- **Vector DB**: Pluggable interface supporting Pinecone and Qdrant
- **Embeddings**: OpenAI ada-002 (1536 dimensions)
- **Chunking**: Token-based chunking using tiktoken

## 📝 API Endpoints

- `POST /upload` - Upload and process documents
- `POST /api/question` - Submit questions and get answers

See `vault-web-server/main.go` for full API documentation.

## 🔧 Configuration

### File Size Limits
Default limits are set in `vault-web-server/postapi/fileupload.go`:
- Individual file: 25 MB
- Total upload: 50 MB

Adjust `MAX_FILE_SIZE` and `MAX_TOTAL_UPLOAD_SIZE` as needed.

### Vector Database
Switch between Pinecone and Qdrant by modifying the initialization in `vault-web-server/main.go`.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🙏 Attribution

**Askara** is based on [OP Vault](https://github.com/pashpashpash/vault-ai) by [pashpashpash](https://github.com/pashpashpash).

The original OP Vault project provided the foundation for this knowledge base system, including:
- Core architecture for document processing and embeddings
- Integration patterns for OpenAI and vector databases
- React frontend components and user interface

This project has been modified and is maintained by [itaiguardiola](https://github.com/itaiguardiola).

---

**Original Project**: [pashpashpash/vault-ai](https://github.com/pashpashpash/vault-ai)
**License**: MIT
**Maintainer**: itaiguardiola

For more information about RAG (Retrieval-Augmented Generation) with vector databases, see:
- [Pinecone: OpenAI Generative Q&A](https://www.pinecone.io/learn/openai-gen-qa/)
