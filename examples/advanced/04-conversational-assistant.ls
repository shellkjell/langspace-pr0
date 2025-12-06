# LangSpace Advanced Example: Conversational Assistant
# A stateful multi-turn assistant with memory, context management, RAG, and tool routing.
#
# This example demonstrates:
# - Session state and conversation memory
# - Retrieval-Augmented Generation (RAG)
# - Dynamic tool routing based on intent
# - Multi-persona support
# - Context window management
# - Follow-up and clarification handling

# ============================================================================
# CONFIGURATION
# ============================================================================

config {
  default_model: "claude-sonnet-4-20250514"

  providers: {
    anthropic: {
      api_key: env("ANTHROPIC_API_KEY")
    }
  }

  # Assistant settings
  assistant: {
    name: "Atlas"
    max_conversation_turns: 50
    memory_window: 10  # Recent turns to keep in context
    summarize_after: 20  # Summarize conversation after N turns
    embedding_model: "text-embedding-3-small"
  }

  # Vector store for RAG
  vector_store: {
    type: "pinecone"
    api_key: env("PINECONE_API_KEY")
    index: "knowledge-base"
    namespace: "default"
  }

  # Session storage
  session: {
    type: "redis"
    url: env("REDIS_URL")
    ttl: 86400  # 24 hours
  }
}

# ============================================================================
# KNOWLEDGE BASE FILES
# ============================================================================

file "system-prompt" {
  contents: ```
    You are Atlas, an intelligent assistant for {{company_name}}.

    Core Principles:
    1. Be helpful, accurate, and concise
    2. Admit when you don't know something
    3. Ask clarifying questions when the request is ambiguous
    4. Use the appropriate tools for tasks
    5. Remember context from the conversation

    Communication Style:
    - Warm but professional
    - Use the user's name when known
    - Match the user's communication style
    - Use formatting (lists, code blocks) when helpful

    Capabilities:
    - Answer questions using the knowledge base
    - Help with tasks using available tools
    - Maintain context across conversation turns
    - Escalate to humans when appropriate

    Limitations:
    - Cannot access external websites (use search tool instead)
    - Cannot make purchases or transactions
    - Cannot access personal data without permission
    - Should not provide legal, medical, or financial advice
  ```
}

file "persona-technical" {
  contents: ```
    # Technical Support Persona

    You are a technical support specialist. Adjust your communication:
    - Use precise technical terminology
    - Provide step-by-step instructions
    - Include code examples when relevant
    - Reference documentation links
    - Anticipate follow-up questions

    When troubleshooting:
    1. Gather symptoms and context
    2. Ask about recent changes
    3. Suggest diagnostic steps
    4. Provide solutions with explanations
    5. Verify the issue is resolved
  ```
}

file "persona-friendly" {
  contents: ```
    # Friendly Helper Persona

    You are a friendly, approachable assistant. Adjust your communication:
    - Use conversational language
    - Avoid jargon unless the user uses it first
    - Be encouraging and supportive
    - Use emoji sparingly but appropriately
    - Keep explanations simple and relatable

    Focus on:
    - Understanding the user's actual need
    - Providing actionable answers
    - Celebrating successes with the user
    - Making complex topics accessible
  ```
}

# ============================================================================
# TOOLS
# ============================================================================

mcp "knowledge" {
  transport: "stdio"
  command: "npx"
  args: ["-y", "@company/mcp-knowledge-base"]
}

tool "search_knowledge" {
  description: "Search the internal knowledge base for relevant information"

  parameters: {
    query: string required "Search query"
    filters: object optional "Filters like category, date, author"
    limit: number optional 5 "Maximum results to return"
  }

  handler: http {
    method: "POST"
    url: env("KNOWLEDGE_API_URL") + "/search"
    headers: {
      "Authorization": "Bearer " + env("KNOWLEDGE_API_KEY")
    }
    body: {
      query: params.query,
      filters: params.filters,
      limit: params.limit,
      include_embeddings: false
    }
  }
}

tool "create_ticket" {
  description: "Create a support ticket for issues requiring human attention"

  parameters: {
    title: string required "Ticket title"
    description: string required "Detailed description"
    priority: string optional "normal" "low, normal, high, urgent"
    category: string optional "general"
    user_email: string optional "User's email for follow-up"
  }

  handler: http {
    method: "POST"
    url: env("TICKET_API_URL") + "/tickets"
    headers: {
      "Authorization": "Bearer " + env("TICKET_API_KEY")
    }
    body: params
  }
}

tool "schedule_meeting" {
  description: "Schedule a meeting with a human agent"

  parameters: {
    topic: string required "Meeting topic"
    preferred_times: array required "List of preferred time slots"
    duration: number optional 30 "Duration in minutes"
    attendee_email: string required "User's email"
  }

  handler: http {
    method: "POST"
    url: env("CALENDAR_API_URL") + "/meetings"
    body: params
  }
}

tool "send_email" {
  description: "Send an email to the user"

  parameters: {
    to: string required "Recipient email"
    subject: string required "Email subject"
    body: string required "Email body (markdown supported)"
    template: string optional "Template to use"
  }

  handler: builtin("email.send")
}

# ============================================================================
# SCRIPTS FOR EFFICIENT OPERATIONS
# ============================================================================

# Memory management script
script "manage-memory" {
  language: "python"
  runtime: "python3"

  capabilities: [database]

  parameters: {
    session_id: string required
    action: string required "get, add, summarize, clear"
    content: object optional "Content to add"
  }

  code: ```python
    import db
    import json
    from datetime import datetime

    session_id = session_id
    action = action

    if action == "get":
        # Get recent conversation history
        history = db.query(f"""
            SELECT role, content, timestamp
            FROM conversation_memory
            WHERE session_id = '{session_id}'
            ORDER BY timestamp DESC
            LIMIT 10
        """)
        print(json.dumps(list(reversed(history)), default=str))

    elif action == "add":
        # Add new message to memory
        content = json.loads(content) if isinstance(content, str) else content
        db.insert("conversation_memory", {
            "session_id": session_id,
            "role": content.get("role"),
            "content": content.get("content"),
            "timestamp": datetime.now().isoformat(),
            "metadata": json.dumps(content.get("metadata", {}))
        })
        print(f"Added message to session {session_id}")

    elif action == "summarize":
        # Get all messages for summarization
        history = db.query(f"""
            SELECT role, content
            FROM conversation_memory
            WHERE session_id = '{session_id}'
            ORDER BY timestamp ASC
        """)
        # Return for LLM summarization
        print(json.dumps(history, default=str))

    elif action == "clear":
        # Clear session memory
        db.execute(f"DELETE FROM conversation_memory WHERE session_id = '{session_id}'")
        print(f"Cleared memory for session {session_id}")
  ```
}

# Context retrieval script
script "retrieve-context" {
  language: "python"
  runtime: "python3"

  capabilities: [network]

  parameters: {
    query: string required
    namespace: string optional "default"
    top_k: number optional 5
    threshold: number optional 0.7
  }

  code: ```python
    import os
    import json
    import urllib.request

    # Get embeddings
    embed_response = urllib.request.urlopen(urllib.request.Request(
        "https://api.openai.com/v1/embeddings",
        data=json.dumps({
            "input": query,
            "model": "text-embedding-3-small"
        }).encode(),
        headers={
            "Authorization": f"Bearer {os.environ['OPENAI_API_KEY']}",
            "Content-Type": "application/json"
        }
    ))
    embedding = json.loads(embed_response.read())["data"][0]["embedding"]

    # Query vector store
    search_response = urllib.request.urlopen(urllib.request.Request(
        f"{os.environ['PINECONE_HOST']}/query",
        data=json.dumps({
            "vector": embedding,
            "topK": top_k,
            "namespace": namespace,
            "includeMetadata": True
        }).encode(),
        headers={
            "Api-Key": os.environ["PINECONE_API_KEY"],
            "Content-Type": "application/json"
        }
    ))
    results = json.loads(search_response.read())

    # Filter by threshold and format
    relevant = []
    for match in results.get("matches", []):
        if match["score"] >= threshold:
            relevant.append({
                "content": match["metadata"].get("content", ""),
                "source": match["metadata"].get("source", ""),
                "score": round(match["score"], 3)
            })

    print(json.dumps(relevant, indent=2))
  ```
}

# User preference script
script "user-preferences" {
  language: "python"
  runtime: "python3"

  capabilities: [database]

  parameters: {
    user_id: string required
    action: string required "get, set, update"
    preferences: object optional
  }

  code: ```python
    import db
    import json

    if action == "get":
        prefs = db.query_one(f"SELECT * FROM user_preferences WHERE user_id = '{user_id}'")
        if prefs:
            print(json.dumps(prefs, default=str))
        else:
            print(json.dumps({"user_id": user_id, "preferences": {}}))

    elif action == "set":
        prefs = json.loads(preferences) if isinstance(preferences, str) else preferences
        db.upsert("user_preferences", {
            "user_id": user_id,
            "preferences": json.dumps(prefs),
            "updated_at": "NOW()"
        })
        print(f"Preferences saved for {user_id}")

    elif action == "update":
        existing = db.query_one(f"SELECT preferences FROM user_preferences WHERE user_id = '{user_id}'")
        current = json.loads(existing.get("preferences", "{}")) if existing else {}
        new_prefs = json.loads(preferences) if isinstance(preferences, str) else preferences
        merged = {**current, **new_prefs}
        db.upsert("user_preferences", {
            "user_id": user_id,
            "preferences": json.dumps(merged),
            "updated_at": "NOW()"
        })
        print(f"Preferences updated for {user_id}")
  ```
}

# ============================================================================
# SPECIALIZED AGENTS
# ============================================================================

agent "intent-classifier" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.1  # Very deterministic

  instruction: ```
    You classify user messages into intents for routing.

    Intents:
    - question: User is asking for information
    - task: User wants to accomplish something
    - feedback: User is providing feedback or opinion
    - complaint: User is expressing dissatisfaction
    - greeting: User is saying hello/goodbye
    - clarification: User is responding to a clarifying question
    - confirmation: User is confirming or denying
    - smalltalk: Casual conversation
    - escalation: User wants human help
    - unknown: Cannot determine intent

    Also extract:
    - entities: Named entities (people, places, products, dates)
    - sentiment: positive, neutral, negative
    - urgency: low, normal, high
    - requires_tools: List of tools that might be needed

    Output JSON only:
    {
      "intent": "question",
      "confidence": 0.95,
      "entities": {"product": "Widget Pro"},
      "sentiment": "neutral",
      "urgency": "normal",
      "requires_tools": ["search_knowledge"],
      "clarification_needed": false
    }
  ```
}

agent "conversation-summarizer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3

  instruction: ```
    You summarize conversation history to preserve context while reducing tokens.

    Create a summary that captures:
    1. Main topics discussed
    2. User's questions and your answers
    3. Any pending issues or follow-ups
    4. User preferences learned
    5. Current task state (if any)

    Format:
    {
      "summary": "Brief narrative summary",
      "topics": ["topic1", "topic2"],
      "resolved_questions": ["..."],
      "pending_items": ["..."],
      "user_preferences": {"key": "value"},
      "current_task": null or {...}
    }
  ```
}

agent "knowledge-retriever" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.2

  instruction: ```
    You retrieve and synthesize information from the knowledge base.

    Process:
    1. Formulate effective search queries from the user's question
    2. Retrieve relevant documents
    3. Synthesize information into a coherent answer
    4. Cite sources where appropriate
    5. Identify gaps if information is incomplete

    Always:
    - Prefer official documentation over community content
    - Note when information might be outdated
    - Suggest related topics the user might find helpful
  ```

  tools: [
    tool("search_knowledge"),
    mcp("knowledge").get_document,
  ]

  scripts: [
    script("retrieve-context")
  ]
}

agent "task-executor" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.2

  instruction: ```
    You help users accomplish tasks using available tools.

    Process:
    1. Understand what the user wants to achieve
    2. Break down into steps if complex
    3. Execute each step using appropriate tools
    4. Report progress and results
    5. Handle errors gracefully

    Before executing:
    - Confirm understanding with user for important actions
    - Explain what you're about to do
    - Request any missing information

    After executing:
    - Summarize what was done
    - Provide next steps if applicable
    - Ask if there's anything else
  ```

  tools: [
    tool("create_ticket"),
    tool("schedule_meeting"),
    tool("send_email"),
  ]
}

agent "conversational-agent" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.7

  instruction: file("system-prompt")

  # All available tools for general conversation
  tools: [
    tool("search_knowledge"),
    tool("create_ticket"),
  ]
}

# ============================================================================
# CONVERSATION PIPELINE
# ============================================================================

pipeline "handle-message" {
  # Step 1: Get session context
  step "load-context" {
    execute: script("manage-memory") {
      session_id: $session.id
      action: "get"
    }
  }

  # Step 2: Get user preferences
  step "load-preferences" {
    execute: script("user-preferences") {
      user_id: $session.user_id
      action: "get"
    }
  }

  # Step 3: Classify intent
  step "classify" {
    use: agent("intent-classifier")

    input: $input.message

    context: [
      step("load-context").output
    ]
  }

  # Step 4: Route based on intent
  branch step("classify").output.intent {
    "question" => step "answer" {
      use: agent("knowledge-retriever")

      input: $input.message

      context: [
        step("load-context").output,
        step("classify").output.entities
      ]
    }

    "task" => step "execute" {
      use: agent("task-executor")

      input: $input.message

      context: [
        step("load-context").output,
        step("load-preferences").output
      ]
    }

    "escalation" => step "escalate" {
      tools: [tool("create_ticket")]

      input: {
        title: "User escalation request",
        description: $input.message,
        priority: "high",
        context: step("load-context").output
      }
    }

    "greeting" => step "greet" {
      use: agent("conversational-agent")
      input: $input.message
      context: [step("load-preferences").output]
    }

    "smalltalk" => step "chat" {
      use: agent("conversational-agent")
      input: $input.message
    }
  }

  # Step 5: Get response from branch
  step "get-response" {
    input: $branch.output
  }

  # Step 6: Store in memory
  step "save-memory" {
    execute: script("manage-memory") {
      session_id: $session.id
      action: "add"
      content: {
        role: "user",
        content: $input.message,
        metadata: step("classify").output
      }
    }
  }

  step "save-response" {
    execute: script("manage-memory") {
      session_id: $session.id
      action: "add"
      content: {
        role: "assistant",
        content: step("get-response").output
      }
    }
  }

  # Step 7: Check if summarization needed
  step "check-summarize" {
    input: step("load-context").output

    condition: length(step("load-context").output) > 20
  }

  branch step("check-summarize").condition {
    true => step "summarize" {
      use: agent("conversation-summarizer")
      input: step("load-context").output
    }
  }

  output: {
    response: step("get-response").output,
    intent: step("classify").output.intent,
    sentiment: step("classify").output.sentiment
  }
}

# Follow-up handling pipeline
pipeline "handle-followup" {
  step "load-context" {
    execute: script("manage-memory") {
      session_id: $session.id
      action: "get"
    }
  }

  step "understand" {
    use: agent("intent-classifier")

    input: $input.message

    context: [
      step("load-context").output,
      $session.last_response
    ]

    instruction: ```
      This is a follow-up to the previous response. Determine:
      1. Is this a clarification request?
      2. Is it a new but related question?
      3. Is it a confirmation/negation?
      4. Is it feedback on the previous response?
    ```
  }

  step "respond" {
    use: agent("conversational-agent")

    input: $input.message

    context: [
      step("load-context").output,
      $session.last_response,
      step("understand").output
    ]
  }

  output: step("respond").output
}

# ============================================================================
# TRIGGERS
# ============================================================================

# WebSocket handler for real-time chat
trigger "chat-message" {
  event: websocket("/chat") {
    message_type: "text"
  }

  run: pipeline("handle-message") {
    session: websocket.session,
    input: {
      message: websocket.message.content
    }
  }

  on_complete: {
    websocket.send({
      type: "response",
      content: output.response,
      metadata: {
        intent: output.intent,
        sentiment: output.sentiment
      }
    })
  }

  on_error: {
    websocket.send({
      type: "error",
      content: "I'm sorry, I encountered an issue. Please try again.",
      error_id: error.id
    })
  }
}

# HTTP API handler
trigger "api-message" {
  event: http("/api/chat") {
    method: "POST"
  }

  run: pipeline("handle-message") {
    session: {
      id: http.body.session_id,
      user_id: http.body.user_id
    },
    input: {
      message: http.body.message
    }
  }

  on_complete: {
    http.respond(200, {
      success: true,
      response: output.response,
      metadata: output
    })
  }
}

# Slack integration
trigger "slack-message" {
  event: slack.message {
    channel_types: ["im", "mpim"]  # Direct messages only
  }

  run: pipeline("handle-message") {
    session: {
      id: "slack:" + slack.user.id,
      user_id: slack.user.id
    },
    input: {
      message: slack.message.text
    }
  }

  on_complete: {
    slack.respond(output.response)
  }
}

# ============================================================================
# CLI ENTRYPOINTS
# ============================================================================

# Interactive chat mode
intent "chat" {
  params: {
    session_id: string optional "Session ID for continuity"
  }

  mode: "interactive"

  run: {
    loop {
      input: stdin("You: ")
      break_if: input == "exit" || input == "quit"

      result: pipeline("handle-message") {
        session: { id: params.session_id || uuid() },
        input: { message: input }
      }

      print("Atlas: " + result.response)
    }
  }
}

# Single message (non-interactive)
intent "ask" {
  params: {
    message: string required "Your question or request"
    session_id: string optional "Session ID"
  }

  run: pipeline("handle-message") {
    session: { id: params.session_id || uuid() },
    input: { message: params.message }
  }

  output: stdout
}

# Clear session
intent "clear-session" {
  params: {
    session_id: string required "Session to clear"
  }

  execute: script("manage-memory") {
    session_id: params.session_id
    action: "clear"
  }

  output: "Session cleared"
}

# Export conversation
intent "export" {
  params: {
    session_id: string required
    format: string optional "json" "json, markdown, txt"
  }

  execute: script("manage-memory") {
    session_id: params.session_id
    action: "get"
  }

  transform: {
    if params.format == "markdown" {
      output.map(m => "**" + m.role + "**: " + m.content).join("\n\n")
    } else if params.format == "txt" {
      output.map(m => m.role + ": " + m.content).join("\n")
    } else {
      json(output)
    }
  }

  output: file("conversations/{{params.session_id}}.{{params.format}}")
}
