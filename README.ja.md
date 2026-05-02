# vibe-workspace

人間とAIエージェント（Claude Code）が同じドキュメントを読み、同じルールで働く、**生きた開発標準ワークスペース**です。

🇯🇵 日本語 | [🇰🇷 한국어](README.md)

---

## このリポジトリについて

プロジェクト集ではなく、**ワークスペースそのものが成果物**です。

- プロジェクトが増えるたびに、標準とガイドラインが厚みを増していく
- 人間もAIエージェントも**同じドキュメント**を参照し、同じ成果物を生み出す
- 新しいプロジェクトを始めても、コンテキストを再説明する必要がない

このリポジトリが示すのは「何を作ったか」ではなく、**それらのプロジェクトが育つ土壌をどう設計したか**です。

## ワークスペースの核: 生きた開発標準

`knowledge_base/` には言語/ドメインを横断する開発標準ドキュメントが蓄積されています。どのプロジェクトを始めても、AIエージェントはまずこれらのドキュメントを読んでからコードを生成します。

| ドキュメント | 適用範囲 |
|--------------|----------|
| [`common/api-response.md`](knowledge_base/common/api-response.md) | すべてのバックエンドAPIのレスポンスエンベロープ・エラーコード体系 — 言語/フレームワーク非依存 |
| [`common/git-workflow.md`](knowledge_base/common/git-workflow.md) | 全プロジェクト共通のブランチ戦略・コミットメッセージ規約 |
| [`go/conventions.md`](knowledge_base/go/conventions.md) | Goバックエンド作成時の命名・エラー処理・ロギング規約 |
| [`go/design-patterns.md`](knowledge_base/go/design-patterns.md) | Goプロジェクトのアーキテクチャパターン（Clean Architecture、DIなど） |
| [`flutter/conventions.md`](knowledge_base/flutter/conventions.md) | Flutterアプリ作成時のウィジェット・状態管理・ルーティング規約 |
| [`flutter/design-patterns.md`](knowledge_base/flutter/design-patterns.md) | Flutter Feature-First構成とRiverpodパターン |
| [`flutter/gotchas.md`](knowledge_base/flutter/gotchas.md) | Flutter作業中に繰り返し遭遇した落とし穴と解決方法 |

### どう適用されるか

新しいバックエンドを始めるとき:

1. プロジェクトルートの `CLAUDE.md` に **「このプロジェクトは `knowledge_base/go/` と `knowledge_base/common/` に従う」** と一行
2. その瞬間からAIはすべてのエンドポイントを同じレスポンスエンベロープ・同じエラーコード体系で記述する
3. 別のバックエンドプロジェクトに移っても同じドキュメントを参照 → **同じ送受信仕様が自動的に強制される**

### 「生きている」ということ

標準は一度作って終わりではありません。実プロジェクトで見つけたパターンや失敗が再び標準に反映されるフィードバックループを持っています。

- Flutter `gotchas.md` は作業中に遭遇した落とし穴の蓄積
- Go `design-patterns.md` は複数のバックエンドを経て洗練された構造判断
- 新しい言語/ドメインが加われば `knowledge_base/` 配下に新しいフォルダが増える

## コンテキスト階層: CLAUDE.md

`CLAUDE.md` は人間とAIが共に読む**コンテキストの入口**です。ワークスペース → プロジェクトへと下るにつれて具体化されていく階層構造を持ちます。

```
ワークスペース CLAUDE.md       ← 作業者/環境/共通禁止事項
    └── プロジェクト CLAUDE.md   ← そのプロジェクトのスタック・アーキテクチャ・テスト規約
            └── knowledge_base/ 参照  ← 言語・ドメイン別の標準に委譲
```

上位は変わらず、下位はプロジェクトごとに異なり、標準は一箇所に集まっている。人間が新メンバーをオンボーディングするときに見せるドキュメントと、AIエージェントが作業開始時に読むドキュメントが完全に一致します。

設計思想の詳細 → [environment/workspace-structure.ja.md](environment/workspace-structure.ja.md)

## 環境構成

| ディレクトリ | 役割 |
|--------------|------|
| `environment/` | ワークスペースのAIコンテキスト + 構造設計の思想 |
| `knowledge_base/` | 言語別・ドメイン別の開発標準（上記参照） |
| `cases/` | 標準システムが実プロジェクトでどう機能したかの事例 |

## 事例: システムは実際に機能するか

### [OpenClaw Workbench](cases/claw-dash/)

AIエージェントのオーケストレーションダッシュボード。Go BFF + Svelte 5 SPAの構成で、
OpenClaw GatewayとWebSocketで接続し、リアルタイムにエージェントセッションを可視化する。

このケースは、上で説明したワークスペース構造が実プロジェクトの進行中にどう機能したかを追跡します — コンテキストをどう構成し、AIとどこで試行錯誤し、どんなルールを作って解決したか。プロジェクト自慢ではなく、**標準システムの検証事例**です。

## 推奨の読み順

1. **[ここから]** ワークスペース概要（このファイル）
2. **[核心]** [knowledge_base/](knowledge_base/) — 実際の標準ドキュメント群
3. **[補足]** [environment/CLAUDE.ja.md](environment/CLAUDE.ja.md) — コンテキスト階層の原本
4. **[事例]** [cases/claw-dash/](cases/claw-dash/) — システム検証のケーススタディ
