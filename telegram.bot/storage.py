from __future__ import annotations

import mimetypes
import os
import re
from pathlib import Path

import boto3
from botocore.config import Config


def safe_filename(name: str) -> str:
    name = Path(name.replace("\\", "/")).name.strip().replace("\x00", "")
    name = re.sub(r"[\x00-\x1f\x7f/:*?\"<>|]+", "_", name)
    name = re.sub(r"\s+", " ", name).strip(" .")
    return name[:240] or "unnamed"


class RustFS:
    def __init__(self):
        self.endpoint = os.getenv("RUSTFS_ENDPOINT", "http://rustfs:9000")
        self.bucket = os.getenv("RUSTFS_BUCKET", "telegram-resources")
        self.region = os.getenv("RUSTFS_REGION", "us-east-1")
        access_key = os.getenv("RUSTFS_ACCESS_KEY", "")
        secret_key = os.getenv("RUSTFS_SECRET_KEY", "")
        self.client = boto3.client(
            "s3",
            endpoint_url=self.endpoint,
            aws_access_key_id=access_key,
            aws_secret_access_key=secret_key,
            region_name=self.region,
            config=Config(signature_version="s3v4", s3={"addressing_style": "path"}),
        )

    def ensure_bucket(self) -> None:
        try:
            self.client.head_bucket(Bucket=self.bucket)
        except self.client.exceptions.ClientError:
            kwargs = {"Bucket": self.bucket}
            if self.region != "us-east-1":
                kwargs["CreateBucketConfiguration"] = {"LocationConstraint": self.region}
            self.client.create_bucket(**kwargs)

    def upload(self, path: Path, object_key: str, mime_type: str | None = None) -> None:
        self.ensure_bucket()
        content_type = mime_type or mimetypes.guess_type(path.name)[0] or "application/octet-stream"
        self.client.upload_file(str(path), self.bucket, object_key, ExtraArgs={"ContentType": content_type})

    def delete(self, object_key: str) -> None:
        self.client.delete_object(Bucket=self.bucket, Key=object_key)

    def presigned_url(self, object_key: str, filename: str, inline: bool, expires: int = 300) -> str:
        disposition = "inline" if inline else "attachment"
        return self.client.generate_presigned_url(
            "get_object",
            Params={
                "Bucket": self.bucket,
                "Key": object_key,
                "ResponseContentDisposition": f'{disposition}; filename="{safe_filename(filename)}"',
            },
            ExpiresIn=expires,
        )
