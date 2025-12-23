{
  "mcpServers": {
    "codacy": {
      "command": "npx",
      "args": ["-y", "@codacy/codacy-mcp@latest"],
      "env": {
        "CODACY_ACCOUNT_TOKEN": "{{CODACY_TOKEN}}"
      }
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "{{GITHUB_TOKEN}}"
      }
    },
    "taskwarrior": {
      "command": "npx",
      "args": ["-y", "mcp-server-taskwarrior"],
      "env": {}
    }
  }
}
