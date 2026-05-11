APP_NAME       := OZX Room Editor
APP_BUNDLE     := build/$(APP_NAME).app
EXEC_NAME      := OZXRoomEditor
BACKEND_DIR    := ../tile-backend
BACKEND_BINARY := $(BACKEND_DIR)/bin/ozx-roomeditor
SWIFT_BUILD    := .build/release/$(EXEC_NAME)
ICON_OUT       := Resources/AppIcon.icns
ICON_STAGING   := .build/AppIcon.iconset
ICON_MASTER    := .build/icon-master-1024.png

.PHONY: help build run clean backend swift bundle icon

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
	@if [ -f "$(ICON_OUT)" ]; then cp "$(ICON_OUT)" "$(APP_BUNDLE)/Contents/Resources/AppIcon.icns"; \
	else echo "warning: $(ICON_OUT) missing — run 'make icon' to regenerate"; fi
	@cp Resources/Info.plist "$(APP_BUNDLE)/Contents/Info.plist"
	@chmod +x "$(APP_BUNDLE)/Contents/MacOS/$(EXEC_NAME)"
	@chmod +x "$(APP_BUNDLE)/Contents/Resources/ozx-roomeditor"
	@echo "Built $(APP_BUNDLE) ($$(du -sh '$(APP_BUNDLE)' | cut -f1))"

# Regenerate the app icon from tools/MakeIcon.swift. Commit the resulting
# .icns alongside the source — the icon is intentionally checked in so
# `make bundle` works without re-running this every build.
icon: ## Regenerate Resources/AppIcon.icns from tools/MakeIcon.swift
	@echo "Generating $(ICON_OUT)..."
	@mkdir -p .build "$(ICON_STAGING)" Resources
	@swift tools/MakeIcon.swift "$(ICON_MASTER)"
	@for s in 16 32 64 128 256 512; do \
	  s2=$$((s * 2)); \
	  sips -z $$s  $$s  "$(ICON_MASTER)" --out "$(ICON_STAGING)/icon_$${s}x$${s}.png"      >/dev/null; \
	  sips -z $$s2 $$s2 "$(ICON_MASTER)" --out "$(ICON_STAGING)/icon_$${s}x$${s}@2x.png"   >/dev/null; \
	done
	@cp "$(ICON_MASTER)" "$(ICON_STAGING)/icon_512x512@2x.png"
	@iconutil --convert icns "$(ICON_STAGING)" --output "$(ICON_OUT)"
	@rm -rf "$(ICON_STAGING)" "$(ICON_MASTER)"
	@echo "Built $(ICON_OUT) ($$(du -h '$(ICON_OUT)' | cut -f1))"

run: build ## Build and open the .app
	@open "$(APP_BUNDLE)"

clean: ## Remove build artifacts
	@swift package clean
	@rm -rf build/ .build/
	@echo "swift-app: cleaned"
