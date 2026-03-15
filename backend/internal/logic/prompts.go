package logic

import _ "embed"

//go:embed prompts/analyze_repository.txt
var analyzeRepositoryPrompt string

//go:embed prompts/analyze_commit.txt
var analyzeCommitPrompt string
