from typing import List, Dict
from langchain_text_splitters import (
    RecursiveCharacterTextSplitter,
    MarkdownHeaderTextSplitter,
)
import fitz
import pymupdf4llm
import io
import re


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

    def read_file(self, file_bytes: bytes, file_name: str):
        file_type = file_name.split(".")[-1]

        if file_type == "pdf":
            return self._process_pdf(file_bytes=file_bytes)
        elif file_type in ["txt", "rtf"]:
            return self._process_txt(file_bytes=file_bytes)
        else:
            raise ValueError(f"Unsupported file type: {file_type}")

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
                    if not len(split.page_content) > 5 or re.match(
                        r"^[^\w]*$", split.page_content
                    ):
                        continue
                    elif (
                        split.metadata and split.page_content[0] == "#"
                    ):  # Header detection
                        pdf_data["sentences"].append(split.page_content)
                        pdf_data["page_number"].append(i + 1)
                    elif (
                        split.page_content[0] == "*"
                        and split.page_content[-1] == "*"
                        and (
                            re.match(
                                r"(\*{2,})(\d+(?:\.\d+)*)\s*(\*{2,})?(.*)$",
                                split.page_content,
                            )
                            or re.match(
                                r"(\*{1,3})?([A-Z][a-zA-Z\s\-]+)(\*{1,3})?$",
                                split.page_content,
                            )
                        )
                    ):  # Sub-Header and Header variant detection
                        pdf_data["sentences"].append(split.page_content)
                        pdf_data["page_number"].append(i + 1)
                    elif (
                        split.page_content[0] == "|" and split.page_content[-1] == "|"
                    ):  # Table detection
                        pdf_data["sentences"].append(split.page_content)
                        pdf_data["page_number"].append(i + 1)
                    else:
                        pdf_data["sentences"].append(split.page_content)
                        pdf_data["page_number"].append(i + 1)
        return pdf_data

    def _process_txt(self, file_bytes: bytes):
        text_data = {"sentences": [], "page_number": []}
        text = file_bytes.decode("utf-8", errors="ignore")
        valid_sentences = text.split(".")
        for sentence in valid_sentences:
            text_data["sentences"].extend(sentence.strip())
        text_data["page_number"].extend([1] * len(valid_sentences))
        return text_data
