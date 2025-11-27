# Database Connection String
DB_DSN="host=localhost user=vishwakarma_user password=password dbname=vishwakarma_db port=5432 sslmode=disable TimeZone=Asia/Kolkata"

# Default target
.PHONY: test test-verbose test-cover test-html watch

# 📝 Run tests showing specific test names and status (Best for "Processing..." view)
test:
	@echo "🚀 Running Tests..."
	@DATABASE_DSN=$(DB_DSN) gotestsum --format testname ./controllers/...

# 📝 Run tests with standard verbose output (useful for debugging)
test-verbose:
	@echo "🔍 Running Verbose Tests..."
	@DATABASE_DSN=$(DB_DSN) gotestsum --format standard-verbose ./controllers/...

# ⏱️ Watch mode: Re-runs tests instantly when you save a file (Dynamic)
watch:
	@echo "👀 Watching for changes..."
	@DATABASE_DSN=$(DB_DSN) gotestsum --watch --format testname --hide-summary=skipped ./controllers/...

# 📊 Run tests and show coverage statistics in terminal
test-cover:
	@echo "🧪 Measuring Coverage..."
	@DATABASE_DSN=$(DB_DSN) gotestsum --format dots -- -coverprofile=coverage.out ./controllers/...
	@echo "\n📊 Coverage Summary:"
	@go tool cover -func=coverage.out

# 🌐 Generate and open visual HTML coverage report
test-html: test-cover
	@echo "Creating HTML Report..."
	@go tool cover -html=coverage.out -o coverage.html