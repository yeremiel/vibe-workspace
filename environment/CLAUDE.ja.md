# CLAUDE.md — ワークスペースガイド

> AIアシスタント（Claude Code）がこのワークスペースで作業する際に参照するメタドキュメント。

---

## 1. 作業者情報 (WHO)

| 項目 | 値 |
|------|-----|
| 名前 | Miel (yeremiel) |
| 主な使用言語 | Java（16年）、Go、TypeScript、Flutter |
| タイムゾーン | Asia/Seoul (KST, UTC+9) |
| 会話言語 | 韓国語 |

---

## 2. ディレクトリ構造 (WHERE)

```
./
├── knowledge_base/                  # ナレッジベース
├── obsidian/                        # Obsidian Vault（ナレッジベース・作業記録）
├── projects/
│   └── self/
│       ├── claw-dash/               # OpenClaw Workbench（repoID: claw-dash）
│       │   └── CLAUDE.md
│       ├── claw-usage-chart/        # OpenClawトークン使用量可視化ツール（→ claw-dash統合予定）
│       │   └── CLAUDE.md
│       ├── kanban-test/             # Kanbanプロジェクト（実験・初期化のみ）
│       ├── Learning-Go-study/       # Go言語学習プロジェクト（ch1〜ch2）
│       ├── scanner/                 # オフライン文書スキャナーアプリ（Flutter + Android CameraX/OpenCV）
│       └── myLib/                   # 個人蔵書管理サービス（メインプロジェクト）
│           ├── CLAUDE.md            # ワークスペースオーケストレーション
│           ├── docs/                # ワークスペースレベルのドキュメント
│           ├── myLib-backend/       # GoバックエンドAPI（独立gitリポジトリ）
│           │   └── CLAUDE.md        # Go/Gin/MongoDB/Redisルール
│           └── myLib-frontend/      # Flutterフロントエンドアプリ（独立gitリポジトリ）
│               └── CLAUDE.md        # Flutter/Riverpod/Dioルール
├── shared/                          # プロジェクト間共有リソース（予定）
└── CLAUDE.md                        # ← このファイル
```

---

## 3. 開発環境 (WHERE)

### システム

| 項目 | バージョン |
|------|-----------|
| OS | macOS 26.3.1 (Darwin 25.3.0 arm64) |
| アーキテクチャ | Apple Silicon (arm64) |

### ランタイム / SDK

| ツール | バージョン | 備考 |
|--------|-----------|------|
| Go | 1.26.1 | myLib-backend: `go 1.25.5` (go.mod) |
| Flutter | 3.38.7 (stable) | myLib-frontend, scanner |
| Dart | 3.10.7 | Flutter同梱 |
| Node.js | 25.8.2 | |
| npm | 11.11.1 | |
| Bun | 1.3.11 | |
| pnpm | 10.33.0 | |
| Docker | — | 未インストール |
| mongosh | 2.8.2 | MongoDBシェル |
| Java | OpenJDK 25.0.2 (Temurin) | scanner Androidビルド用 |

### プロジェクト別技術スタック

| プロジェクト | 言語 | フレームワーク / 主要ライブラリ | DB / インフラ |
|-------------|------|-------------------------------|--------------|
| myLib-backend | Go 1.25 | Gin, JWT v5, Viper, Zap, Swaggo | MongoDB 6.0+, Redis 7.0+ |
| myLib-frontend | Dart 3.10 / Flutter 3.38 | Riverpod 3.x, Dio 5.x, go_router, Freezed | flutter_secure_storage |
| claw-dash (OpenClaw Workbench) | Go + Svelte 5 | Gin (BFF), Vite | OpenClaw Gateway (WebSocket) |
| claw-usage-chart | Go | embed.FS, modernc.org/sqlite | SQLite (WAL), Chart.js |
| scanner | Dart 3.10 / Flutter 3.38 | CameraX, OpenCV (Android NDK) | ローカル保存 (MediaStore) |
| Learning-Go-study | Go 1.22 | 標準ライブラリ | — |

---

## 4. 作業ルール (HOW)

### 言語とコメント
- すべての会話、コメント、コミットメッセージは**韓国語**で記述
- コード識別子（変数名、関数名など）は英語を維持

### Gitワークフロー
- コミットメッセージの規約、ブランチ戦略 → `knowledge_base/common/git-workflow.md` 参照
- コミット前に必ずメッセージをユーザーに提案 → 承認後にコミット

### Knowledge Base（共通開発標準）

`knowledge_base/` ディレクトリにプロジェクト横断の開発標準が定義されている。

**全プロジェクト共通適用:**

| ドキュメント | パス | 内容 |
|-------------|------|------|
| APIレスポンス標準 | `knowledge_base/common/api-response.md` | レスポンスenvelope構造、エラーコードルール、HTTPステータスコードマッピング |
| Gitワークフロー | `knowledge_base/common/git-workflow.md` | ブランチ戦略、コミットメッセージ規約、コミット前の承認手順 |

**言語別ドキュメント**（`go/`、`flutter/`など）は各プロジェクトのCLAUDE.mdから該当するものだけを参照する。

### プロジェクト別詳細ルール
- 各プロジェクトの`CLAUDE.md`にコーディングルール、アーキテクチャパターン、テストガイドが定義されている
- このファイルでは重複させず、**配下のCLAUDE.mdに委譲**
- プロジェクト別のKnowledge Base例外事項は該当プロジェクトのCLAUDE.mdに明記

---

## 5. 禁止事項 (WHAT NOT)

| # | ルール | 説明 |
|---|--------|------|
| 1 | `.env` / APIキーのハードコード禁止 | 環境変数または設定ファイルで管理 |
| 2 | 生成物の直接編集禁止 | `build/`、`node_modules/`、`.g.dart`、`generated/`などのビルド成果物を手動編集しない |
| 3 | 独断での作業実行禁止 | 指示していないリファクタリング、機能追加、ファイル削除などを勝手に実行しない |
| 4 | キャッシュのせいにしない | 問題発生時はキャッシュではなく自分のコードをまず疑う |
| 5 | 不可能を即座に宣言しない | 不可能と判断する前に**最低3つの代替案**を提示する |
| 6 | 推測によるURL生成禁止 | 確認されていないURLを生成・提案しない |
| 7 | 機密情報のログ出力禁止 | トークン、パスワードなどの機密データをログに残さない |

---

## 6. プロジェクト別クイックリファレンス

### myLib-backend

```bash
make run               # ビルド後実行（Swagger含む）
make test              # 全テスト
make lint              # golangci-lint
make swagger           # Swaggerドキュメント生成
```

### myLib-frontend

```bash
make deps              # flutter pub get
make build-runner      # コード生成（freezed, json_serializable）
make test              # ユニットテスト
make api-update        # バックエンドSwagger → Dartクライアント再生成
```

### claw-dash

```bash
make dev               # バックエンド+フロント同時起動
make ci                # ビルド/テスト/セキュリティ/MVP検証の統合実行
make security          # 静的解析
```

### claw-usage-chart

```bash
go build -o claw-usage-chart .   # ビルド
./claw-usage-chart               # 実行（localhost:8585）
./claw-usage-chart -d            # デーモンモード
```

### scanner

```bash
flutter test                     # ユニットテスト
flutter analyze                  # 静的解析（JAVA_HOME設定が必要）
flutter build apk --debug        # デバッグAPKビルド（JAVA_HOME設定が必要）
```

### 共通アーキテクチャ

```
Backend:  Handler → Service → Repository (Clean Architecture)
Frontend: Screen/Widget → Provider → Repository → DataSource (Feature-First)
API連携:  Backend swagger.json → Frontend make api-update → 自動生成クライアント
```

---

*最終更新: 2026-04-05*
