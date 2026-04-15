# vibe-workspace

AIエージェント（Claude Code）と協働するための開発環境ポートフォリオです。

🇯🇵 日本語 | [🇰🇷 한국어](README.md)

---

## このリポジトリについて

単なるプロジェクト集ではありません。AIエージェントと協働するために**環境そのものをどう設計・運用するか**を示しています。

- AIにどんなコンテキストを与え、どんなルールで作業したか
- 機能単位でどう計画し、AIと一緒に実行したか
- 実際のプロジェクトでどんな問題に当たり、どう乗り越えたか

## 推奨の読み順

1. **[ここから]** 環境設計の概要（このファイル）
2. **[核心]** [cases/claw-dash/](cases/claw-dash/) — 実プロジェクトのケーススタディ
3. **[参考]** [environment/CLAUDE.md](environment/CLAUDE.md) — AIコンテキスト設計の原本
4. **[参考]** [knowledge_base/](knowledge_base/) — 開発標準ドキュメント

## 環境構成

| ディレクトリ | 役割 |
|-------------|------|
| `environment/` | ワークスペースのAIコンテキスト + 構造設計の思想 |
| `knowledge_base/` | 言語別・ドメイン別の開発標準 |
| `cases/` | プロジェクト別ケーススタディ |

構造設計の詳細な理由 → [environment/workspace-structure.md](environment/workspace-structure.md)

## ケーススタディ

### [OpenClaw Workbench](cases/claw-dash/)

AIエージェントのオーケストレーションダッシュボード。Go BFF + Svelte 5 SPAの構成で、
OpenClaw GatewayとWebSocketで接続し、リアルタイムにエージェントセッションを可視化する。

このプロジェクトがAIと最も試行錯誤し、最も多くを学んだ場所だった。

## 並行開発中のプロジェクト

- **myLib** — 個人蔵書管理サービス（Go + Flutter、非公開）
- **scanner** — Flutter + OpenCV オフライン文書スキャナー（非公開）
