import logging
import grpc
from concurrent import futures

from app.functions.embedding_functions import EmbeddingFunctions
from app.functions.process_functions import ProcessorFunctions
from proto import document_process_pb2 as pb2
from proto import document_process_pb2_grpc as pb2_grpc

logger = logging.getLogger(__name__)


class ProcessorService:
    def __init__(self):
        self.processor = ProcessorFunctions()
        self.embedder = EmbeddingFunctions()

    def ProcessDocument(self, request, context):
        try:
            document_id = request.document_id
            filename = request.filename
            content = request.content
            content_type = request.content_type

            logger.info(
                f"Received document: {document_id}, {filename}, size: {len(content)} bytes"
            )

            processed_data = self.processor.read_file(
                file_bytes=content, content_type=content_type
            )

            logger.info(f"Processed document: {document_id}")

            embeddings = self.embedder.create_embeddings_from_sentences(
                sentences=processed_data["sentences"]
            )

            logger.info(f"Embeded sentences of document: {document_id}")

            response = pb2.ProcessResponse(document_id=document_id, status="completed")

            for sen, embed in zip(processed_data["sentences"], embeddings):
                chunk = pb2.ProcessedChunk(text=sen, vector=embed)
                response.chunks.append(chunk)

            logger.info(
                f"Successfully processed document {document_id}: {len(processed_data['sentences'])} chunks"
            )

            return response
        except Exception as e:
            logger.error(f"Error processing document {request.document_id}: {e}")
            return pb2.ProcessResponse(
                document_id=request.document_id, status="failed", error=e
            )

    def CreateEmbeddingsFromInput(self, request, context):
        try:
            text = request.text
            logger.info("Creating embedding for text")

            embeddings = self.embedder.create_embedding_from_input(text)

            response = pb2.EmbeddingResponse(vector=embeddings)

            logger.info("Successfully created embedding")
            return response
        except Exception as e:
            logger.error(f"error creating embedding: {e}")
            return pb2.EmbeddingResponse(error=str(e))


class GRPCServer:
    def __init__(self, port=50052, max_workers=10):
        self.port = port
        self.max_workers = max_workers
        self.server = None
        self.service = ProcessorService()

    def start(self):
        options = [
            ("grpc.max_receive_message_length", 10 * 1024 * 1024),  # 10MB
            ("grpc.max_send_message_length", 10 * 1024 * 1024),  # 10MB
        ]

        self.server = grpc.server(
            futures.ThreadPoolExecutor(max_workers=self.max_workers), options=options
        )

        pb2_grpc.add_DocumentProcessorServiceServicer_to_server(
            self.service, self.server
        )

        server_addr = f"[::]:{self.port}"
        self.server.add_insecure_port(server_addr)
        self.server.start()

        logger.info(f"gRPC server started on {server_addr}")
        return self.server

    def stop(self, grace=None):
        if self.server:
            self.server.stop(grace)
            self.server = None
            logger.info("gRPC server stopped")

    def __del__(self):
        self.stop()
