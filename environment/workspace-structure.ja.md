# ワークスペース構造の設計

AIエージェント（Claude Code）と協働するために設計したローカルワークスペースの構造。

## ディレクトリ構造

```
~/workspace/
├── knowledge_base/    # プロジェクト横断の開発標準
├── obsidian/          # 作業ログ、意思決定記録、レビュー（Obsidian Vault）
├── projects/
│   └── self/          # 個人トイプロジェクト集
└── CLAUDE.md          # ワークスペースルートのAIコンテキスト
```

## 設計原則

### knowledge_base/ — 「AIに渡す共通ルール」

言語別・ドメイン別の開発標準を一箇所にまとめて管理する。
すべてのプロジェクトのCLAUDE.mdから`@`参照で読み込む。

- `common/`: APIレスポンス標準、Gitワークフローなど言語に依存しないルール
- `go/`: Goのコーディング規約、デザインパターン
- `flutter/`: Flutter/Dartの規約

この構造のおかげで、新しいプロジェクトを始める際にCLAUDE.mdで標準ドキュメントを参照するだけで、
AIエージェントが一貫したコーディングスタイルで作業できる。

### obsidian/ — 「作業記録ハブ」

プロジェクト進行中の意思決定、試行錯誤の記録、PhaseごとのレビューをObsidian Vaultで管理する。
コードリポジトリに入れるのが難しい「なぜこうしたか」を残す場所だ。

### projects/self/ — 「実験場」

外部の要件なしに自ら企画・開発するトイプロジェクト集。
各プロジェクトにCLAUDE.mdを置いて、プロジェクト固有のAIコンテキストを管理する。

## CLAUDE.mdの階層構造

```
workspace/CLAUDE.md                          # ワークスペース共通コンテキスト
└── projects/self/<project>/CLAUDE.md        # プロジェクト固有コンテキスト
    └── @knowledge_base/go/conventions.md   # 標準ドキュメントの参照
```

AIエージェントは現在の作業ディレクトリのCLAUDE.mdを読み込み、
その中で参照しているknowledge_baseのドキュメントも合わせてロードする。
