package digest

var DefaultExcludePatterns = []string{
	// Version control
	".git/",
	".svn/",
	".hg/",
	".cvs/",
	".DS_Store",

	// Dependencies and Modules
	"node_modules/",
	"bower_components/",
	"vendor/",
	"target/",
	"build/",
	"dist/",
	"bin/",
	"obj/",
	"pkg/",

	// Python
	"__pycache__/",
	"*.pyc",
	"*.pyo",
	"*.pyd",
	".pytest_cache/",
	".tox/",
	".mypy_cache/",
	".ruff_cache/",
	"*.egg-info/",
	"venv/",
	".venv/",
	"env/",
	"ENV/",
	"pip-wheel-metadata/",

	// JavaScript / Node
	"package-lock.json",
	"yarn.lock",
	".npm/",
	".yarn/",
	"*.log",
	"coverage/",
	".env",
	".next/",
	"*.lock",
	"*.lockb",

	// IDEs and editors
	".idea/",
	".vscode/",
	".vs/",
	"*.sublime-project",
	"*.sublime-workspace",
	"*.suo",
	"*.user",
	"*.userosscache",
	"*.sln.docstates",

	// Other
	"Thumbs.db",          // Windows
	"desktop.ini",        // Windows
	"terraform.tfstate*", // Terraform states can be sensitive
	".terraform/",        // Terraform directory
	"*.tfvars",           // Terraform variables (can be sensitive) - consider
	"crash.dump",

	// Gitingest / RepoLlama specific
	"digest.txt",            // If the default output is called this
	"pathdigest_digest.txt", // Tooling specific
}
