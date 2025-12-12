# syntax=docker/dockerfile:1.4
# hadolint global ignore=DL3008
# =============================================================================
# Kodflow Base Development Image
# =============================================================================
# Single base image with all tools + Node.js (for Claude Code)
# Languages are added via devcontainer features at runtime
#
# Build:
#   docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/kodflow/devcontainer:latest .
# =============================================================================

ARG BASE_IMAGE=mcr.microsoft.com/devcontainers/base:ubuntu-24.04
# hadolint ignore=DL3006
FROM ${BASE_IMAGE}

ARG TARGETARCH
ARG BUILDKIT_INLINE_CACHE=1

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8

# =============================================================================
# System Dependencies
# =============================================================================
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && \
    apt-get install -y --no-install-recommends \
        curl ca-certificates gnupg gnupg2 lsb-release \
        build-essential g++ gcc make cmake pkg-config \
        git jq yq zsh unzip zip tar xz-utils file \
        libssl-dev libpam0g-dev software-properties-common apt-transport-https \
        libffi-dev libbz2-dev libreadline-dev libsqlite3-dev \
        libncurses5-dev libncursesw5-dev liblzma-dev zlib1g-dev \
        libgdbm-dev libyaml-dev && \
    apt-get clean

# =============================================================================
# HashiCorp Tools (Terraform, Vault, Consul, Nomad, Packer)
# =============================================================================
RUN curl -fsSL https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" > /etc/apt/sources.list.d/hashicorp.list && \
    apt-get update && apt-get install -y --no-install-recommends terraform vault consul nomad packer terraform-ls && \
    rm -rf /var/lib/apt/lists/*

# =============================================================================
# Cloud CLIs (AWS, GCP, Azure)
# =============================================================================
# AWS CLI v2
RUN ARCH_AWS=$([ "$TARGETARCH" = "arm64" ] && echo "aarch64" || echo "x86_64") && \
    curl -fsSL "https://awscli.amazonaws.com/awscli-exe-linux-${ARCH_AWS}.zip" -o /tmp/awscliv2.zip && \
    unzip -q /tmp/awscliv2.zip -d /tmp && /tmp/aws/install && rm -rf /tmp/aws /tmp/awscliv2.zip

# Google Cloud SDK
RUN curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" > /etc/apt/sources.list.d/google-cloud-sdk.list && \
    apt-get update && apt-get install -y --no-install-recommends google-cloud-cli google-cloud-cli-gke-gcloud-auth-plugin && \
    rm -rf /var/lib/apt/lists/*

# Azure CLI
RUN curl -fsSL https://aka.ms/InstallAzureCLIDeb | bash

# =============================================================================
# Kubernetes Tools (kubectl, Helm)
# =============================================================================
RUN ARCH_K8S=$([ "$TARGETARCH" = "arm64" ] && echo "arm64" || echo "amd64") && \
    curl -fsSL "https://dl.k8s.io/release/$(curl -sL https://dl.k8s.io/release/stable.txt)/bin/linux/${ARCH_K8S}/kubectl" -o /usr/local/bin/kubectl && \
    chmod +x /usr/local/bin/kubectl && \
    curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# =============================================================================
# Ansible
# =============================================================================
RUN add-apt-repository --yes --update ppa:ansible/ansible && \
    apt-get install -y --no-install-recommends ansible && rm -rf /var/lib/apt/lists/*

# =============================================================================
# Bazelisk (Bazel version manager)
# =============================================================================
RUN ARCH_BAZEL=$([ "$TARGETARCH" = "arm64" ] && echo "arm64" || echo "amd64") && \
    curl -fsSL "https://github.com/bazelbuild/bazelisk/releases/latest/download/bazelisk-linux-${ARCH_BAZEL}" -o /usr/local/bin/bazel && \
    chmod +x /usr/local/bin/bazel

# =============================================================================
# User Setup
# =============================================================================
USER vscode
WORKDIR /home/vscode

# Create cache and config directories
RUN mkdir -p .cache .config .local/bin .local/share .zsh_history_dir \
    .cache/{terraform,helm,aws,bazel,npm,yarn,pnpm,nvm} \
    .config/{aws,gcloud,azure} .kube \
    .claude .config/@anthropic .cache/@anthropic .local/share/@anthropic

# Oh My Zsh + Powerlevel10k (skip if already installed in base image)
RUN if [ ! -d "$HOME/.oh-my-zsh" ]; then \
        sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended; \
    fi && \
    rm -rf ~/.oh-my-zsh/custom/themes/powerlevel10k && \
    git clone --depth=1 https://github.com/romkatv/powerlevel10k.git ~/.oh-my-zsh/custom/themes/powerlevel10k && \
    rm -rf ~/.oh-my-zsh/custom/plugins/zsh-autosuggestions && \
    git clone --depth=1 https://github.com/zsh-users/zsh-autosuggestions.git ~/.oh-my-zsh/custom/plugins/zsh-autosuggestions && \
    rm -rf ~/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting && \
    git clone --depth=1 https://github.com/zsh-users/zsh-syntax-highlighting.git ~/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting

# Copy p10k configuration
COPY --chown=vscode:vscode .p10k.zsh /home/vscode/.p10k.zsh

# Configure zsh
RUN sed -i 's/^ZSH_THEME=.*/ZSH_THEME="powerlevel10k\/powerlevel10k"/' ~/.zshrc && \
    sed -i 's/^plugins=.*/plugins=(git docker aws gcloud kubectl helm terraform zsh-autosuggestions zsh-syntax-highlighting)/' ~/.zshrc && \
    printf '\n[[ ! -f ~/.p10k.zsh ]] || source ~/.p10k.zsh\nexport HISTFILE=~/.zsh_history_dir/.zsh_history\nexport HISTSIZE=50000\nexport SAVEHIST=50000\n[[ -f ~/.kodflow-env.sh ]] && source ~/.kodflow-env.sh\n' >> ~/.zshrc

# =============================================================================
# Node.js + Claude Code
# =============================================================================
ENV NVM_DIR=/home/vscode/.cache/nvm \
    npm_config_cache=/home/vscode/.cache/npm

RUN mkdir -p $NVM_DIR ~/.local/share/npm-global && \
    curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/master/install.sh | bash && \
    . "$NVM_DIR/nvm.sh" && \
    nvm install lts/* && \
    nvm use lts/* && \
    nvm alias default lts/* && \
    npm install -g npm@latest @anthropic-ai/claude-code && \
    npm config set prefix ~/.local/share/npm-global && \
    ln -sf "$NVM_DIR/versions/node/$(ls $NVM_DIR/versions/node | head -1)/bin/node" ~/.local/bin/node && \
    ln -sf "$NVM_DIR/versions/node/$(ls $NVM_DIR/versions/node | head -1)/bin/npm" ~/.local/bin/npm && \
    ln -sf "$NVM_DIR/versions/node/$(ls $NVM_DIR/versions/node | head -1)/bin/npx" ~/.local/bin/npx && \
    ln -sf "$NVM_DIR/versions/node/$(ls $NVM_DIR/versions/node | head -1)/bin/claude" ~/.local/bin/claude

# =============================================================================
# Environment Variables
# =============================================================================
ENV TF_PLUGIN_CACHE_DIR=/home/vscode/.cache/terraform \
    HELM_CACHE_HOME=/home/vscode/.cache/helm \
    KUBECONFIG=/home/vscode/.kube/config \
    BAZEL_USER_ROOT=/home/vscode/.cache/bazel \
    CLAUDE_CONFIG_DIR=/home/vscode/.claude \
    PATH=/home/vscode/.local/bin:/home/vscode/.local/share/npm-global/bin:$PATH

WORKDIR /workspace

# =============================================================================
# Labels
# =============================================================================
LABEL org.opencontainers.image.source="https://github.com/kodflow/.repository" \
      org.opencontainers.image.description="Kodflow Base Development Image - All tools + Claude Code included" \
      org.opencontainers.image.licenses="MIT"
