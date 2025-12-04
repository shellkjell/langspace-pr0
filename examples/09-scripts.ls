# LangSpace Scripts
# Code-first agent actions for efficient multi-step operations
#
# The Problem: MCP/tool-heavy approaches consume excessive context window space.
# Each tool call loads full data into context, even when agents only need to
# make small modifications.
#
# The Solution: Scripts let agents write and execute code that performs
# multiple operations in a single execution, returning only the results.

# ============================================================================
# BASIC SCRIPT DEFINITION
# ============================================================================

# Simple inline script
script "hello-world" {
  language: "python"
  runtime: "python3"

  code: ```python
    print("Hello from LangSpace script!")
  ```
}

# Script with parameters
script "greet-user" {
  language: "python"
  runtime: "python3"

  parameters: {
    name: string required "The user's name"
    formal: bool optional false "Use formal greeting"
  }

  code: ```python
    if formal:
        print(f"Good day, {name}.")
    else:
        print(f"Hey {name}!")
  ```
}

# ============================================================================
# DATABASE OPERATIONS (The Primary Use Case)
# ============================================================================

# The classic example: find, modify, save a record
# Without scripts: 3+ tool calls, full record in context each time
# With scripts: 1 execution, only result returned
script "update-record" {
  language: "python"
  runtime: "python3"

  capabilities: [database]

  parameters: {
    table: string required "Database table name"
    id: string required "Record ID to update"
    field: string required "Field to modify"
    value: string required "New value for the field"
  }

  code: ```python
    import db

    # All operations happen outside the LLM context window
    record = db.find(table, {"id": id})
    if record is None:
        print(f"Error: Record {id} not found in {table}")
        exit(1)

    old_value = record.get(field, "<unset>")
    record[field] = value
    db.save(table, record)

    # Only this result goes back to the agent
    print(f"Updated {table}/{id}: {field} changed from '{old_value}' to '{value}'")
  ```

  timeout: "30s"
}

# Batch operations - even more efficient
script "batch-update" {
  language: "python"
  runtime: "python3"

  capabilities: [database]

  parameters: {
    table: string required
    updates: array required "Array of {id, field, value} objects"
  }

  code: ```python
    import db
    import json

    results = []
    for update in json.loads(updates):
        record = db.find(table, {"id": update["id"]})
        if record:
            record[update["field"]] = update["value"]
            db.save(table, record)
            results.append(f"✓ {update['id']}")
        else:
            results.append(f"✗ {update['id']} not found")

    print(f"Batch complete: {len(results)} operations")
    for r in results:
        print(f"  {r}")
  ```

  timeout: "2m"
}

# Complex query and aggregation
script "analyze-data" {
  language: "python"
  runtime: "python3"

  capabilities: [database.read]  # Read-only access

  parameters: {
    table: string required
    group_by: string required
    aggregate: string optional "count" "count, sum, avg"
  }

  code: ```python
    import db
    from collections import defaultdict

    records = db.find_all(table)
    groups = defaultdict(list)

    for record in records:
        key = record.get(group_by, "unknown")
        groups[key].append(record)

    print(f"Analysis of {table} grouped by {group_by}:")
    for key, items in sorted(groups.items()):
        if aggregate == "count":
            print(f"  {key}: {len(items)}")
        elif aggregate == "sum":
            total = sum(r.get("value", 0) for r in items)
            print(f"  {key}: {total}")
        elif aggregate == "avg":
            values = [r.get("value", 0) for r in items]
            avg = sum(values) / len(values) if values else 0
            print(f"  {key}: {avg:.2f}")
  ```
}

# ============================================================================
# FILE OPERATIONS
# ============================================================================

# Batch file processing
script "process-files" {
  language: "python"
  runtime: "python3"

  capabilities: [filesystem]

  parameters: {
    pattern: string required "Glob pattern for files"
    operation: string required "Operation: count_lines, find_todos, stats"
  }

  code: ```python
    import glob
    import os

    files = glob.glob(pattern, recursive=True)

    if operation == "count_lines":
        total = 0
        for f in files:
            with open(f) as fp:
                lines = len(fp.readlines())
                total += lines
                print(f"  {f}: {lines} lines")
        print(f"Total: {total} lines across {len(files)} files")

    elif operation == "find_todos":
        for f in files:
            with open(f) as fp:
                for i, line in enumerate(fp, 1):
                    if "TODO" in line or "FIXME" in line:
                        print(f"  {f}:{i}: {line.strip()}")

    elif operation == "stats":
        total_size = sum(os.path.getsize(f) for f in files)
        print(f"Files: {len(files)}")
        print(f"Total size: {total_size / 1024:.1f} KB")
  ```
}

# Code transformation
script "add-headers" {
  language: "python"
  runtime: "python3"

  capabilities: [filesystem]

  parameters: {
    pattern: string required
    header: string required "Header text to add"
  }

  code: ```python
    import glob

    files = glob.glob(pattern, recursive=True)
    modified = 0

    for f in files:
        with open(f, 'r') as fp:
            content = fp.read()

        if not content.startswith(header):
            with open(f, 'w') as fp:
                fp.write(header + "\n" + content)
            modified += 1
            print(f"  ✓ {f}")

    print(f"Added headers to {modified}/{len(files)} files")
  ```
}

# ============================================================================
# API/HTTP OPERATIONS
# ============================================================================

script "fetch-and-process" {
  language: "python"
  runtime: "python3"

  capabilities: [network, filesystem.write]

  parameters: {
    url: string required
    output: string required "Output file path"
    transform: string optional "none" "none, json_pretty, extract_links"
  }

  code: ```python
    import urllib.request
    import json
    import re

    response = urllib.request.urlopen(url)
    data = response.read().decode('utf-8')

    if transform == "json_pretty":
        data = json.dumps(json.loads(data), indent=2)
    elif transform == "extract_links":
        links = re.findall(r'href=["\']([^"\']+)["\']', data)
        data = "\n".join(links)

    with open(output, 'w') as f:
        f.write(data)

    print(f"Fetched {url} -> {output} ({len(data)} bytes)")
  ```

  timeout: "60s"
}

# ============================================================================
# SECURITY: SANDBOXED EXECUTION
# ============================================================================

# Restricted script with explicit permissions
script "safe-read-only" {
  language: "python"
  runtime: "python3"

  # Granular capability control
  capabilities: [
    database.read,      # Can read from database
    filesystem.read,    # Can read files
    # Note: write capabilities NOT granted
  ]

  # Resource limits
  limits: {
    timeout: "30s"
    memory: "128MB"
    cpu: "0.5 cores"
  }

  # Module allowlist
  sandbox: {
    network: false
    allowed_modules: ["json", "datetime", "re", "collections"]
  }

  parameters: {
    query: string required
  }

  code: ```python
    import db
    import json

    # This script can only read, never write
    results = db.query(query)
    print(json.dumps(results, indent=2))
  ```
}

# ============================================================================
# AGENT INTEGRATION
# ============================================================================

# Agent that uses scripts instead of tools
agent "efficient-data-manager" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.2

  instruction: ```
    You are a data management agent optimized for efficiency.

    IMPORTANT: Instead of making multiple tool calls to interact with
    databases or files, you should write and execute scripts that
    perform all necessary operations in a single execution.

    This approach:
    - Keeps context window usage minimal
    - Reduces latency from multiple round-trips
    - Allows complex multi-step operations atomically

    Available scripts:
    - update-record: Find and update a single record
    - batch-update: Update multiple records at once
    - analyze-data: Run aggregation queries
    - process-files: Batch file operations

    When you need to perform data operations, invoke the appropriate
    script with the required parameters. The script executes outside
    your context, and you receive only the result summary.
  ```

  # Scripts this agent can execute
  scripts: [
    script("update-record"),
    script("batch-update"),
    script("analyze-data"),
    script("process-files")
  ]
}

# Agent that can generate custom scripts
agent "script-writer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3

  instruction: ```
    You write efficient Python scripts to accomplish data tasks.

    When given a task, generate a Python script that:
    1. Performs all operations in a single execution
    2. Returns only a concise summary of results
    3. Handles errors gracefully
    4. Uses only the allowed capabilities

    Your generated script will be validated and executed in a sandbox.
  ```

  # This agent can create and execute dynamic scripts
  scripts: [script("dynamic-executor")]
}

# Template for agent-generated scripts
script "dynamic-executor" {
  language: "python"
  runtime: "python3"

  # Code is provided dynamically by the agent
  code: $agent_generated_code

  # Strict sandbox for agent-generated code
  sandbox: {
    network: false
    filesystem: "readonly"
    allowed_modules: ["json", "datetime", "re", "collections", "db"]
  }

  limits: {
    timeout: "60s"
    memory: "256MB"
  }

  capabilities: [database]
}

# ============================================================================
# PIPELINE INTEGRATION
# ============================================================================

# Pipeline that uses scripts for efficient data processing
pipeline "etl-pipeline" {
  # Step 1: Extract data using a script (not individual tool calls)
  step "extract" {
    execute: script("analyze-data") {
      table: "raw_events"
      group_by: "event_type"
      aggregate: "count"
    }
  }

  # Step 2: Agent decides on transformation based on extraction results
  step "plan" {
    use: agent("script-writer")
    input: step("extract").output

    instruction: ```
      Based on the data analysis, write a transformation script
      that normalizes the events and prepares them for reporting.
    ```
  }

  # Step 3: Execute the generated transformation script
  step "transform" {
    execute: script("dynamic-executor") {
      code: step("plan").output
    }
  }

  output: step("transform").output
}

# ============================================================================
# MULTI-LANGUAGE SUPPORT
# ============================================================================

# JavaScript/Node.js script
script "node-processor" {
  language: "javascript"
  runtime: "node"

  capabilities: [filesystem]

  code: ```javascript
    const fs = require('fs');
    const path = require('path');

    const files = fs.readdirSync('.');
    console.log(`Found ${files.length} files`);
    files.forEach(f => console.log(`  - ${f}`));
  ```
}

# Shell script for system operations
script "system-info" {
  language: "bash"
  runtime: "bash"

  capabilities: [system.read]

  code: ```bash
    echo "System Information:"
    echo "  Hostname: $(hostname)"
    echo "  OS: $(uname -s)"
    echo "  Uptime: $(uptime -p 2>/dev/null || uptime)"
    echo "  Disk: $(df -h / | tail -1 | awk '{print $4 " available"}')"
  ```

  timeout: "10s"
}

# SQL script for direct database queries
script "complex-query" {
  language: "sql"
  runtime: "postgresql"

  capabilities: [database.read]

  parameters: {
    start_date: string required
    end_date: string required
  }

  code: ```sql
    WITH daily_stats AS (
      SELECT
        DATE(created_at) as day,
        COUNT(*) as events,
        COUNT(DISTINCT user_id) as users
      FROM events
      WHERE created_at BETWEEN :start_date AND :end_date
      GROUP BY DATE(created_at)
    )
    SELECT
      day,
      events,
      users,
      events::float / users as events_per_user
    FROM daily_stats
    ORDER BY day;
  ```
}
