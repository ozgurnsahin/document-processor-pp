# Document Processor - Microservices Backend System

A medium-scale document processing backend system built with microservices architecture, designed as a portfolio project to demonstrate modern backend development practices and inter-service communication.

## 🎯 Project Overview

This project implements a complete document processing pipeline that can ingest documents, extract and process content, generate embeddings, and provide search capabilities. The system is built using a microservices architecture with Go and Python services communicating via gRPC.

## 🏗️ Architecture

### Core Services

**Document Ingestion Service (Go)**
- Handles file uploads through HTTP endpoints
- Performs initial document validation
- Manages document metadata and storage
- Communicates with processing service via gRPC

**Document Processing Service (Python)**
- Extracts text and metadata from documents (PDF, TXT)
- Generates embeddings using OpenAI models
- Implements intelligent text chunking strategies
- Processes documents asynchronously via gRPC

**Storage Layer**
- **MongoDB**: Document metadata, processed chunks, and vector embeddings
- **File System**: Original document storage

### Communication
- **gRPC**: Inter-service communication with Protocol Buffers
- **HTTP/REST**: Client-facing APIs and file upload endpoints

## 🛠️ Technology Stack

- **Backend Languages**: Go, Python
- **Communication**: gRPC, Protocol Buffers
- **Database**: MongoDB with vector storage
- **Web Framework**: FastAPI (Python), Native HTTP (Go)
- **AI/ML**: OpenAI Embeddings API
- **Containerization**: Docker, Docker Compose
- **Text Processing**: LangChain, PyMuPDF

## 🚀 Features

### Project Implementation
- ✅ Multi-format document ingestion (PDF, TXT)
- ✅ Intelligent text chunking and processing
- ✅ Vector embedding generation using OpenAI models
- ✅ AI-powered document search with similarity matching
- ✅ Microservices architecture with gRPC communication
- ✅ Docker containerization with multi-service orchestration
- ✅ MongoDB storage with vector indexing
- ✅ Interactive web interface with upload and search functionality
- ✅ Real-time document processing and feedback


## 📁 Project Structure

```
document-processor-pp/
├── document-ingestion/          # Go service for file handling
│   ├── main.go                 # HTTP server and routing
│   ├── reader/                 # File processing logic
│   ├── processor/              # gRPC client
│   ├── storage/                # MongoDB integration
│   └── static/                 # Simple HTML frontend
├── document-process/           # Python processing service
│   ├── app/                    # FastAPI application
│   ├── functions/              # Core processing logic
│   └── grpc_server/           # gRPC server implementation
├── proto/                      # Protocol Buffer definitions
└── docker-compose.yaml        # Multi-service orchestration
```

## 🔧 Setup and Installation

### Prerequisites
- Docker and Docker Compose
- OpenAI API key
- MongoDB instance

### Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/ozgurnsahin/document-processor-pp.git
   cd document-processor-pp
   ```

2. **Set up environment variables**
   ```bash
   # Create .env file with:
   OPENAI_API_KEY=your_openai_api_key
   MONGODB_STRING=your_mongodb_connection_string
   MONGODB_DB=docDev
   ```

3. **Start the services**
   ```bash
   docker-compose up -d
   ```

4. **Access the application**
   - Upload interface: http://localhost:8080
   - Processing service: http://localhost:8081
   - Upload documents (PDF/TXT, up to 20MB)
   - Search through processed documents using AI similarity
   - Health checks available at `/health` endpoints

## 🎓 Learning Objectives

This project demonstrates:
- **Microservices Architecture**: Service separation and communication patterns
- **gRPC Implementation**: Type-safe inter-service communication with Protocol Buffers
- **Document Processing**: Text extraction and intelligent chunking strategies
- **Vector Embeddings**: AI-powered document understanding and similarity search
- **Containerization**: Multi-service Docker deployment and orchestration
- **Database Design**: Document and vector storage with efficient search capabilities
- **Full-Stack Integration**: Backend services with interactive web interface


## 🤝 Contributing

This is a portfolio project focused on learning and skill development. Feel free to explore the code, suggest improvements, or use it as a reference for your own microservices projects.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Note**: This project is designed as a learning exercise and portfolio piece. It demonstrates various backend technologies and architectural patterns in a practical, real-world scenario.
