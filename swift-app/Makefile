APP_NAME       := OZX Room Editor
APP_BUNDLE     := build/$(APP_NAME).app
EXEC_NAME      := OZXRoomEditor
BACKEND_DIR    := ../tile-backend
BACKEND_BINARY := $(BACKEND_DIR)/bin/ozx-roomeditor
SWIFT_BUILD    := .build/release/$(EXEC_NAME)

.PHONY: help build run clean backend swift bundle

help: ## Show targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: bundle ## Build the .app bundle (default)

backend: ## Build the embedded Go server (delegates to tile-backend)
	@$(MAKE) -C $(BACKEND_DIR) build-local

swift: ## Build the Swift release binary
	@swift build -c release

bundle: backend swift ## Assemble the .app bundle
	@echo "Assembling $(APP_BUNDLE)..."
	@rm -rf "$(APP_BUNDLE)"
	@mkdir -p "$(APP_BUNDLE)/Contents/MacOS"
	@mkdir -p "$(APP_BUNDLE)/Contents/Resources"
	@cp "$(SWIFT_BUILD)" "$(APP_BUNDLE)/Contents/MacOS/$(EXEC_NAME)"
	@cp "$(BACKEND_BINARY)" "$(APP_BUNDLE)/Contents/Resources/ozx-roomeditor"
	@cp Resources/Info.plist "$(APP_BUNDLE)/Contents/Info.plist"
	@chmod +x "$(APP_BUNDLE)/Contents/MacOS/$(EXEC_NAME)"
	@chmod +x "$(APP_BUNDLE)/Contents/Resources/ozx-roomeditor"
	@echo "Built $(APP_BUNDLE) ($$(du -sh '$(APP_BUNDLE)' | cut -f1))"

run: build ## Build and open the .app
	@open "$(APP_BUNDLE)"

clean: ## Remove build artifacts
	@swift package clean
	@rm -rf build/ .build/
	@echo "swift-app: cleaned"
