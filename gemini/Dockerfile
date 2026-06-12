# Start from the official, robust Google Gemini sandbox environment
FROM us-docker.pkg.dev/gemini-code-dev/gemini-cli/sandbox:0.42.0-nightly.20260428.g59b2dea0e

# Switch to root temporarily to configure system environment requirements
USER root

# Ensure git and standard terminal tools are fully up-to-date
RUN apt-get update && apt-get install -y --no-install-recommends \
  git \
  curl \
  ca-certificates \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

# Set the active working directory inside the container
WORKDIR /app

# Switch back to the standard container user context
USER node

# Default to standard bash shell execution
CMD ["/bin/bash"]
