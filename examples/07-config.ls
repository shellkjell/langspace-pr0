# LangSpace Configuration and Providers
# Setting up LLM providers and global configuration

# Global configuration block
config {
  # Default model for agents that don't specify one
  default_model: "claude-sonnet-4-20250514"

  # Default temperature
  default_temperature: 0.7

  # Project root directory (for relative paths)
  project_root: "."

  # Provider configurations
  providers: {
    anthropic: {
      api_key: env("ANTHROPIC_API_KEY")

      # Optional: custom base URL (for proxies)
      # base_url: "https://anthropic-proxy.example.com"

      # Rate limiting
      max_requests_per_minute: 50
    }

    openai: {
      api_key: env("OPENAI_API_KEY")
      organization: env("OPENAI_ORG_ID")
    }

    # Local model via Ollama
    ollama: {
      base_url: "http://localhost:11434"

      # Model mapping: LangSpace name -> Ollama name
      models: {
        "local-code": "codellama:13b"
        "local-fast": "mistral:7b"
      }
    }

    # Azure OpenAI
    azure: {
      api_key: env("AZURE_OPENAI_KEY")
      endpoint: env("AZURE_OPENAI_ENDPOINT")
      api_version: "2024-02-15-preview"

      # Deployment mapping
      deployments: {
        "gpt-4o": "my-gpt4-deployment"
        "gpt-4o-mini": "my-gpt4-mini-deployment"
      }
    }

    # AWS Bedrock
    bedrock: {
      region: "us-east-1"
      # Uses AWS credentials from environment/config

      models: {
        "claude-sonnet": "anthropic.claude-3-5-sonnet-20241022-v2:0"
      }
    }
  }

  # Logging configuration
  logging: {
    level: "info"
    format: "json"
    output: "stderr"

    # Log all LLM requests/responses (for debugging)
    log_llm_calls: env("LANGSPACE_DEBUG") == "true"
  }

  # Caching configuration
  cache: {
    enabled: true
    backend: "sqlite"
    path: ".langspace/cache.db"
    ttl: "24h"
  }

  # Telemetry (optional)
  telemetry: {
    enabled: false
    endpoint: "https://telemetry.example.com"
  }
}

# Environment-specific configuration
# Can be selected with: langspace run --env production

env "development" {
  config {
    default_model: "ollama/local-fast"

    logging: {
      level: "debug"
      log_llm_calls: true
    }
  }
}

env "production" {
  config {
    default_model: "claude-sonnet-4-20250514"

    logging: {
      level: "warn"
      log_llm_calls: false
    }

    cache: {
      backend: "redis"
      url: env("REDIS_URL")
    }
  }
}

# Per-agent provider override examples
agent "fast-local" {
  model: "ollama/local-fast"  # Uses local Ollama
  instruction: "You are a quick assistant."
}

agent "premium" {
  model: "gpt-4o"  # Uses OpenAI
  instruction: "You are a thorough analyst."
}

agent "bedrock-claude" {
  model: "bedrock/claude-sonnet"  # Uses AWS Bedrock
  instruction: "You are helpful."
}
