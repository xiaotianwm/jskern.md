# 更新日志

## 2026-07-09 - v0.1.12 Titlebar And Workspace Drag Fixes

### 修复

- 修复无边框标题栏最小化、最大化、关闭按钮偶发点击无效的问题。
- 修复顶层工作区拖拽排序无效的问题：桌面行为拦截器现在只放行带有 JS Kern.md 工作区拖拽标记的元素，继续拦截图片、链接等默认网页拖拽。
- 修复工作区根节点点击或右键时直接显示拖拽手型的问题；现在只有真实 `dragstart` 后才显示 `grabbing`。

### 变更

- 窗口控制按钮区域显式覆盖 Wails 的 `--wails-draggable` 变量为 `no-drag`，并阻断鼠标按下事件继续冒泡到标题栏拖拽逻辑。
- 顶层工作区拖拽状态改为 React 临时 UI 状态，只用于当前拖拽中的样式显示，不参与持久化。
- README 当前版本说明已更新为 `v0.1.12`。
- 将产品版本提升到 `0.1.12`。

### 发布打包

- Windows installer artifact name: `JSKernMD-Setup-0.1.12-x64.exe`。
- Checksum artifact: `SHA256SUMS.txt`。
- Installer SHA256: `fe0b2ccce4031a2181d73914c0a30b5c43bb62bb383352209ed13cb674c3aa32`。
- Published release target: `v0.1.12`。
- GitHub Release URL: `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.12`。

### 验证

- `go test ./...` passed。
- `wails generate module` passed。
- `npm.cmd run build` passed from `frontend/`。
- `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities。
- `wails build` passed and produced `build/bin/jskernmd.exe`。
- `scripts/package-windows.ps1` passed with process-local `-ExecutionPolicy Bypass` and produced `dist/releases/v0.1.12/JSKernMD-Setup-0.1.12-x64.exe`。
- `SHA256SUMS.txt` was generated with SHA256 `fe0b2ccce4031a2181d73914c0a30b5c43bb62bb383352209ed13cb674c3aa32`。
- `git diff --check` passed after reverting unrelated Wails-generated whitespace-only file rewrites。
- Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped。
- GitHub Release `v0.1.12` was prepared for `JSKernMD-Setup-0.1.12-x64.exe` and `SHA256SUMS.txt`。

---

## 2026-07-09 - v0.1.11 Multi-Workspace And Explorer Integration

### 新增

- 新增多顶层工作区模型：`settings.json` 升级到 `storage_version: 3`，持久化 `workspaces[]`、`active_workspace_id` 和顶层排序。
- 新增 `RestoreWorkspaces()`、`RefreshWorkspaces()`、`AddWorkspace(path)`、`RemoveWorkspace(workspaceID)`、`ReorderWorkspaces(workspaceIDs)` 和 `ConsumeLaunchRequest()` Wails API。
- 新增旧配置迁移：已有 `last_workspace` 会自动迁移成一个工作区条目，保留原有启动恢复体验。
- 新增顶层工作区拖拽排序，排序结果由 Go 写入 AppData。
- 新增顶层工作区右键“移除工作区”，只从 JS Kern.md 中移除，不删除磁盘文件。
- 新增 Windows Explorer 右键入口：
  - Markdown 文件：`Open with JS Kern.md`
  - 文件夹：`Add to JS Kern.md workspace`
- 新增 Wails 单实例转发与启动参数处理，支持 `--open-file`、`--add-workspace` 和直接传入文件/文件夹路径。
- 新增 Go 测试覆盖旧 `last_workspace` 迁移、多工作区排序/移除、Explorer/CLI 文件打开参数。

### 变更

- 打开文件夹现在是加入工作区集合，不再替换已有目录树。
- 左侧目录树改为多根工作区渲染，恢复启动时不主动展开所有根目录和子目录。
- 工作区搜索现在覆盖全部工作区；多工作区时搜索结果路径会带上工作区名称前缀。
- 阅读会话恢复现在可以从多个已恢复工作区中收集仍有效的打开标签页，并保持活动文档。
- `SaveOpenTabs()` 按文档所属工作区分组保存会话，避免多工作区标签页被错误归到单一根目录。
- `OpenDocument()` 会根据文档所在工作区更新 Go 侧活动工作区，保持相对链接、资源和阅读记忆归属正确。
- Windows NSIS 模板注册当前用户级 Explorer 菜单，并在卸载时只删除 `JSKernMD.Open` / `JSKernMD.AddWorkspace` 产品键。
- README、产品范围、架构、约束、决策记录和项目状态已同步为多工作区与 Explorer 入口设计。
- 将产品版本提升到 `0.1.11`。

### 发布打包

- Windows installer artifact name: `JSKernMD-Setup-0.1.11-x64.exe`。
- Checksum artifact: `SHA256SUMS.txt`。
- Installer SHA256: `83da01550453aa3721b0c2a313ed54b1ac7861348080f8ed2c203ad43c01f56f`。
- Published release target: `v0.1.11`。
- GitHub Release URL: `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.11`。

### 验证

- `go test ./...` passed。
- `wails generate module` passed。
- `npm.cmd run build` passed from `frontend/`。
- `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities。
- `wails build` passed and produced `build/bin/jskernmd.exe`。
- `scripts/package-windows.ps1` passed with process-local `-ExecutionPolicy Bypass` and produced `dist/releases/v0.1.11/JSKernMD-Setup-0.1.11-x64.exe`。
- `SHA256SUMS.txt` was generated with SHA256 `83da01550453aa3721b0c2a313ed54b1ac7861348080f8ed2c203ad43c01f56f`。
- `git diff --check` passed。
- Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped。
- GitHub Release `v0.1.11` was created at `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.11`。
- GitHub Release asset verification passed: installer label/name are `JSKernMD-Setup-0.1.11-x64.exe`, checksum label/name are `SHA256SUMS.txt`, and the installer digest is `sha256:83da01550453aa3721b0c2a313ed54b1ac7861348080f8ed2c203ad43c01f56f`。

---

## 2026-07-09 - v0.1.10 Manifest-Owned Go Version Fix

### 修复

- 修复 `v0.1.9` 安装后仍提示更新的问题。
- 根因是 `product.manifest.json`、`wails.json` 和安装包元数据已升到 `0.1.9`，但 Go 后端内部 `appVersion` 常量仍为 `0.1.8`，导致应用检查 GitHub Releases 时认为自己还是旧版。

### 变更

- Go 现在嵌入并解析 `product.manifest.json`，Bootstrap 产品信息、更新检查当前版本、更新下载信息、AppData slug 和更新请求 User-Agent 都从 manifest 派生。
- 移除独立手写的 Go `appVersion` / `appSlug` 常量，避免以后版本号再次分裂。
- 新增测试，确保 Bootstrap 产品版本来自 manifest，并确保当前 manifest 版本不会被 GitHub Release 检查误判为可更新。
- 将产品版本提升到 `0.1.10`。

### 发布打包

- Windows installer artifact name: `JSKernMD-Setup-0.1.10-x64.exe`。
- Checksum artifact: `SHA256SUMS.txt`。
- Installer SHA256: `c50caf64bbad0318e64e1bd06d1b81771467129c53b47d7daa791a76e32a2840`。
- Published release target: `v0.1.10`。
- GitHub Release URL: `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.10`。

### 验证

- GitHub latest before the fix was verified as `v0.1.9`, confirming the repeated prompt was not caused by the Release latest marker.
- `go test ./...` passed。
- `npm.cmd run build` passed from `frontend/`。
- `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities。
- `wails build` passed and produced `build/bin/jskernmd.exe`。
- `scripts/package-windows.ps1` passed with process-local `-ExecutionPolicy Bypass` and produced `dist/releases/v0.1.10/JSKernMD-Setup-0.1.10-x64.exe`。
- `SHA256SUMS.txt` was generated with SHA256 `c50caf64bbad0318e64e1bd06d1b81771467129c53b47d7daa791a76e32a2840`。
- Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped。
- GitHub Release `v0.1.10` was created at `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.10`。
- GitHub Release asset verification passed: installer label/name are `JSKernMD-Setup-0.1.10-x64.exe`, checksum label/name are `SHA256SUMS.txt`, and the installer digest is `sha256:c50caf64bbad0318e64e1bd06d1b81771467129c53b47d7daa791a76e32a2840`。

---
## 2026-07-09 - v0.1.9 Closed-Tab Reading Memory Cleanup

### 修复

- 关闭文档标签页后，Go-owned `SaveOpenTabs()` 会删除该文档在 `data/reading-memory.json` 中的阅读位置记录。
- 从目录树重新打开已关闭文档时，不再恢复旧滚动位置，而是从文档顶部开始。
- `GetReadingPosition()` 现在会忽略旧版 AppData 中残留的、已经不在当前 `open_tabs` 列表里的文档位置记录，避免历史脏数据重新生效。

### 变更

- 将产品版本提升到 `0.1.9`。
- README、产品范围、架构文档、约束文档和决策日志都记录了“关闭标签即清除该文档阅读位置”的行为边界。

### 发布打包

- Windows installer artifact name: `JSKernMD-Setup-0.1.9-x64.exe`。
- Checksum artifact: `SHA256SUMS.txt`。
- Installer SHA256: `b0609f41ed32484f2022aa04a7411817b2f3cb9cd2ce0d95cdb30bb7d9c9ea09`。
- Published release target: `v0.1.9`。
- GitHub Release URL: `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.9`。

### 验证

- `go test ./...` passed。
- `npm.cmd run build` passed from `frontend/`。
- `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities。
- `wails build` passed and produced `build/bin/jskernmd.exe`。
- `scripts/package-windows.ps1` passed with process-local `-ExecutionPolicy Bypass` and produced `dist/releases/v0.1.9/JSKernMD-Setup-0.1.9-x64.exe`。
- `SHA256SUMS.txt` was generated with SHA256 `b0609f41ed32484f2022aa04a7411817b2f3cb9cd2ce0d95cdb30bb7d9c9ea09`。
- Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped。
- GitHub Release `v0.1.9` was created at `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.9`。
- GitHub Release asset verification passed: installer label/name are `JSKernMD-Setup-0.1.9-x64.exe`, checksum label/name are `SHA256SUMS.txt`, and the installer digest is `sha256:b0609f41ed32484f2022aa04a7411817b2f3cb9cd2ce0d95cdb30bb7d9c9ea09`。

---
## 2026-07-09 - README 关于与更新日志中文化

### 变更

- 将根 `README.md` 的产品介绍区调整为中文 `关于 JS Kern.md`，补充本地 Markdown 文档库阅读场景说明。
- 在根 `README.md` 新增中文 `更新日志` 区块，链接 GitHub Releases 和 `docs/CHANGELOG.md`，并补充 `v0.1.8` 用户可见更新摘要。
- 将 `docs/CHANGELOG.md` 的主标题和常用三级结构标题中文化，减少文档入口的英文残留。
- 将 GitHub 仓库 About 描述同步为中文：轻量、快速、目录树优先的桌面 Markdown 阅读器。

### 验证

- `README.md`、`docs/CHANGELOG.md`、`docs/PROJECT_STATE.md` 均通过 UTF-8 读取校验。
- `gh repo view xiaotianwm/jskern.md` 已确认仓库 About 描述为中文。
- `git diff --check` passed。

---
## 2026-07-09 - README Chinese Product Overview

### 变更

- Rewrote the root `README.md` in Simplified Chinese.
- Expanded the README feature overview around the product's lightweight, fast, folder-first Markdown reading workflow.
- Documented the current user-facing capabilities: workspace directory tree, Markdown rendering, outline navigation, multi-tab reading, reading memory, workspace search, current-document find, directory auto-sync, weak external-change reminders, context menus, rename, Shiki highlighting, desktop anti-web behavior, language/theme support, installer/update flow, and development constraints.

### 验证

- `README.md` was rewritten as UTF-8 text.
- `git diff --check` passed.

---

## 2026-07-09 - v0.1.8 Reader Layout And Tree Rename

### 新增

- Added `RenamePath(path, newName)` as a Go-owned Wails API for directory-tree renames.
- Added Go validation for rename targets: source must exist inside the current workspace, the workspace root cannot be renamed, new names must be single path segments, target paths must remain inside the workspace, duplicate targets are rejected, and renamed files must keep a Markdown extension.
- Added Go tests for Markdown file rename, directory rename, refreshed tree state, unsafe target rejection, duplicate target rejection, workspace-root rejection, and outside-workspace source rejection.
- Added an app-owned inline rename editor to directory-tree rows, opened from the tree context menu without using browser prompt UI.
- Added localized Chinese and English `Rename` context-menu text through Go-owned locale JSON.

### 变更

- The outline panel now uses a fixed shell plus an internal scroll region, so long document outlines can scroll independently.
- The center reader column now expands with the available window width up to a wider desktop reading cap instead of staying locked to the earlier narrow 820px content width.
- The custom titlebar now handles double-click maximize/restore and isolates all right-side window buttons with explicit no-drag behavior and stopped double-click propagation.
- Directory-tree rename refreshes the Go-scanned tree immediately and remaps open tabs, selected paths, expanded directories, and the active document path when a renamed directory contains open tabs.
- Advanced the product version to `0.1.8` for the reader layout and tree rename release.

### 发布打包

- Windows installer artifact name: `JSKernMD-Setup-0.1.8-x64.exe`.
- Checksum artifact: `SHA256SUMS.txt`.
- Installer SHA256: `e57cbbfb441cc6c705f3363c1484774ed3ff402883d9ff8ba2518a6c374ace86`.
- Published release target: `v0.1.8`.
- GitHub Release URL: `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.8`.

### 验证

- `go test ./...` passed.
- `wails generate module` passed.
- `npm.cmd run build` passed from `frontend/`.
- `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
- `wails build` passed and produced `build/bin/jskernmd.exe`.
- `scripts/package-windows.ps1` passed with process-local `-ExecutionPolicy Bypass` and produced `dist/releases/v0.1.8/JSKernMD-Setup-0.1.8-x64.exe`.
- `SHA256SUMS.txt` was generated with SHA256 `e57cbbfb441cc6c705f3363c1484774ed3ff402883d9ff8ba2518a6c374ace86`.
- Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
- GitHub Release `v0.1.8` was created at `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.8`.
- GitHub Release asset verification passed: installer label/name are `JSKernMD-Setup-0.1.8-x64.exe`, checksum label/name are `SHA256SUMS.txt`, and the installer digest is `sha256:e57cbbfb441cc6c705f3363c1484774ed3ff402883d9ff8ba2518a6c374ace86`.

---

## 2026-07-08 - v0.1.7 Installer Locale And Upgrade Path

### 新增

- Added English and Simplified Chinese NSIS Modern UI language support to the Windows installer.
- Added installer and uninstaller startup language selection based on the current Windows UI language, without adding a language picker step.
- Added Windows upgrade directory detection that reads `InstallLocation` from the app uninstall registry entry before the directory page is shown.
- Added compatibility fallback for older installers by deriving the previous install directory from the quoted `UninstallString` path when `InstallLocation` is missing.
- Added installer registry writes for `InstallLocation` and `InstallerLanguage`.

### 变更

- Advanced the product version to `0.1.7` for the installer locale and upgrade-path release.
- `scripts/package-windows.ps1` now synchronizes `wails.json.info` from `product.manifest.json` before invoking Wails NSIS packaging, keeping installer metadata aligned with the product manifest.
- `scripts/package-windows.ps1` now writes generated JSON and checksum text as UTF-8 with LF line endings to avoid Windows formatting churn.
- Kept `project.nsi` ASCII-only after validation showed `makensis` reads the script as ACP by default; localized wizard text now relies on NSIS built-in language tables.

### 发布打包

- Windows installer artifact name: `JSKernMD-Setup-0.1.7-x64.exe`.
- Checksum artifact: `SHA256SUMS.txt`.
- Installer SHA256: `7a3d782997a37412ab1b20922a462b0ce825fdd9e050219533a8e5636e9822ff`.
- Published release target: `v0.1.7`.
- GitHub Release URL: `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.7`.

### 验证

- `go test ./...` passed.
- `npm.cmd run build` passed from `frontend/`.
- `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
- `wails build` passed and produced `build/bin/jskernmd.exe`.
- `scripts/package-windows.ps1` passed with process-local `-ExecutionPolicy Bypass` and produced `dist/releases/v0.1.7/JSKernMD-Setup-0.1.7-x64.exe`.
- Initial NSIS validation failed when Chinese custom strings were placed directly in `project.nsi`; the final ASCII-only NSIS template packaged successfully.
- `SHA256SUMS.txt` was generated with SHA256 `7a3d782997a37412ab1b20922a462b0ce825fdd9e050219533a8e5636e9822ff`.
- Installer version resource verification passed: `ProductName=JS Kern.md`, `ProductVersion=0.1.7`, `FileDescription=JS Kern.md Installer`, `CompanyName=JS Labs`.
- GitHub Release `v0.1.7` was created at `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.7`.
- GitHub Release asset verification passed: installer label/name are `JSKernMD-Setup-0.1.7-x64.exe`, checksum label/name are `SHA256SUMS.txt`, and the installer digest is `sha256:7a3d782997a37412ab1b20922a462b0ce825fdd9e050219533a8e5636e9822ff`.
- `git diff --check` passed.
- Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

## 2026-07-08 - v0.1.6 Directory Tree And Tab Context Menus

### 新增

- Added app-owned right-click context menus for directory-tree rows.
- Added directory-tree menu actions for opening Markdown files, expanding or collapsing directories, refreshing the workspace tree, copying the path, and showing the item in the system file manager.
- Added app-owned right-click context menus for open document tabs.
- Added tab menu actions for switching to a tab, closing a tab, closing other tabs, closing tabs to the right, copying the path, and showing the document in the system file manager.
- Added `RevealPath(path)` as a Go Wails API that validates the target is an existing file or directory inside the current workspace before launching the platform file manager.
- Added Wails runtime clipboard usage for copying Go-provided file and directory paths from context menus.
- Added localized context-menu labels for Chinese and English.
- Added Go tests for workspace-boundary validation before file-manager reveal actions.

### 变更

- Advanced the product version to `0.1.6` for the directory-tree and tab context-menu release.
- Reused the existing Go-owned `RefreshWorkspace()` flow for the directory-tree refresh menu action, keeping filesystem truth in Go.
- Reused the existing tab-session persistence flow for tab context-menu close operations so tab order and active tab state stay in AppData reading memory.
- Recorded the release workflow rule that every meaningful product update must synchronize the installer and checksum to GitHub Releases unless release work is explicitly paused.
- Updated project constraints, architecture notes, product scope, and decision log to record that context-menu file actions remain Go-validated.

### 发布打包

- Windows installer artifact name: `JSKernMD-Setup-0.1.6-x64.exe`.
- Checksum artifact: `SHA256SUMS.txt`.
- Installer SHA256: `1cd6de5ba0fd880e098f1b0bd519bb74977eb8fb95ec4498cecb34ba03401cc8`.
- Published release target: `v0.1.6`.
- GitHub Release URL: `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.6`.

### 验证

- `go test ./...` passed.
- `wails generate module` passed.
- `npm.cmd run build` passed from `frontend/`.
- `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
- `wails build` passed and produced `build/bin/jskernmd.exe`.
- `scripts/package-windows.ps1` passed with process-local `-ExecutionPolicy Bypass` and produced `dist/releases/v0.1.6/JSKernMD-Setup-0.1.6-x64.exe`.
- `SHA256SUMS.txt` was generated with SHA256 `1cd6de5ba0fd880e098f1b0bd519bb74977eb8fb95ec4498cecb34ba03401cc8`.
- GitHub Release `v0.1.6` was created at `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.6`.
- GitHub Release asset verification passed: installer label/name are `JSKernMD-Setup-0.1.6-x64.exe`, checksum label/name are `SHA256SUMS.txt`, and the installer digest is `sha256:1cd6de5ba0fd880e098f1b0bd519bb74977eb8fb95ec4498cecb34ba03401cc8`.
- `git diff --check` passed.
- Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

## 2026-07-08 - v0.1.5 Multi-Tab Reading Session

### 新增

- Added Go-owned workspace tab-session memory stored inside AppData `data/reading-memory.json`.
- Added `open_tabs` and `active_document` fields to each workspace reading-memory record while keeping the legacy `last_document` field for compatibility.
- Added `GetReadingSession()` to restore open tabs, active document, and the active document reading position after `RestoreWorkspace()`.
- Added `SaveOpenTabs(paths, activePath)` so React can report tab order and active tab without owning durable storage.
- Added startup restoration for the previous workspace's tab strip and active tab.
- Added a compact center-reader tab bar for open Markdown documents.
- Added tab close behavior, active-tab fallback to the nearest remaining tab, and empty-tab cleanup.
- Added `Ctrl/Cmd+W` for closing the current tab and `Ctrl/Cmd+Tab` / `Ctrl/Cmd+Shift+Tab` for tab cycling.
- Added localized tab accessibility labels for Chinese and English.
- Added Go tests for tab-session persistence, outside-workspace rejection, and legacy last-document normalization.

### 变更

- Advanced the product version to `0.1.5` for the multi-tab reading-session release.
- Switching tabs now force-saves the outgoing document's current reading position before opening the target tab.
- Switching to a previously opened tab now restores that document's own saved reading position instead of jumping to the top.
- Reloading a changed document from the weak disk-change reminder now preserves the current reader offset using the reloaded document metadata.
- Updated project constraints, architecture notes, product scope, and decision log to record that tab sessions are Go-owned AppData state.

### 发布打包

- Windows installer artifact name: `JSKernMD-Setup-0.1.5-x64.exe`.
- Checksum artifact: `SHA256SUMS.txt`.
- Installer SHA256: `57f682aeab4fcd8f0e33f1e289585aea738c4a4a840aab874c39eaa68e028b57`.
- Published release target: `v0.1.5`.
- GitHub Release URL: `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.5`.

### 验证

- `go test ./...` passed.
- `wails generate module` passed.
- `npm.cmd run build` passed from `frontend/`.
- `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
- `wails build` passed and produced `build/bin/jskernmd.exe`.
- `scripts/package-windows.ps1` passed with process-local `-ExecutionPolicy Bypass` and produced `dist/releases/v0.1.5/JSKernMD-Setup-0.1.5-x64.exe`.
- `SHA256SUMS.txt` was generated with SHA256 `57f682aeab4fcd8f0e33f1e289585aea738c4a4a840aab874c39eaa68e028b57`.
- GitHub Release `v0.1.5` was created at `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.5`.
- GitHub Release asset verification passed: installer label/name are `JSKernMD-Setup-0.1.5-x64.exe`, checksum label/name are `SHA256SUMS.txt`, and the installer digest is `sha256:57f682aeab4fcd8f0e33f1e289585aea738c4a4a840aab874c39eaa68e028b57`.
- Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

## 2026-07-08 - v0.1.4 Reading Memory

### 新增

- Added Go-owned reading memory stored at AppData `data/reading-memory.json`.
- Added versioned reading-memory storage with `.bad-*` backup recovery for invalid JSON.
- Added workspace-scoped last-document memory so startup can reopen the last read Markdown file after `RestoreWorkspace()`.
- Added bounded per-workspace document position records capped at 300 documents.
- Added per-document saved scroll offset, scroll ratio, nearest heading ID, document modified time, document size, and update time.
- Added `GetReadingMemory()` to return the current workspace's last readable document position.
- Added `GetReadingPosition(path)` so reopening a document from the tree can restore its own saved position.
- Added `SaveReadingPosition(path, scrollTop, scrollRatio, headingID, modifiedAt, size)` with workspace boundary validation.
- Added frontend debounced scroll-position reporting through Go, with cleanup for timers and scroll listeners.
- Added startup restore that opens the last remembered document after the workspace tree is restored.
- Added exact scroll restoration when saved document metadata still matches the current file.
- Added changed-document fallback restoration that targets the saved heading ID when possible and otherwise opens from the top.
- Added Go tests for reading-memory persistence, restore, outside-workspace rejection, bad-file backup, and pruning.

### 变更

- Advanced the product version to `0.1.4` for the reading memory release.
- Updated project constraints, architecture notes, product scope, and decision log to record that reading memory is Go-owned AppData state, not frontend storage.
- Opening a new document no longer inherits the previous reader scroll; it uses that document's own saved memory when present, otherwise starts at the top.

### 发布打包

- Windows installer artifact name: `JSKernMD-Setup-0.1.4-x64.exe`.
- Checksum artifact: `SHA256SUMS.txt`.
- Installer SHA256: `e2dc5aacbfe3cc9f48032c1d73320211fbaef1439b08adce98b657db5cfe3068`.
- Published release target: `v0.1.4`.
- GitHub Release URL: `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.4`.

### 验证

- `go test ./...` passed.
- `wails generate module` passed.
- `npm.cmd run build` passed from `frontend/`.
- `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
- `wails build` passed and produced `build/bin/jskernmd.exe`.
- `scripts/package-windows.ps1` passed with process-local `-ExecutionPolicy Bypass` and produced `dist/releases/v0.1.4/JSKernMD-Setup-0.1.4-x64.exe`.
- `SHA256SUMS.txt` was generated with SHA256 `e2dc5aacbfe3cc9f48032c1d73320211fbaef1439b08adce98b657db5cfe3068`.
- GitHub Release `v0.1.4` was created at `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.4`.
- GitHub Release asset verification passed: installer label/name are `JSKernMD-Setup-0.1.4-x64.exe`, checksum label/name are `SHA256SUMS.txt`, and the installer digest is `sha256:e2dc5aacbfe3cc9f48032c1d73320211fbaef1439b08adce98b657db5cfe3068`.
- Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

## 2026-07-08

### 新增

- Added `RefreshWorkspace()` as a Go-owned Wails API for refreshing the currently open workspace directory tree without reopening the folder picker.
- Added a Go workspace structure signature that compares directory and Markdown-file paths while ignoring document content changes.
- Added frontend workspace auto-sync polling that calls Go every 3 seconds while a workspace is open.
- Added auto-sync expansion preservation so directories that still exist stay expanded, the workspace root stays expanded, and newly discovered child directories remain collapsed by default.
- Added Go tests proving workspace refresh reports unchanged trees, Markdown file additions, Markdown file deletions, content-only edits, and skipped heavy folders correctly.
- Added Go-owned update checking against the official GitHub Releases feed for `xiaotianwm/jskern.md`.
- Added installer asset filtering so update prompts only accept canonical `JSKernMD-Setup-<version>-x64.exe` release assets.
- Added Go-owned update installer download into AppData `temp/update/`.
- Added SHA256 verification for downloaded update installers when release metadata provides a digest.
- Added `DismissUpdate(version)` persistence through AppData `ignored_update_version`.
- Added `OpenDownloadedUpdate(path)` so the app can open the downloaded installer only after the user explicitly clicks install.
- Added localized toolbar update reminder UI with release notes, download, install, and ignore actions.
- Added Go tests for update release parsing, installer asset filtering, checksum validation, and ignored update persistence.
- Added app-owned current-document find for the active Markdown document.
- Added a `Ctrl/Cmd+F` handoff from the desktop guard layer to the reader find UI, keeping browser default find blocked.
- Added rendered Markdown match highlighting with current-match emphasis and previous/next navigation.
- Added localized current-document find labels for Chinese and English through Go-owned locale dictionaries.
- Added Go-managed persistent `locale` and `theme` settings under AppData `config/settings.json`.
- Added `SwitchLanguage(locale)` and `SwitchTheme(theme)` Wails APIs.
- Added toolbar language and theme controls that call Go APIs and consume Go-owned locale strings.
- Added system/light/dark theme support using existing CSS variables and a `prefers-color-scheme` listener for system mode.
- Added localized shell labels for system, light, and dark theme options.
- Added Go tests for persisted language/theme switching and normalization.
- Added `scripts/package-windows.ps1` to build a Wails NSIS installer, stage it under `dist/releases/v<version>/`, and generate `SHA256SUMS.txt`.

### 变更

- Advanced the product version to `0.1.3` for the directory auto-sync release.
- Workspace search state now clears after an auto-synced tree change so stale search results do not point at removed or renamed files.
- Architecture, constraints, product scope, and decisions now explicitly record that directory refresh is Go-owned and separate from active document content-change reminders.
- Clarified that JS Kern.md is an independent desktop app and must not use the shared Cloudflare/control-plane release system.
- Replaced the previous control-plane update-source follow-up with a GitHub Releases-only product boundary.
- Advanced the product version to `0.1.2` for the update-check and find-focus release.
- `Ctrl/Cmd+F` now focuses and selects the current-document find input after the find bar is mounted, fixing the previous timing race.
- `settings.json` storage advanced to version 2 for the ignored update version field.
- Advanced the product version to `0.1.1` for the current-document find release.
- Switching documents now clears transient current-document find state and removes match highlights.
- Closing the find bar now removes all current-document highlights from the rendered Markdown DOM.
- Architecture notes now record current-document find as transient React-owned UI state and move language/theme switching into the implemented API list.
- Startup bootstrap now calls `GetBootstrap("")` so Go settings choose the current locale and theme instead of hardcoding `zh-CN` in React.
- `settings.json` now preserves locale/theme defaults while keeping existing workspace persistence behavior.
- GitHub Release packaging policy now treats installers as the primary user-facing artifact and reserves raw `jskernmd.exe` for local validation.
- Windows installer naming was corrected to the user-facing `JSKernMD-Setup-<version>-x64.exe` pattern instead of the internal binary-style name.
- GitHub Release asset labels now match their filenames exactly so the download list is readable.

### 发布打包

- Windows installer artifact name: `JSKernMD-Setup-<version>-x64.exe`.
- Checksum artifact: `SHA256SUMS.txt`.
- Published release target: `v0.1.3`.
- Windows installer artifact: `JSKernMD-Setup-0.1.3-x64.exe`.
- Installer SHA256: `d596cc6d02b1ebc43822a9c7bafbbf3b59e7b6dbb82299c624260a0eda3dfeb5`.
- Published release target: `v0.1.2`.
- Windows installer artifact: `JSKernMD-Setup-0.1.2-x64.exe`.
- Installer SHA256: `449f99550137ea0d36457860b18b456ed3224825cd14a7ac273661f9985b4574`.
- Published release target: `v0.1.1`.
- Windows installer artifact: `JSKernMD-Setup-0.1.1-x64.exe`.
- Installer SHA256: `83513e2681d3a753136a60c6d777f3722ea67d4169a6dd022bd85565bae910a7`.

### 验证

- `v0.1.3` directory auto-sync release:
  - Product version sources were updated to `0.1.3`.
  - `go test ./...` passed.
  - `wails generate module` passed.
  - `npm.cmd run build` passed from `frontend/`.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - `scripts/package-windows.ps1` passed with process-local `-ExecutionPolicy Bypass` and produced `dist/releases/v0.1.3/JSKernMD-Setup-0.1.3-x64.exe`.
  - `SHA256SUMS.txt` was generated with SHA256 `d596cc6d02b1ebc43822a9c7bafbbf3b59e7b6dbb82299c624260a0eda3dfeb5`.
  - GitHub Release `v0.1.3` was created at `https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.3`.
  - GitHub Release asset verification passed: installer label/name are `JSKernMD-Setup-0.1.3-x64.exe`, checksum label/name are `SHA256SUMS.txt`, and the installer digest is `sha256:d596cc6d02b1ebc43822a9c7bafbbf3b59e7b6dbb82299c624260a0eda3dfeb5`.
  - Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
- `v0.1.2` update-check and find-focus release:
  - Product version sources were updated to `0.1.2`.
  - `go test ./...` passed.
  - `wails generate module` passed.
  - `npm.cmd run build` passed from `frontend/`.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - `scripts/package-windows.ps1` passed with process-local `-ExecutionPolicy Bypass` and produced `dist/releases/v0.1.2/JSKernMD-Setup-0.1.2-x64.exe`.
  - `SHA256SUMS.txt` was generated with SHA256 `449f99550137ea0d36457860b18b456ed3224825cd14a7ac273661f9985b4574`.
  - Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
- `v0.1.1` installer release:
  - `go test ./...` passed.
  - `npm.cmd run build` passed from `frontend/`.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - Direct `scripts\package-windows.ps1` execution was blocked by PowerShell policy; rerunning with process-local `-ExecutionPolicy Bypass` succeeded.
  - `scripts/package-windows.ps1` passed and produced `dist/releases/v0.1.1/JSKernMD-Setup-0.1.1-x64.exe`.
  - `SHA256SUMS.txt` was generated with SHA256 `83513e2681d3a753136a60c6d777f3722ea67d4169a6dd022bd85565bae910a7`.
  - Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
- Current-document find:
  - `npm.cmd run build` passed from `frontend/`.
  - `go test ./...` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
- Language/theme settings and installer staging:
  - `go test ./...` passed.
  - `wails generate module` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - Initial `scripts/package-windows.ps1` run built `build/bin/jskernmd.exe` but could not create the NSIS installer because `makensis` was missing.
  - Installed `NSIS.NSIS` through `winget` and updated `scripts/package-windows.ps1` to detect common NSIS install paths when PATH is not refreshed.
  - `scripts/package-windows.ps1` passed and produced `dist/releases/v0.1.0/jskernmd-v0.1.0-windows-amd64-setup.exe`.
  - `SHA256SUMS.txt` was generated with SHA256 `3cbbca75ffbbf8561f12599ab575a031c2e79e5530746af42801be8544ddf2c0`.
  - GitHub Release `v0.1.0` now contains the installer and checksum file; the previous raw exe asset was removed.
  - Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
- Installer rename correction:
  - `scripts/package-windows.ps1` passed and produced `dist/releases/v0.1.0/JSKernMD-Setup-0.1.0-x64.exe`.
  - `SHA256SUMS.txt` was regenerated with SHA256 `f591d4b676e4cb5b05184e4c9c71ccbab5c869f7029f43225be51d8a898d0bfb`.
  - GitHub Release `v0.1.0` was updated to use `JSKernMD-Setup-0.1.0-x64.exe`.
- Release asset label correction:
  - `JSKernMD-Setup-0.1.0-x64.exe` label now matches the filename.
  - `SHA256SUMS.txt` label now matches the filename.

---

## 2026-07-07

### 新增

- Added `SearchWorkspace(query)` as a Go-owned Wails API for workspace Markdown search.
- Added bounded on-demand Markdown search across the current workspace:
  - matches Markdown file names and workspace-relative paths
  - matches document body text and returns a compact snippet
  - skips hidden entries and heavy folders such as `node_modules`, `dist`, `build`, and `vendor`
  - keeps search results capped at 50 items
- Added Go tests for file-name hits, content hits, skipped folders, and searching without an open workspace.
- Added a toolbar search input with debounced Wails calls, stale-response protection, keyboard Enter/Escape handling, and click-to-open results.
- Added localized search UI text in `zh-CN` and `en`.
- Regenerated Wails frontend bindings for the new `SearchResult` model and `SearchWorkspace` API.

### 变更

- Moved `SearchWorkspace(query)` from planned architecture work into the implemented Wails API surface.

### 验证

- Workspace search:
  - `go test ./...` passed.
  - `wails generate module` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

### 变更

- Reader position changes are now direct offset assignments instead of browser scroll animations.
- Removed smooth scrolling from the reader container so newly opened documents appear at the top immediately.
- Moved the external document-change reminder out of the Markdown document flow and into a bottom overlay inside the center reader area.
- Split the reader surface into a fixed shell plus an internal scroll container so status reminders stay visible regardless of document scroll position.
- Opening or reloading a document from the workspace tree now resets the center reader scroll position to the top instead of inheriting the previous document's scroll offset.
- Workspace-relative Markdown links with heading fragments still navigate to their requested heading after the new document renders.

### 验证

- Instant reader positioning:
  - `go test ./...` passed.
  - `npm.cmd run build` passed.
  - Initial `npm.cmd audit --audit-level=moderate` hit an npm registry `ECONNRESET`; retrying through `127.0.0.1:10808` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
- Reader status banner and scroll reset:
  - `go test ./...` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

### 新增

- Added Go-provided document freshness metadata:
  - `modifiedAt`
  - `size`
- Added `StatDocument(path, knownModifiedAt, knownSize)` as the Go-owned document status check API.
- Added Go tests for unchanged, changed, deleted, and outside-workspace document status checks.
- Added localized reader-surface error copy for document open failures.
- Added localized weak external-change reminder copy with reload and dismiss actions.

### 变更

- Failed document opens now clear stale reader content and show a visible error panel instead of silently leaving the previous document onscreen.
- The current document now polls Go for disk freshness and shows a non-modal reminder when the file changes externally.
- Reloading from the reminder reuses the existing document open path, so workspace boundary validation and Markdown rendering stay Go-owned.
- Dismissing an external-change reminder suppresses only that exact changed snapshot; a later file change can surface a new reminder.
- Wails frontend bindings were regenerated for the new `DocumentStatus` model and `StatDocument` API.

### 验证

- Document status notices:
  - `go test ./...` passed.
  - `wails generate module` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

### 新增

- Added `markdown-reader-icon.svg` as the product app icon source artwork.
- Added a converted 1024x1024 alpha PNG app icon at `build/appicon.png`.
- Added a regenerated Windows ICO at `build/windows/icon.ico` so Wails embeds the new icon into `jskernmd.exe`.

### 变更

- Replaced the default Wails application icon with the JS Kern.md Markdown reader icon.

### 验证

- App icon integration:
  - Rendered `markdown-reader-icon.svg` to `build/appicon.png` with transparent corners.
  - Regenerated `build/windows/icon.ico` from the new PNG through Wails.
  - `go test ./...` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

### 新增

- Added Shiki-based syntax highlighting for rendered Markdown code blocks.
- Added a focused frontend highlighter module that scans Go-sanitized Markdown HTML after document render.
- Added explicit language alias handling for common Markdown fence labels such as `js`, `ts`, `sh`, `ps1`, and `yml`.
- Added a Go sanitizer allowance for `class` attributes on `pre` and `code` elements so fenced code language markers survive into the renderer.
- Added a Go test proving fenced code blocks preserve `language-*` classes for the Shiki handoff.

### 变更

- Code highlighting now remains a frontend display enhancement while Markdown parsing, HTML rendering, and sanitization stay in Go.
- Shiki now uses a fine-grained bundled language/theme set instead of importing the full Shiki language catalog.
- Unsupported or unlabeled code blocks intentionally fall back to the existing plain code-block rendering.

### 验证

- Shiki code highlighting:
  - `go test ./...` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

### 新增

- Added a Go-controlled `/kern-asset` endpoint through the Wails asset server for local Markdown images.
- Added Markdown AST rewriting for workspace-local bitmap image references.
- Added Markdown AST rewriting for relative Markdown document links.
- Added `OpenWorkspaceDocument(path)` Wails API for opening workspace-relative Markdown links.
- Added frontend Markdown click handling for links marked with `data-kern-document`.
- Added image sizing styles for rendered Markdown images.
- Added Go tests for local image rewriting, relative Markdown link rewriting, workspace-relative document opening, and asset endpoint path rejection.

### 变更

- `OpenDocument(path)` now renders Markdown through the App instance so it can resolve workspace-local resources.
- Workspace-local image serving now streams files through Go instead of embedding image bytes in JSON.
- Workspace-relative document links now round-trip through Go path validation instead of letting the WebView resolve local paths.
- SVG images are intentionally not served in this first local asset pass; bitmap formats are supported first.

### 验证

- Local image and relative Markdown link support:
  - `go test ./...` passed.
  - `wails generate module` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

### 新增

- Added Go-managed AppData initialization for the `jskernmd` data root.
- Added the required local storage layout:
  - `config/`
  - `data/`
  - `logs/`
  - `cache/`
  - `temp/`
  - `runtime/`
  - `crash/`
- Added versioned `config/settings.json` with `storage_version` and `last_workspace`.
- Added atomic settings writes through a temporary file and rename.
- Added `.bad-*` backup behavior for invalid JSON settings files before falling back to defaults.
- Added `RestoreWorkspace()` Wails API.
- Added startup restore in the frontend so the last valid workspace tree reappears automatically.
- Added Go tests for AppData layout creation, settings persistence, workspace restore, and bad settings backup.

### 变更

- `ScanWorkspace(path)` now persists the successfully opened workspace directory.
- Startup workspace restoration keeps the root directory expanded while child directories remain collapsed by default.
- Project constraints, architecture notes, and decision log now record that directory-tree workspace state belongs in Go-managed AppData, not frontend storage.

### 验证

- AppData workspace persistence:
  - `go test ./...` passed.
  - `wails generate module` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
  - AppData smoke check passed: `C:\Users\cool\AppData\Roaming\jskernmd\config\settings.json` was created with `storage_version: 1`.

---

### 新增

- Added expand/collapse behavior to the workspace tree.
- Added root-only default expansion: the workspace root opens, child directories start collapsed.
- Added an internal scroll region for the left workspace tree panel.
- Added `goldmark` Markdown parsing and GFM support in the Go backend.
- Added `bluemonday` sanitization for rendered Markdown HTML.
- Added `OpenDocument(path)` Wails API.
- Added current-workspace path boundary validation before opening a document.
- Added symlink-aware real-path validation so files linked from inside a workspace cannot resolve outside the workspace root.
- Added document model fields for path, filename, title, sanitized HTML, and heading outline.
- Added heading extraction from the goldmark AST, including generated heading IDs for outline navigation.
- Added tests for Markdown rendering, sanitization, preserved heading IDs, and rejecting documents outside the workspace.
- Connected the frontend directory tree so clicking a Markdown file opens it through Go.
- Replaced the reader placeholder with a real Markdown reading view.
- Added document title and selectable path display.
- Added right-side outline rendering and heading scroll navigation.

### 变更

- Directory rows now act as toggles instead of disabled labels.
- The reader shell now clears the selected document when a new workspace is opened.
- The Markdown body and document path explicitly allow text selection while the rest of the shell remains anti-web.

### 验证

- Directory tree collapse/scroll update:
  - `go test ./...` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
- `go test ./...` passed.
- `npm.cmd run build` passed.
- `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
- `wails build` passed and produced `build/bin/jskernmd.exe`.
- Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

### 备注

- `go get @latest` could not reach `proxy.golang.org` or GitHub from this environment, so dependencies were added from the local module cache: `goldmark v1.7.4` and `bluemonday v1.0.27`.

---

### 新增

- Initialized the Wails React + TypeScript project for `JS Kern.md`.
- Added durable project memory files:
  - `AGENTS.md`
  - `docs/PRODUCT.md`
  - `docs/ARCHITECTURE.md`
  - `docs/CONSTRAINTS.md`
  - `docs/PROJECT_STATE.md`
  - `docs/DECISIONS.md`
  - `docs/CHANGELOG.md`
- Added `product.manifest.json` as the current product identity source.
- Added Go-managed locale files for `zh-CN` and `en`.
- Added Go bootstrap API returning product info, shell locale, and business locale.
- Added Go workspace directory-tree scanning API.
- Replaced the generated Wails demo UI with a desktop Markdown reader shell:
  - frameless custom titlebar
  - workspace toolbar
  - left directory tree panel
  - center reader surface
  - right outline panel
- Added frontend desktop guards for context menu, refresh, find, zoom, F12, dragstart, and Ctrl-wheel behavior.
- Initialized the local Git repository on `main`.
- Created and pushed the public GitHub repository: `https://github.com/xiaotianwm/jskern.md`.

### 变更

- Set Wails output filename to `jskernmd`.
- Set app display title to `JS Kern.md`.
- Removed the default Wails demo interaction from the main UI.
- Removed the generated web font usage from active styles and switched to system fonts.
- Upgraded the frontend development toolchain to current Vite, TypeScript, and React plugin packages after npm audit found vulnerabilities in the Wails template defaults.
- Updated TypeScript config to modern `moduleResolution: "Bundler"` so the upgraded toolchain builds cleanly.

### 已记录约束

- The MVP must be directory-tree based.
- Wails is the only allowed desktop runtime.
- Electron is forbidden.
- Go owns filesystem access, Markdown parsing, durable state, and i18n.
- React only renders Go-provided data and short-lived interaction state.

### 验证

- `go test ./...` passed.
- `npm.cmd run build` passed.
- `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
- `wails build` passed and produced `build/bin/jskernmd.exe`.
- Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
- GitHub push to `origin/main` completed.
