# Claude Company 🤖

**AI-powered project management system with intelligent task delegation**

Claude Company transforms your development workflow by creating an AI-powered team where one Claude AI acts as a project manager, orchestrating multiple worker Claude AIs to collaboratively complete complex tasks.

## ✨ Key Features

### 🎯 **AI Project Manager System**
- **Smart Task Delegation**: Parent pane analyzes tasks and breaks them into manageable subtasks
- **Intelligent Worker Management**: Automatically creates and manages child panes for parallel work
- **Quality Control**: Built-in review system and integration testing
- **Real-time Progress Monitoring**: Track completion status across all workers

### ⚡ **STORM Session Manager**
- **Lightning-fast tmux session management**
- **Cross-shell compatibility** (bash, zsh, fish)
- **Clean command interface** for session lifecycle management

### 🔄 **Automated Workflow**
- **Role Separation**: Manager for oversight, workers for implementation
- **Automatic Pane Creation**: Dynamic scaling based on task complexity
- **Quality Assurance**: Mandatory code review and build testing
- **Seamless Integration**: Built-in tmux and Claude AI integration

## 🚀 Quick Start

### 1. Installation
```bash
# Clone and build
git clone https://github.com/yourusername/claude-company.git
cd claude-company
go build -o bin/ccs

# Or use the install script
./install.sh
```

### 2. Setup Claude Company Session
```bash
# Create tmux session with AI-powered workspace
./bin/ccs
# This creates a structured environment with manager and worker panes
```

### 3. Assign Tasks to AI Team
```bash
# AI Manager Mode (full functionality with database)
./bin/ccs --task "Implement user authentication system with JWT tokens"
```

### 4. Watch the Magic Happen
1. **Manager pane** analyzes the task and creates a project plan
2. **Worker panes** are automatically created and assigned specific subtasks
3. **Implementation** happens in parallel across multiple Claude AIs
4. **Quality control** - Manager reviews all work and coordinates testing
5. **Integration** - Final build and validation

## 🏗️ Architecture Overview

### **Role-Based AI Team Structure**

```
┌─────────────────┬─────────────────┐
│ 🎯 Manager Pane │ 🔧 Worker Pane  │
│ (Claude AI #1)  │ (Claude AI #2)  │
│                 │                 │
│ • Task Analysis │ • Code Writing  │
│ • Planning      │ • File Creation │
│ • Review        │ • Implementation│
│ • Quality Check │ • Bug Fixes     │
├─────────────────┼─────────────────┤
│ 🔍 Review Pane  │ 🧪 Test Pane    │
│ (Claude AI #3)  │ (Claude AI #4)  │
│                 │                 │
│ • Code Review   │ • Unit Testing  │
│ • Standards     │ • Integration   │
│ • Optimization  │ • Validation    │
│ • Documentation │ • Build Checks  │
└─────────────────┴─────────────────┘
```

### **Workflow Process**

1. **📋 Task Input**: You provide a high-level task description
2. **🧠 Analysis**: Manager AI breaks down the task into subtasks  
3. **🏭 Scaling**: Manager creates additional worker panes as needed
4. **⚡ Parallel Execution**: Multiple Claude AIs work simultaneously
5. **🔍 Quality Control**: Manager reviews all implementation work
6. **🧪 Testing**: Automated build verification and testing
7. **✅ Integration**: Final validation and completion

## 📚 Usage Examples

### **Software Development Tasks**

```bash
# Full-stack application development
./bin/ccs --task "Create a REST API with authentication, user management, and a React frontend"

# Code refactoring and optimization  
./bin/ccs --task "Refactor the existing codebase for better maintainability and add comprehensive tests"

# Bug fixing and enhancement
./bin/ccs --task "Fix all build errors and add logging functionality throughout the application"
```

### **Project Management Tasks**

```bash
# Architecture design
./bin/ccs --task "Design a microservices architecture for the e-commerce platform and implement the user service"

# Documentation creation
./bin/ccs --task "Create comprehensive API documentation and add inline code comments"

# Performance optimization
./bin/ccs --task "Profile the application, identify bottlenecks, and implement performance improvements"
```

## 🛠️ コマンドリファレンス (Command Reference)

### **メインコマンド (Main Commands)**
```bash
# tmuxセッションのセットアップ（デフォルト動作）
./bin/ccs
./bin/ccs --setup

# AIチームにタスクを割り当て（データベースが必要）
./bin/ccs --task "TASK_DESCRIPTION"

# APIサーバーモード
./bin/ccs --api
```

### **STORMセッション管理 (STORM Session Management)**
```bash
# 全セッションをリスト表示
./bin/storm list     # または 'ls'

# 新しいセッションを作成
./bin/storm new <session-name>

# セッションにアタッチ
./bin/storm attach <session-name>  # または 'a'

# セッションを終了
./bin/storm kill <session-name>    # または 'k'

# セッション間の切り替え
./bin/storm switch <session-name>  # または 's'

# セッション名を変更
./bin/storm rename <old-name> <new-name>  # または 'r'
```

## 🎯 動作原理 (How It Works)

### **マネージャーAIの責任 (Manager AI Responsibilities)**
- ❌ **コードを直接書くことはしない**  
- ✅ **タスクを分析・分解する**
- ✅ **ワーカーペインを作成・管理する**
- ✅ **特定のサブタスクをワーカーに割り当てる**
- ✅ **完了した作業の品質をレビューする**
- ✅ **テストと統合を調整する**
- ✅ **最終承認と完了確認を行う**

### **ワーカーAIの責任 (Worker AI Responsibilities)**  
- ✅ **割り当てられたサブタスクを実装する**
- ✅ **実際のコードを書きファイルを作成する**
- ✅ **成果物と共に完了を報告する**
- ✅ **フィードバックと修正要請に対応する**
- ✅ **レビューで特定された問題を修正する**

### **コミュニケーションプロトコル (Communication Protocol)**
```bash
# ワーカー → マネージャー 報告フォーマット
"実装完了：internal/auth/jwt.go - JWT token generation and validation implemented"

# マネージャー → ワーカー タスク割り当てフォーマット
"サブタスク: Create user authentication middleware in internal/auth/middleware.go. Include JWT validation and error handling. Report completion when done."

# マネージャー → ワーカー レビュー要請
"レビュー要請: Please review internal/auth/jwt.go for code quality, security best practices, and integration compatibility."
```

## 🔧 インストール・セットアップ (Installation & Setup)

### **システム要件 (System Requirements)**
- **Go 1.21+** - ソースからのビルド用
- **tmux** - ペイン管理に必要
- **Claude AIアクセス** - Claude CLIツール経由
- **Docker & Docker Compose** - データベースサービス用
- **Unix系OS** - Linux、macOS、またはWSL

### **段階的インストール (Step-by-Step Installation)**

1. **依存関係のインストール (Install Dependencies)**
```bash
# macOS
brew install tmux go docker

# Ubuntu/Debian  
sudo apt install tmux golang-go docker.io docker-compose

# Claude CLIのインストール（公式ドキュメントに従う）
```

2. **Claude Companyのビルド (Build Claude Company)**
```bash
git clone https://github.com/yourusername/claude-company.git
cd claude-company
go mod tidy
go build -o bin/ccs
```

3. **PATH設定（オプション）(Setup PATH (optional))**
```bash
cp bin/ccs ~/bin/ccs
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

## 🗄️ データベース設定・構成 (Database Setup & Configuration)

### **前提条件 (Prerequisites)**

開始前に、以下がインストールされていることを確認してください：

```bash
# Dockerバージョンの確認（最低要件：20.10+）
docker --version

# Docker Composeバージョンの確認（最低要件：2.0+）
docker-compose --version

# Dockerデーモンの動作確認
docker info
```

**必要なバージョン (Required versions):**
- **Docker**: 20.10.0以上
- **Docker Compose**: 2.0.0以上
- **利用可能ポート (Available ports)**: 5432 (PostgreSQL), 8080 (pgAdmin)

### **データベース構成 (Database Configuration)**

このプロジェクトはPostgreSQLをデータベースバックエンドとして使用します：

**🟢 PostgreSQL**
- 高度なJSONサポートと複雑なクエリ
- 分析ワークロードでのより良いパフォーマンス
- pgAdmin Webインターフェースを含む
- Claude Company機能のデフォルト選択

### **1. 初期セットアッププロセス (Initial Setup Process)**

#### **Step 1: 環境設定 (Environment Configuration)**

データベース設定用の環境ファイルを作成：

```bash
# プロジェクトルートに.envファイルを作成
cat > .env << 'EOF'
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=claude_user
DB_PASSWORD=claude_password
DB_NAME=claude_company
DB_SSLMODE=disable

# API Server Configuration
PORT=8081
GIN_MODE=release

# Optional: Choose database type (postgres/mysql)
DB_TYPE=postgres
EOF
```

#### **Step 2: データベースサービスの開始 (Start Database Services)**

```bash
# 全データベースサービスをバックグラウンドで開始
docker-compose up -d

# PostgreSQLサービスを開始
docker-compose up -d postgres pgadmin
```

#### **Step 3: サービス状態の確認 (Verify Service Status)**

```bash
# 全サービスの状態を確認
docker-compose ps

# 期待される出力:
# NAME                     IMAGE               STATUS
# claude-company-db        postgres:15-alpine  Up (healthy)
# claude-company-pgadmin   dpage/pgadmin4      Up
```

**開始されたサービス (Services Started):**
- **PostgreSQLデータベース**: `localhost:5432`
- **pgAdmin Webインターフェース**: `http://localhost:8080`

### **2. 接続確認・テスト (Connection Verification & Testing)**

#### **Step 4: データベース接続の確認 (Verify Database Connection)**

**PostgreSQL接続テスト:**
```bash
# サービスが完全に準備されるまで待機（30-60秒かかる場合があります）
docker-compose logs postgres | grep "ready to accept connections"

# docker execを使用した接続テスト
docker exec claude-company-db psql -U claude_user -d claude_company -c "SELECT version();"

# ホストマシンからの接続テスト（psqlクライアントが必要）
PGPASSWORD=claude_password psql -h localhost -p 5432 -U claude_user -d claude_company -c "\\dt"
```


#### **Step 5: Webインターフェースへのアクセス (Access Web Interfaces)**

**pgAdmin (PostgreSQL管理):**
1. ブラウザで`http://localhost:8080`を開く
2. ログイン認証情報:
   - **Email**: `admin@claude-company.local`
   - **Password**: `admin123`
3. サーバー接続を追加:
   - **Host**: `postgres` (Dockerサービス名)
   - **Port**: `5432`
   - **Username**: `claude_user`
   - **Password**: `claude_password`


### **3. 環境変数リファレンス (Environment Variables Reference)**

データベース接続設定の構成：

```bash
# PostgreSQL用の必須環境変数
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=claude_user
export DB_PASSWORD=claude_password
export DB_NAME=claude_company
export DB_SSLMODE=disable


# APIサーバー設定
export PORT=8081
export GIN_MODE=release
```

**環境ファイル（.env）テンプレート:**
```bash
# PostgreSQL Configuration (default)
DB_HOST=localhost
DB_PORT=5432
DB_USER=claude_user
DB_PASSWORD=claude_password
DB_NAME=claude_company
DB_SSLMODE=disable


# API Configuration
PORT=8081
GIN_MODE=release

# Development settings
DB_TYPE=postgres
```

### **4. データベーススキーマ・初期化 (Database Schema & Initialization)**

データベーススキーマはコンテナ起動時に自動的に初期化されます。初期化には以下が含まれます：

- **階層構造とULID主キーを持つタスクテーブル**
- **リアルタイム監視用の進捗追跡テーブル**
- **パフォーマンス最適化用のインデックス**
- **タスク階層管理と計算用の関数**
- **テスト・デモ用のサンプルデータ**

**自動初期化プロセス:**
```bash
# スキーマファイルは初回起動時に自動実行されます
# 場所: ./db/init/01_schema.sql and ./db/init/02_sample_data.sql

# 初期化ログの確認
docker-compose logs postgres | grep -i "database system is ready"
```

**手動スキーマ検査:**
```bash
# PostgreSQL - テーブル構造の表示
docker exec claude-company-db psql -U claude_user -d claude_company -c "\\dt"

# PostgreSQL - 特定テーブルの詳細確認
docker exec claude-company-db psql -U claude_user -d claude_company -c "\\d tasks"

```

### **5. 開発環境設定 (Development Environment Configuration)**

#### **推奨開発設定 (Recommended Development Settings)**

開発作業用の最適化された設定：

```bash
# 開発用.env設定
cat > .env << 'EOF'
# Database (Development optimized)
DB_HOST=localhost
DB_PORT=5432
DB_USER=claude_user
DB_PASSWORD=claude_password
DB_NAME=claude_company_dev
DB_SSLMODE=disable
DB_MAX_CONNECTIONS=20
DB_IDLE_TIMEOUT=300

# API Server (Development mode)
PORT=8081
GIN_MODE=debug
LOG_LEVEL=debug
ENABLE_CORS=true

# Development features
AUTO_MIGRATE=true
SEED_DATABASE=true
EOF
```

#### **開発用パフォーマンスチューニング (Performance Tuning for Development)**

```bash
# PostgreSQL開発最適化
docker exec claude-company-db psql -U claude_user -d claude_company -c "
  -- 複雑なクエリ用のワークメモリ増加
  ALTER SYSTEM SET work_mem = '64MB';
  
  -- デバッグ用クエリログ有効化
  ALTER SYSTEM SET log_statement = 'all';
  ALTER SYSTEM SET log_min_duration_statement = 1000;
  
  -- 設定のリロード
  SELECT pg_reload_conf();
"
```

### **6. よくある問題のトラブルシューティング (Troubleshooting Common Issues)**

#### **🔴 データベース接続問題 (Database Connection Issues)**

**問題**: "Connection refused"エラー
```bash
# 解決策1: コンテナが動作しているか確認
docker-compose ps

# 解決策2: ポートの利用可能性を確認
netstat -tulpn | grep :5432

# 解決策3: クリーンな状態でサービスを再起動
docker-compose down -v && docker-compose up -d
```

**問題**: "Password authentication failed"
```bash
# 解決策: データベースコンテナのリセット
docker-compose down
docker volume rm claude-company_postgres_data
docker-compose up -d postgres
```

#### **🔴 Docker問題 (Docker Issues)**

**問題**: "Port already in use"
```bash
# ポートを使用しているプロセスを探す
lsof -i :5432

# プロセスを終了するか、docker-compose.ymlでポートを変更
# 例: "5432:5432"を"5433:5432"に変更
```

**問題**: "Volume mount failed"
```bash
# Dockerがプロジェクトディレクトリにアクセス権限を持っていることを確認
# macOS: Docker Desktop > Settings > Resources > File Sharing
# Linux: SELinux/AppArmorの権限を確認
```

#### **🔴 パフォーマンス問題 (Performance Issues)**

**問題**: データベースクエリが遅い
```bash
# データベース統計の確認
docker exec claude-company-db psql -U claude_user -d claude_company -c "
  SELECT schemaname, tablename, attname, n_distinct, correlation 
  FROM pg_stats WHERE tablename = 'tasks';
"

# クエリパフォーマンスの分析
docker exec claude-company-db psql -U claude_user -d claude_company -c "
  EXPLAIN ANALYZE SELECT * FROM tasks WHERE status = 'pending';
"
```

#### **🔴 Webインターフェース問題 (Web Interface Issues)**

**問題**: pgAdminにアクセスできない
```bash
# コンテナログの確認
docker-compose logs pgadmin

# 正しいURLとポートの確認
echo "pgAdmin: http://localhost:8080"

# ブラウザキャッシュをクリアして再試行
```

### **7. クイックヘルスチェックコマンド (Quick Health Check Commands)**

```bash
# 完全なシステムヘルスチェック
#!/bin/bash
echo "=== Claude Company Database Health Check ==="

# Dockerの確認
echo "Docker version: $(docker --version)"

# コンテナの確認
echo -e "\nContainer status:"
docker-compose ps

# データベース接続の確認
echo -e "\nDatabase connectivity:"
docker exec claude-company-db psql -U claude_user -d claude_company -c "SELECT 'PostgreSQL Connected' as status;"

# Webインターフェースの確認
echo -e "\nWeb interface availability:"
curl -s -o /dev/null -w "pgAdmin: %{http_code}\n" http://localhost:8080

echo -e "\nSetup complete! ✅"
```

## 🌐 API使用ガイド

### **APIサーバーの起動**

```bash
# データベースを有効にして起動
./bin/ccs --api --port 8081

# または環境変数を使用
PORT=8081 ./bin/ccs --api
```

**API ベースURL:** `http://localhost:8081/api/v1`

### **コアタスク管理エンドポイント**

#### **タスク作成**
```bash
# メインタスク作成
curl -X POST http://localhost:8081/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Implement user authentication system",
    "mode": "manager",
    "pane_id": "pane_1",
    "priority": 3
  }'

# サブタスク作成
curl -X POST http://localhost:8081/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "01HXAMPLE123456789",
    "description": "Create JWT middleware",
    "mode": "worker",
    "pane_id": "pane_2",
    "priority": 2
  }'
```

#### **タスク取得**
```bash
# ペーン別でタスク取得
curl "http://localhost:8081/api/v1/tasks?pane_id=pane_1"

# ステータス別でタスク取得
curl "http://localhost:8081/api/v1/tasks?status=in_progress"

# 子タスク取得
curl "http://localhost:8081/api/v1/tasks?parent_id=01HXAMPLE123456789"

# 特定タスク取得
curl "http://localhost:8081/api/v1/tasks/01HXAMPLE123456789"
```

#### **タスク更新**
```bash
# タスク詳細更新
curl -X PUT http://localhost:8081/api/v1/tasks/01HXAMPLE123456789 \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated task description",
    "status": "in_progress",
    "result": "Middleware implemented successfully"
  }'

# ステータスのみ更新
curl -X PATCH http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/status/completed

# 関連タスクへの伝播を含むステータス更新
curl -X PATCH http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/status-propagate/completed
```

#### **タスク階層**
```bash
# 完全なタスク階層取得
curl "http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/hierarchy"
```

### **進捗監視エンドポイント**

#### **進捗サマリー取得**
```bash
# 特定ペーンの進捗取得
curl "http://localhost:8081/api/v1/progress?pane_id=pane_1"

# レスポンス例:
{
  "total_tasks": 10,
  "completed_tasks": 7,
  "pending_tasks": 2,
  "in_progress_tasks": 1,
  "progress_percent": 70.0
}
```

#### **タスク統計取得**
```bash
# 詳細統計取得
curl "http://localhost:8081/api/v1/statistics?pane_id=pane_1"
```

## 🤝 タスク共有機能

### **個別タスク共有**

```bash
# 特定ペーンとタスク共有
curl -X POST http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/share \
  -H "Content-Type: application/json" \
  -d '{
    "pane_id": "pane_3",
    "permission": "write"
  }'

# 兄弟タスク全てと共有（同じ親）
curl -X POST http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/share-siblings

# タスクファミリー全体と共有（親+子）
curl -X POST http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/share-family
```

### **タスク共有管理**

```bash
# タスクの全ての共有取得
curl "http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/shares"

# ペーンと共有されている全タスク取得
curl "http://localhost:8081/api/v1/shared-tasks?pane_id=pane_2"

# 共有削除
curl -X DELETE http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/share/pane_2
```

### **権限レベル**
- **`read`**: タスク詳細の閲覧のみ（デフォルト）
- **`write`**: タスクステータス更新と進捗追加が可能
- **`admin`**: 共有と削除を含む完全制御

## ⚡ 非同期実行ガイド

### **非同期タスク処理**

Claude Companyはバックグラウンド処理と並列作業分散のための非同期タスク実行をサポートしています：

#### **1. 非同期モード有効化**
```bash
# 非同期処理を有効にして起動
./bin/ccs --async --workers 4

# またはAPIモードと組み合わせ
./bin/ccs --api --async --workers 4
```

#### **2. 非同期タスク作成**
```bash
# 非同期フラグ付きタスク作成
curl -X POST http://localhost:8081/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Process large dataset analysis",
    "mode": "async_worker",
    "pane_id": "async_1",
    "priority": 1,
    "metadata": "{\"async\": true, \"timeout\": 300}"
  }'
```

#### **3. 非同期進捗監視**
```bash
# 非同期タスクステータス確認
curl "http://localhost:8081/api/v1/tasks/01HXAMPLE123456789"

# 全ての非同期タスク監視
curl "http://localhost:8081/api/v1/tasks?status=in_progress"
```

### **非同期実行パターン**

#### **並列タスク分散**
```bash
# 親タスク作成
PARENT_ID=$(curl -X POST http://localhost:8081/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Parallel data processing job",
    "mode": "manager",
    "pane_id": "manager_pane"
  }' | jq -r '.id')

# 複数の非同期サブタスク作成
for i in {1..4}; do
  curl -X POST http://localhost:8081/api/v1/tasks \
    -H "Content-Type: application/json" \
    -d "{
      \"parent_id\": \"$PARENT_ID\",
      \"description\": \"Process data chunk $i\",
      \"mode\": \"async_worker\",
      \"pane_id\": \"worker_$i\"
    }"
done
```

#### **自動共有によるタスク協調**
```bash
# ファミリーと自動共有する協調タスク作成
curl -X POST http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/share-family

# 関連タスク全てが互いの進捗を可視化
```

### **ヘルスチェック&サービス状態**
```bash
# API健全性確認
curl "http://localhost:8081/health"

# レスポンス:
{
  "status": "ok",
  "message": "Claude Company API is running"
}
```

## 🔄 データベース管理

### **サービス停止/開始**
```bash
# 全サービス停止
docker-compose down

# ボリューム削除して停止（全データクリア）
docker-compose down -v

# サービス再起動
docker-compose restart

# リソース使用量表示
docker-compose top
```

### **バックアップ&復元**
```bash
# データベースバックアップ
docker exec claude-company-db pg_dump -U claude_user claude_company > backup.sql

# データベース復元
docker exec -i claude-company-db psql -U claude_user claude_company < backup.sql
```

## 🎭 実世界での例

ユーザー認証システムを追加したい場合：

```bash
./bin/ccs --task "Add JWT-based user authentication with registration, login, and protected routes"
```

**自動的に起こること：**

1. **マネージャー分析**（親ペーン）：
   - 「これを分割する必要がある：ユーザーモデル、JWTサービス、認証ミドルウェア、登録エンドポイント、ログインエンドポイント、テスト」

2. **ワーカー作成&割り当て**：
   - 3つの子ペーンを作成
   - バックエンド作業をワーカー#1に割り当て
   - テストをワーカー#2に割り当て
   - 統合をワーカー#3に割り当て

3. **並列実装**：
   - ワーカー#1：ユーザーモデル、JWT関数、エンドポイント作成
   - ワーカー#2：ユニットテストと統合テスト作成
   - ワーカー#3：ミドルウェアとルート保護設定

4. **品質管理**：
   - マネージャーが各コンポーネントをレビュー
   - 必要に応じて修正要求
   - 最終統合テストを協調

5. **完了**：
   - 全コードが動作しテスト済み
   - ビルドが成功
   - 機能が使用可能

## 🚨 トラブルシューティング

### **一般的な問題**

**❌ "tmux: command not found"**
```bash
# 最初にtmuxをインストール
brew install tmux      # macOS
sudo apt install tmux  # Ubuntu
```

**❌ "claude: command not found"**
```bash
# 公式ドキュメントに従ってClaude CLIをインストール
# PATHに含まれていることを確認
```

**❌ ペーンが応答しない**
```bash
# 各ペーンでClaudeが実行されているか確認
tmux list-panes -s -t claude-squad -F '#{pane_id}: #{pane_current_command}'

# 必要に応じてセッション再起動
./bin/ccs --setup
```

**❌ タスクが分散されない**
```bash
# マネージャー/ワーカー分離のために少なくとも2つのペーンが存在することを確認
# マネージャーにはタスクを委譲するワーカーペーンが必要
```

## 🤝 貢献ガイドライン

貢献を歓迎します！詳細は[貢献ガイドライン](CONTRIBUTING.md)をご覧ください。

### **開発環境セットアップ**
```bash
git clone https://github.com/yourusername/claude-company.git
cd claude-company
go mod tidy
go build -o bin/ccs
./bin/ccs --task "Help improve this project"  # メタ！😄
```

## 📄 ライセンス

MIT License - 詳細は[LICENSE](LICENSE)ファイルをご覧ください。

## 🙏 謝辞

- **Claude AI** - インテリジェントなコラボレーションを可能にしてくれて
- **tmux** - 堅牢なターミナル多重化のために
- **Goコミュニティ** - 優秀なツールとライブラリのために

---

**AI駆動のチームコラボレーションで開発ワークフローを変革しよう** 🚀

*AIと協働したい、単に使うだけではない開発者のために❤️で作成*