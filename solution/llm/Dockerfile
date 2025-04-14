# this Dockerfile taken from https://github.com/alpine-docker/ollama/blob/master/Dockerfile.llama3.2 with some changes

FROM ollama/ollama:0.5.11 AS ollama

FROM cgr.dev/chainguard/wolfi-base

RUN apk add --no-cache libstdc++ curl

COPY --from=ollama /usr/bin/ollama /usr/bin/ollama
# COPY --from=ollama /usr/lib/ollama/runners/cpu /usr/lib/ollama/runners/cpu

# In this image, we download model directly
RUN /usr/bin/ollama serve & sleep 5 && \
      /usr/bin/ollama pull qwen2.5:0.5b

# Environment variable setup
ENV OLLAMA_HOST=REDACTED

# Expose port for the service
EXPOSE 11434

ENTRYPOINT ["/usr/bin/ollama"]
CMD ["serve"]