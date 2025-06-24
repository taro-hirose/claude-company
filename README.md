# Claude Company

**AI 駆動の tmux セッション管理とタスク委譲ツール**

Claude Company は、tmux セッション内で複数の Claude AI インスタンスを管理し、管理者と作業者の役割分離により効率的なタスク実行を実現するツールです。管理者 AI はタスクを分析・計画し、作業者 AI が実際の実装を行います。

## インストール

```bash
git clone https://github.com/yourusername/claude-company.git
cd claude-company
make build
```

## 基本的な使用方法

```bash
# tmuxセッションのセットアップ
./claude-company

# AIチームへのタスク割り当て
./claude-company --task "タスクの説明"
```

### 使用例

```bash
# Webアプリケーション開発
./claude-company --task "ユーザー認証機能付きREST APIを作成"

# バグ修正
./claude-company --task "すべてのビルドエラーを修正"

# リファクタリング
./claude-company --task "既存コードベースをリファクタリングしてテストを追加"
```

### 動作原理

- **管理者 AI**: タスク分析、計画立案、品質レビュー（コードは書かない）
- **作業者 AI**: 実際のコード記述、ファイル作成、実装作業

管理者は必要に応じて複数の作業者ペインを作成し、サブタスクを並列実行させます。

### システム要件

- Go 1.21+
- tmux
- Claude AI アクセス
- Unix系OS（Linux、macOS、WSL）

### 詳細情報

詳細な機能説明、オーケストレーターモード、アーキテクチャ図、設定オプション、トラブルシューティング、開発者向け情報については、[DETAILED_GUIDE.md](DETAILED_GUIDE.md) を参照してください。

## ライセンス

MIT ライセンス - 詳細は[LICENSE](LICENSE)ファイルを参照してください。