FROM apache/tika:3.3.0.0-full

USER root
ADD --chmod=0644 --checksum=sha256:a5fcb6f0db1e1d6d8522f39db4e848f05984669172e584e8d76b6b3141e1f730 \
    https://raw.githubusercontent.com/tesseract-ocr/tessdata_fast/65727574dfcd264acbb0c3e07860e4e9e9b22185/chi_sim.traineddata \
    /usr/share/tesseract-ocr/5/tessdata/chi_sim.traineddata

COPY tika-config.xml /tika-config.xml

USER 35002:35002
CMD ["-c", "/tika-config.xml"]
