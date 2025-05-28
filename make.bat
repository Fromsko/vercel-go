# Variables
REMOTE_REPO ?=
COMMIT_MSG ?= "Update project"

# Default target
all: help

# Help target
push: check-remote
	@echo 启动自动化发布流程...
	@echo 当前工作分支: %CURRENT_BRANCH%
	\
	echo 正在提交未保存的变更...
	git add . || echo 添加文件失败 && exit /b 1
	@git diff-index --quiet HEAD -- || (
		git commit -m "🔖 [自动提交] 版本发布前预处理" || echo 提交失败 && exit /b 1
		echo 变更已提交（提交消息：🔖 [自动提交] 版本发布前预处理）
	) || (
		echo 工作区干净，无待提交变更
	)
	\
	echo 生成新版本标签...
	$(MAKE) bump-version || echo 版本标签生成失败 && exit /b 1
	\
	echo 同步代码至GitHub...
	git push origin %CURRENT_BRANCH% --follow-tags || echo 代码/标签推送失败 && exit /b 1
	\
	echo.
	echo 发布流程完成！以下步骤将自动进行：
	echo   1. GitHub Actions 将触发构建流程（约1-2分钟）
	echo   2. GoReleaser 将生成多平台二进制文件
	echo   3. 新版本文档将自动发布到 GitHub Releases
	echo.
	echo 实时进度查看: https://github.com/package-register/go-genius/actions
	echo 发布结果查看: https://github.com/package-register/go-genius/releases

check-remote:
	@echo 检查远程仓库配置...
	@git remote | findstr origin >nul
	@if errorlevel 1 (
		echo ❌ 错误：未配置远程仓库
		echo 请先执行以下命令配置仓库地址：
		echo    make add-remote <仓库URL>
		echo 或通过交互模式配置：make add-remote
		exit /b 1
	) else (
		for /f "tokens=*" %%i in ('git remote get-url origin') do set REMOTE_URL=%%i
		echo ✓ 已配置远程仓库: %REMOTE_URL%
	)

help:
	@echo Makefile Usage:
	@echo   make add-remote         - 配置/更新Git远程仓库
	@echo   make commit             - 提交变更并选择提交信息
	@echo   make push               - 自动提交、创建新版本并推送到远程仓库
	@echo   make bump-version       - 创建新的语义化版本标签
	@echo   make test               - 运行所有测试
	@echo   make clean              - 清理生成文件

# Add/update remote repository (Windows compatible)
add-remote:
	@if not "%~1"=="" (
		set RAW_ARGS=%~1
	) else (
		@set /p RAW_ARGS=Enter repository URL: 
	)
	@if not "%RAW_ARGS%"=="" (
		@git remote | findstr origin >nul
		@if errorlevel 1 (
			git remote add origin "%RAW_ARGS%" >nul
			echo ✓ Remote origin added: %RAW_ARGS%
		) else (
			git remote set-url origin "%RAW_ARGS%" >nul
			echo ✓ Remote origin updated to: %RAW_ARGS%
		)
		exit /b 0
	)
	@if not "%RAW_ARGS%"=="" (
		echo ⚠️ Invalid repository URL: '%RAW_ARGS%'
		echo Valid formats: git@... or https://...
		exit /b 1
	)
	@git remote | findstr origin >nul
	@if errorlevel 1 (
		@set /p REMOTE_REPO=Enter repository URL: 
		git remote add origin "%REMOTE_REPO%" >nul
		echo ✓ Remote origin added
	) else (
		@echo 当前远程仓库: %REMOTE_URL%
		@set /p confirm=更新? [y/N]: 
		if /i not "%confirm%"=="y" (
			echo ℹ️ Keeping existing URL
			exit /b 0
		)
		@set /p REMOTE_REPO=Enter new URL: 
		git remote set-url origin "%REMOTE_REPO%" >nul
		echo ✓ Remote URL updated
	)

# Commit changes with a message (include emoji) - Windows compatible
commit:
	@git status --porcelain | findstr . >nul
	@if errorlevel 1 (
		echo No changes to commit. Exiting.
		exit /b 0
	)
	@echo Select a commit message:
	@echo 1. 🚀 Initial commit
	@echo 2. ✨ Add new feature
	@echo 3. 🐛 Fix bug
	@echo 4. 📝 Update documentation
	@echo 5. 🔧 Refactor code
	@echo 6. 👍 Other
	@set /p choice=Enter your choice (1-6): 
	@if "%choice%"=="1" set COMMIT_MSG=🚀 Initial commit
	@if "%choice%"=="2" set COMMIT_MSG=✨ Add new feature
	@if "%choice%"=="3" set COMMIT_MSG=🐛 Fix bug
	@if "%choice%"=="4" set COMMIT_MSG=📝 Update documentation
	@if "%choice%"=="5" set COMMIT_MSG=🔧 Refactor code
	@if "%choice%"=="6" (
		@set /p COMMIT_MSG=Enter custom commit message: 
	) else (
		echo ❌ Invalid choice. Exiting.
		exit /b 1
	)
	git add .
	@git commit -m "%COMMIT_MSG%" || echo Commit failed (no changes to commit).

# Bump version number
bump-version:
	@git describe --tags --abbrev=0 2>nul | findstr . >nul
	@if errorlevel 1 (
		set NEW_VERSION=v0.1.0
	) else (
		for /f "tokens=1-3 delims=." %%a in ('git describe --tags --abbrev=0') do (
			set MAJOR=%%a
			set MINOR=%%b
			set PATCH=%%c
		)
		set /a PATCH+=1
		set NEW_VERSION=v%MAJOR%.%MINOR%.%PATCH%
	)
	git tag -a %NEW_VERSION% -m "Release %NEW_VERSION%"
	echo New version tag %NEW_VERSION% created

build:
	@echo Building binaries...
	goreleaser build --snapshot --clean

# Run all tests
test:
	@go test ./...
	@echo All tests completed.

# Clean generated files
clean:
	@go clean -testcache
	@for /r %%f in (*.out *.test VERSION) do @del "%%f" >nul 2>&1
	@echo Cleaned up generated files.