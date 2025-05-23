from typing import List, Dict
from langchain_text_splitters import (
    RecursiveCharacterTextSplitter,
    MarkdownHeaderTextSplitter,
)
import fitz
import pymupdf4llm
import io


class ProcessorFunctions:
    def __init__(self):
        self.text_splitters = RecursiveCharacterTextSplitter(
            chunk_size=200,
            chunk_overlap=50,
            length_function=len,
            is_separator_regex=False,
        )
        self.headers_to_split_on = [
            ("#", "Header 1"),
            ("##", "Header 2"),
            ("###", "Header 3"),
            ("####", "Header 4"),
        ]
        self.markdown_splitter = MarkdownHeaderTextSplitter(
            self.headers_to_split_on, strip_headers=False, return_each_line=True
        )

    def read_file(self, file_bytes: bytes, content_type: str):
        if content_type == "application/pdf":
            return self._process_pdf(file_bytes=file_bytes)
        elif content_type in ["text/plain; charset=utf-8", "text/rtf; charset=utf-8"]:
            return self._process_txt(file_bytes=file_bytes)
        else:
            raise ValueError(f"Unsupported file type: {content_type}")

    def process_document_content(self, file) -> List[Dict[str, any]]:
        processed_text = self.read_file(file)

        return processed_text

    def _process_pdf(self, file_bytes: bytes):
        pdf_file = io.BytesIO(file_bytes)
        pdf_data = {"sentences": [], "page_number": []}
        with fitz.open(stream=pdf_file, filetype="pdf") as pdf:
            markdown_pages = pymupdf4llm.to_markdown(
                pdf, page_chunks=True, show_progress=False, margins=0
            )
            for i, page in enumerate(markdown_pages):
                splits = self.markdown_splitter.split_text(page["text"])
                for split in splits:
                    if not len(split.page_content) > 5:
                        continue
                    else:
                        pdf_data["sentences"].append(split.page_content)
                        pdf_data["page_number"].append(i + 1)
        return pdf_data

    def _process_txt(self, file_bytes: bytes):
        text_data = {"sentences": [], "page_number": []}
        text = file_bytes.decode("utf-8", errors="ignore")
        splits = self.text_splitters.split_text(text)
        for sentence in splits:
            text_data["sentences"].append(sentence.strip())
        text_data["page_number"].extend([1] * len(splits))
        return text_data
